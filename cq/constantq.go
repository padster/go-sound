package cq

import (
	"fmt"
	"math"

	"github.com/mjibson/go-dsp/fft"
)

const DEBUG_CQ = false

type ConstantQ struct {
	kernel *CQKernel

	octaves       int
	bigBlockSize  int
	OutputLatency int
	buffers       [][]float64

	latencies  []int
	decimators []*Resampler
}

func NewConstantQ(params CQParams) *ConstantQ {
	octaves := int(math.Ceil(math.Log(params.maxFrequency/params.minFrequency) / math.Log(2))) // TODO: math.Log2
	if octaves < 1 {
		panic("Need at least one octave")
	}

	kernel := NewCQKernel(params)
	p := kernel.Properties

	// Use exact powers of two for resampling rates. They don't have
	// to be related to our actual samplerate: the resampler only
	// cares about the ratio, but it only accepts integer source and
	// target rates, and if we start from the actual samplerate we
	// risk getting non-integer rates for lower octaves
	sourceRate := unsafeShift(octaves)
	latencies := make([]int, octaves, octaves)
	decimators := make([]*Resampler, octaves, octaves)

	// top octave, no resampling
	latencies[0] = 0
	decimators[0] = nil

	for i := 1; i < octaves; i++ {
		factor := unsafeShift(i)

		r := NewResampler(sourceRate, sourceRate/factor, 50, 0.05)
		if DEBUG_CQ {
			fmt.Printf("Forward: octave %d: resample from %v to %v\n", i, sourceRate, sourceRate/factor)
		}

		// We need to adapt the latencies so as to get the first input
		// sample to be aligned, in time, at the decimator output
		// across all octaves.
		//
		// Our decimator uses a linear phase filter, but being causal
		// it is not zero phase: it has a latency that depends on the
		// decimation factor. Those latencies have been calculated
		// per-octave and are available to us in the latencies
		// array. Left to its own devices, the first input sample will
		// appear at output sample 0 in the highest octave (where no
		// decimation is needed), sample number latencies[1] in the
		// next octave down, latencies[2] in the next one, etc. We get
		// to apply some artificial per-octave latency after the
		// decimator in the processing chain, in order to compensate
		// for the differing latencies associated with different
		// decimation factors. How much should we insert?
		//
		// The outputs of the decimators are at different rates (in
		// terms of the relation between clock time and samples) and
		// we want them aligned in terms of time. So, for example, a
		// latency of 10 samples with a decimation factor of 2 is
		// equivalent to a latency of 20 with no decimation -- they
		// both result in the first output sample happening at the
		// same equivalent time in milliseconds.
		//
		// So here we record the latency added by the decimator, in
		// terms of the sample rate of the undecimated signal. Then we
		// use that to compensate in a moment, when we've discovered
		// what the longest latency across all octaves is.
		latencies[i] = r.GetLatency() * factor
		decimators[i] = r
	}

	bigBlockSize := p.fftSize * unsafeShift(octaves-1)

	// Now add in the extra padding and compensate for hops that must
	// be dropped in order to align the atom centres across
	// octaves. Again this is a bit trickier because we are doing it
	// at input rather than output and so must work in per-octave
	// sample rates rather than output blocks

	emptyHops := p.firstCentre / p.atomSpacing

	drops := make([]int, octaves, octaves)
	for i := 0; i < octaves; i++ {
		factor := unsafeShift(i)
		dropHops := emptyHops*unsafeShift(octaves-i-1) - emptyHops
		drops[i] = ((dropHops * p.fftHop) * factor) / p.atomsPerFrame
	}

	maxLatPlusDrop := 0
	for i := 0; i < octaves; i++ {
		latPlusDrop := latencies[i] + drops[i]
		if latPlusDrop > maxLatPlusDrop {
			maxLatPlusDrop = latPlusDrop
		}
	}

	totalLatency := maxLatPlusDrop

	lat0 := totalLatency - latencies[0] - drops[0]
	totalLatency = int(math.Ceil(float64((lat0/p.fftHop)*p.fftHop))) + latencies[0] + drops[0]

	// We want (totalLatency - latencies[i]) to be a multiple of 2^i
	// for each octave i, so that we do not end up with fractional
	// octave latencies below. In theory this is hard, in practice if
	// we ensure it for the last octave we should be OK.
	finalOctLat := float64(latencies[octaves-1])
	finalOneFactInt := unsafeShift(octaves - 1)
	finalOctFact := float64(finalOneFactInt)

	totalLatency = int(finalOctLat + finalOctFact*math.Ceil((float64(totalLatency)-finalOctLat)/finalOctFact) + .5)

	if DEBUG_CQ {
		fmt.Printf("total latency = %v\n", totalLatency)
	}

	// Padding as in the reference (will be introduced with the
	// latency compensation in the loop below)
	outputLatency := totalLatency + bigBlockSize - p.firstCentre*unsafeShift(octaves-1)

	if DEBUG_CQ {
		fmt.Printf("bigBlockSize = %v, firstCentre = %v, octaves = %v, so outputLatency = %v\n",
			bigBlockSize, p.firstCentre, octaves, outputLatency)
	}

	buffers := make([][]float64, octaves, octaves)

	for i := 0; i < octaves; i++ {
		factor := unsafeShift(i)

		// Calculate the difference between the total latency applied
		// across all octaves, and the existing latency due to the
		// decimator for this octave, and then convert it back into
		// the sample rate appropriate for the output latency of this
		// decimator -- including one additional big block of padding
		// (as in the reference).

		octaveLatency := float64(totalLatency-latencies[i]-drops[i]+bigBlockSize) / float64(factor)

		if DEBUG_CQ {
			rounded := float64(round(octaveLatency))
			fmt.Printf("octave %d: resampler latency = %v, drop = %v, (/factor = %v), octaveLatency = %v -> %v (diff * factor = %v * %v = %v)\n",
				i, latencies[i], drops[i], drops[i]/factor, octaveLatency, rounded, octaveLatency-rounded, factor, (octaveLatency-rounded)*float64(factor))
		}

		sz := int(octaveLatency + 0.5)
		buffers[i] = make([]float64, sz, sz)
	}

	// m_fft = new FFTReal(m_p.fftSize);
	fmt.Printf("~~~~~~FFT Size = %d\n", kernel.Properties.fftSize)

	return &ConstantQ{
		// params,
		kernel,

		octaves,
		bigBlockSize,
		outputLatency,
		buffers,

		latencies,
		decimators,
	}
}

func round(v float64) int {
	return int(v + 0.5)
}

func (cq *ConstantQ) Process(td []float64, print bool) [][]complex128 {
	for _, v := range td {
		// TODO - faster array append in golang?
		cq.buffers[0] = append(cq.buffers[0], v)
	}
	if print {
		fmt.Printf("buffer 0 size: %d\n", len(cq.buffers[0]))
	}

	for i := 1; i < cq.octaves; i++ {
		dec := cq.decimators[i].Process(td)
		for _, v := range dec {
			cq.buffers[i] = append(cq.buffers[i], v)
		}
		if print {
			fmt.Printf("Buffer %d size: %d\n", i, len(cq.buffers[i]))
		}
	}

	out := [][]complex128{}

	for {
		// We could have quite different remaining sample counts in
		// different octaves, because (apart from the predictable
		// added counts for decimator output on each block) we also
		// have variable additional latency per octave
		enough := true
		for i := 0; i < cq.octaves; i++ {
			required := cq.kernel.Properties.fftSize * unsafeShift(cq.octaves-i-1)
			if len(cq.buffers[i]) < required {
				enough = false
			}
		}
		if !enough {
			break
		}

		base := len(out)
		totalColumns := unsafeShift(cq.octaves-1) * cq.kernel.Properties.atomsPerFrame

		if print {
			fmt.Printf("TOTAL COLUMNS = %d\n", totalColumns)
		}

		for i := 0; i < totalColumns; i++ {
			out = append(out, []complex128{})
		}

		for octave := 0; octave < cq.octaves; octave++ {
			blocksThisOctave := unsafeShift(cq.octaves - octave - 1)

			for b := 0; b < blocksThisOctave; b++ {
				block := cq.processOctaveBlock(octave, print && octave == 0 && b == 0)

				for j := 0; j < cq.kernel.Properties.atomsPerFrame; j++ {
					target := base + (b*(totalColumns/blocksThisOctave) + (j * ((totalColumns / blocksThisOctave) / cq.kernel.Properties.atomsPerFrame)))

					if print && target == 1 {
						fmt.Printf("Adding to 1, octave = %d, block = %d, atom = %d\n", octave, b, j)
					}

					for len(out[target]) < cq.kernel.Properties.binsPerOctave*(octave+1) {
						out[target] = append(out[target], complex(0, 0))
					}

					bpo := cq.kernel.Properties.binsPerOctave
					for i := 0; i < cq.kernel.Properties.binsPerOctave; i++ {
						if print && target == 1 {
							fmt.Printf("out[%d][%d * %d + %d] = block[%d][%d - %d - 1] = (%.4f, %.4f)\n",
								target, bpo, octave, i, j, bpo, i, real(block[j][bpo-i-1]), imag(block[j][bpo-i-1]))
						}
						out[target][cq.kernel.Properties.binsPerOctave*octave+i] = block[j][cq.kernel.Properties.binsPerOctave-i-1]
					}
				}
			}
		}
	}

	return out
}

func (cq *ConstantQ) GetRemainingOutput() [][]complex128 {
	// Same as padding added at start, though rounded up
	pad := int(math.Ceil(float64(cq.OutputLatency)/float64(cq.bigBlockSize))) * cq.bigBlockSize
	zeros := make([]float64, pad, pad)
	return cq.Process(zeros, false)
}

func (cq *ConstantQ) processOctaveBlock(octave int, print bool) [][]complex128 {
	// TODO - merge real pairs into complex array
	// ro, io := make([]float64, cq.kernel.Properties.fftSize, cq.kernel.Properties.fftSize), make([]float64, cq.kernel.Properties.fftSize, cq.kernel.Properties.fftSize)

	if print {
		fmt.Printf("~~ Octave data = %d values: ", len(cq.buffers[octave]))
		maxidx := 0
		for i, v := range cq.buffers[octave] {
			if cq.buffers[octave][maxidx] < v {
				maxidx = i
			}
		}
		fmt.Printf("Max value = buffer[%d] = %.5f\n", maxidx, cq.buffers[octave][maxidx])
		for i := 0; i < 20; i++ {
			fmt.Printf("%.4f, ", cq.buffers[octave][i])
		}
		fmt.Printf("\n")
		sum := 0.0
		for _, v := range cq.buffers[octave] {
			sum += v
		}
		fmt.Printf("Sum is %.5f\n", sum)
	}

	// HACK
	cv := fft.FFTReal(cq.buffers[octave][:cq.kernel.Properties.fftSize])
	// cv := make([]complex128, len(cvFrom))
	// copy(cv, cvFrom)
	if print {
		csum := complex(0, 0)
		for _, v := range cv {
			csum += v
		}
		fmt.Printf("FFT result sum = (%.4f, %.4f)\n", real(csum), imag(csum))
		fmt.Printf("First 10 = ")
		for i := 0; i < 10; i++ {
			fmt.Printf("(%.4f, %.4f), ", real(cv[i]), imag(cv[i]))
		}
		fmt.Printf("\n")
	}

	// if octave == 1 && blockCount == 31 {
	// fmt.Printf("---- %v\n", cq.buffers[octave])
	// }
	// cv := m_fft->forward(m_buffers[octave].data(), ro.data(), io.data());

	lshift := cq.kernel.Properties.fftHop
	// shifted := make([]float64, lshift, lshift)
	// for i := 0; i < lshift; i++ {
	// shifted[i] = cq.buffers[octave][i + cq.kernel.Properties.fftHop]
	// }
	// cq.buffers[octave] = shifted
	if print {
		fmt.Printf("hop = %d, Before / after length = %d / ", lshift, len(cq.buffers[octave]))
	}
	cq.buffers[octave] = cq.buffers[octave][lshift:]
	if print {
		fmt.Printf("%d\n", len(cq.buffers[octave]))
	}

	// ComplexSequence cv;
	// for (int i = 0; i < m_p.fftSize; ++i) {
	// cv.push_back(Complex(ro[i], io[i]));
	// }
	// if (print) {
	//   fmt.Printf("&cv[0] = %p\n", &cv[0])
	//   for i := 0; i < 10; i++ {
	//     fmt.Printf("cv[%d] = %v\n", i, cv[i])
	//   }
	// }
	cqrowvec := cq.kernel.processForward(cv, print)
	if print {
		fmt.Printf("Kernel process, %d values\n", len(cqrowvec))
		reval, imval := 0.0, 0.0
		for _, v := range cqrowvec {
			reval += real(v)
			imval += imag(v)
		}
		fmt.Printf("Kernel sum = (%.4f, %.4f)\n", reval, imval)
	}

	// Reform into a column matrix
	cqblock := make([][]complex128, cq.kernel.Properties.atomsPerFrame, cq.kernel.Properties.atomsPerFrame)
	for j := 0; j < cq.kernel.Properties.atomsPerFrame; j++ {
		cqblock[j] = make([]complex128, cq.kernel.Properties.binsPerOctave, cq.kernel.Properties.binsPerOctave)
		for i := 0; i < cq.kernel.Properties.binsPerOctave; i++ {
			cqblock[j][i] = cqrowvec[i*cq.kernel.Properties.atomsPerFrame+j]
		}
	}

	return cqblock
}

func unsafeShift(s int) int {
	return 1 << uint(s)
}
