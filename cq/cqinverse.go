package cq

import (
	"fmt"
	"math/cmplx"

	"github.com/mjibson/go-dsp/fft"
)

const DEBUG_CQI = false

type CQInverse struct {
	kernel *CQKernel

	OutputLatency int
	buffers       [][]float64
	olaBufs       [][]float64

	latencies  []int
	upsamplers []*Resampler
}

func NewCQInverse(params CQParams) *CQInverse {
	kernel := NewCQKernel(params)
	p := kernel.Properties

	// Use exact powers of two for resampling rates. They don't have
	// to be related to our actual samplerate: the resampler only
	// cares about the ratio, but it only accepts integer source and
	// target rates, and if we start from the actual samplerate we
	// risk getting non-integer rates for lower octaves
	sourceRate := unsafeShift(p.octaves)
	latencies := make([]int, p.octaves, p.octaves)
	upsamplers := make([]*Resampler, p.octaves, p.octaves)

	// top octave, no resampling
	latencies[0] = 0
	upsamplers[0] = nil

	for i := 1; i < p.octaves; i++ {
		factor := unsafeShift(i)

		r := NewResampler(sourceRate/factor, sourceRate, 50, 0.05)

		if DEBUG_CQI {
			fmt.Printf("Inverse: octave %d: resample from %d to %d\n", i, sourceRate/factor, sourceRate)
		}

		// See ConstantQ.go for discussion on latency -- output
		// latency here is at target rate which, this way around, is
		// what we want
		latencies[i] = r.GetLatency()
		upsamplers[i] = r
	}

	// additionally we will have fftHop latency at individual octave
	// rate (before upsampling) for the overlap-add in each octave
	for i := 0; i < p.octaves; i++ {
		latencies[i] += p.fftHop * unsafeShift(i)
	}

	// Now reverse the drop adjustment made in ConstantQ to align the
	// atom centres across different octaves (but this time at output
	// sample rate)
	emptyHops := p.firstCentre / p.atomSpacing

	pushes := make([]int, p.octaves, p.octaves)
	for i := 0; i < p.octaves; i++ {
		factor := unsafeShift(i)
		pushHops := emptyHops*unsafeShift(p.octaves-i-1) - emptyHops
		push := ((pushHops * p.fftHop) * factor) / p.atomsPerFrame
		pushes[i] = push
	}

	maxLatLessPush := 0
	for i := 0; i < p.octaves; i++ {
		latLessPush := latencies[i] - pushes[i]
		if latLessPush > maxLatLessPush {
			maxLatLessPush = latLessPush
		}
	}
	totalLatency := maxLatLessPush + 10
	if totalLatency < 0 {
		totalLatency = 0
	}

	outputLatency := totalLatency + p.firstCentre*unsafeShift(p.octaves-1)
	if DEBUG_CQI {
		fmt.Printf("totalLatency = %d, outputLatency = %d\n", totalLatency, outputLatency)
	}

	buffers := make([][]float64, p.octaves)
	for i := 0; i < p.octaves; i++ {
		// Calculate the difference between the total latency applied
		// across all octaves, and the existing latency due to the
		// upsampler for this octave.
		latencyPadding := totalLatency - latencies[i] + pushes[i]
		if DEBUG_CQI {
			fmt.Printf("octave %d: push %d, resampler latency inc overlap space %d, latencyPadding %d (/factor = %d)\n",
				i, pushes[i], latencies[i], latencyPadding, latencyPadding/unsafeShift(i))
		}

		buffers[i] = make([]float64, latencyPadding, latencyPadding)
	}

	olaBufs := make([][]float64, p.octaves, p.octaves)
	for i := 0; i < p.octaves; i++ {
		// Fixed-size buffer for IFFT overlap-add
		olaBufs[i] = make([]float64, p.fftSize, p.fftSize)
	}

	return &CQInverse{
		kernel,

		outputLatency,
		buffers,
		olaBufs,

		latencies,
		upsamplers,
	}
}

// Return clamped!
func (cqi *CQInverse) ProcessChannel(blocks <-chan []complex128) <-chan float64 {
	result := make(chan float64)

	apf := cqi.kernel.Properties.atomsPerFrame
	octaves := cqi.kernel.Properties.octaves

	required := apf * unsafeShift(octaves-1) // * some integer multiple?

	go func() {
		buffer := make([][]complex128, required, required)
		at := 0
		// HACK := 0
		for s := range blocks {
			// if HACK % 10 == 0 {
				// fmt.Printf("Inverting column %d, blocks of %d\n", HACK, required)
			// }
			if at == required {
				fmt.Printf("Purging CQI block!\n")
				for _, c := range cqi.Process(buffer) {
					result <- c
				}
				at = 0
			}
			buffer[at] = s
			at++
			// HACK++
		}
		/*
			HACK - figure out how to pad this properly...
			Currently we have buffer[:at], but it is too short to cqi.Process()
			for _, c := range cqi.Process(buffer[:at]) {
				result <- c
			}
		*/
		for _, c := range cqi.GetRemainingOutput() {
			result <- c
		}
		close(result)
	}()

	return result
}

// Return clamped!
func (cqi *CQInverse) Process(block [][]complex128) []float64 {
	apf := cqi.kernel.Properties.atomsPerFrame
	bpo := cqi.kernel.Properties.binsPerOctave
	octaves := cqi.kernel.Properties.octaves

	// The input data is of the form produced by ConstantQ::process --
	// an unknown number N of columns of varying height. We assert
	// that N is a multiple of atomsPerFrame * 2^(octaves-1), as must
	// be the case for data that came directly from our ConstantQ
	// implementation.
	widthProvided := len(block)
	if widthProvided == 0 {
		return cqi.drawFromBuffers()
	}

	blockWidth := apf * unsafeShift(octaves-1)
	if widthProvided%blockWidth != 0 {
		fmt.Printf("ERROR: inverse process block size (%d) must be a multiple of atoms * 2^(octaves - 1) = %d * 2^(%d - 1) = %d\n",
			widthProvided, apf, octaves, blockWidth)
		panic("CQI process block size incorrect")
	}

	// Procedure:
	//
	// 1. Slice the list of columns into a set of lists of columns,
	// one per octave, each of width N / (2^octave-1) and height
	// binsPerOctave, containing the values present in that octave
	//
	// 2. Group each octave list by atomsPerFrame columns at a time,
	// and stack these so as to achieve a list, for each octave, of
	// taller columns of height binsPerOctave * atomsPerFrame
	//
	// 3. For each taller column, take the product with the inverse CQ
	// kernel (which is the conjugate of the forward kernel) and
	// perform an inverse FFT
	//
	// 4. Overlap-add each octave's resynthesised blocks (unwindowed)
	//
	// 5. Resample each octave's overlap-add stream to the original
	// rate
	//
	// 6. Sum the resampled streams and return
	for i := 0; i < octaves; i++ {
		// Step 1
		oct := make([][]complex128, 0)

		for j := 0; j < widthProvided; j++ {
			h := len(block[j])
			if h < bpo*(i+1) {
				// TODO: Figure if j loop can be done with steps to only those J needed.
				continue
			}

			oct = append(oct, block[j][bpo*i:bpo*(i+1)])
		}

		// Steps 2, 3, 4, 5
		cqi.processOctave(i, oct)
	}

	// Step 6
	return cqi.drawFromBuffers()
}

// Return clamped!
func (cqi *CQInverse) drawFromBuffers() []float64 {
	octaves := cqi.kernel.Properties.octaves

	// 6. Sum the resampled streams and return
	available := len(cqi.buffers[0])
	for i := 1; i < octaves; i++ {
		available = minInt(available, len(cqi.buffers[i]))
	}

	result := make([]float64, available, available)
	if available > 0 {
		for i := 0; i < octaves; i++ {
			for j := 0; j < available; j++ {
				result[j] += cqi.buffers[i][j]
			}
			cqi.buffers[i] = cqi.buffers[i][available:]
		}
		for i, v := range result {
			result[i] = clampUnit(v)
		}
	}
	fmt.Printf("%d samples inverted\n", len(result))
	return result
}

// Return clamped!
func (cqi *CQInverse) GetRemainingOutput() []float64 {
	bpo := cqi.kernel.Properties.binsPerOctave
	fftHop := cqi.kernel.Properties.fftHop
	octaves := cqi.kernel.Properties.octaves

	for j := 0; j < octaves; j++ {
		factor := unsafeShift(j)
		latency := 0
		if j > 0 {
			latency = cqi.upsamplers[j].GetLatency() / factor
		}

		for i := 0; i < (latency+bpo)/fftHop; i++ {
			padding := make([]float64, len(cqi.olaBufs[j]), len(cqi.olaBufs[j]))
			cqi.overlapAddAndResample(j, padding)
		}
	}
	return cqi.drawFromBuffers()
}

func (cqi *CQInverse) processOctave(octave int, columns [][]complex128) {
	// 2. Group each octave list by atomsPerFrame columns at a time,
	// and stack these so as to achieve a list, for each octave, of
	// taller columns of height binsPerOctave * atomsPerFrame

	bpo := cqi.kernel.Properties.binsPerOctave
	apf := cqi.kernel.Properties.atomsPerFrame

	ncols := len(columns)
	if ncols%apf != 0 {
		fmt.Printf("Error: inverse process octave %d, # columns (%d) must be a multiple of atoms/frame (%d)\n",
			octave, ncols, apf)
		panic("Invalid argument to inverse processOctave")
	}

	for i := 0; i < ncols; i += apf {
		tallCol := make([]complex128, bpo*apf, bpo*apf)

		for b := 0; b < bpo; b++ {
			for a := 0; a < apf; a++ {
				tallCol[b*apf+a] = columns[i+a][bpo-b-1]
			}
		}

		cqi.processOctaveColumn(octave, tallCol)
	}
}

func (cqi *CQInverse) processOctaveColumn(octave int, column []complex128) {
	// 3. For each taller column, take the product with the inverse CQ
	// kernel (which is the conjugate of the forward kernel) and
	// perform an inverse FFT
	bpo := cqi.kernel.Properties.binsPerOctave
	apf := cqi.kernel.Properties.atomsPerFrame
	fftSize := cqi.kernel.Properties.fftSize

	if len(column) != bpo*apf {
		fmt.Printf("Error: column in octave %d has size %d, required = %d * %d = %d\n",
			octave, len(column), bpo, apf, bpo*apf)
		panic("Invalid argument to inverse processOctaveColumn")
	}

	transformed := cqi.kernel.ProcessInverse(column)

	// For inverse real transforms, force symmetric conjugate representation first.
	for i := fftSize/2 + 1; i < fftSize; i++ {
		transformed[i] = cmplx.Conj(transformed[fftSize-i])
	}

	inverse := fft.IFFT(transformed)
	cqi.overlapAddAndResample(octave, realParts(inverse))
}

func (cqi *CQInverse) overlapAddAndResample(octave int, seq []float64) {
	// 4. Overlap-add each octave's resynthesised blocks (unwindowed)
	//
	// and
	//
	// 5. Resample each octave's overlap-add stream to the original
	// rate

	fftHop := cqi.kernel.Properties.fftHop
	fftSize := cqi.kernel.Properties.fftSize

	if len(seq) != len(cqi.olaBufs[octave]) {
		fmt.Printf("Error: inverse overlap add sequence size %d, expected to be OLA buffer size %d\n",
			len(seq), len(cqi.olaBufs[octave]))
		panic("Illegal argument to Inverse overlapAdd")
	}

	toResample := cqi.olaBufs[octave][:fftHop]

	resampled := toResample
	if octave > 0 {
		resampled = cqi.upsamplers[octave].Process(toResample)
	}

	cqi.buffers[octave] = append(cqi.buffers[octave], resampled...)
	cqi.olaBufs[octave] = append(
		cqi.olaBufs[octave][fftHop:],
		make([]float64, fftHop, fftHop)...,
	)

	for i := 0; i < fftSize; i++ {
		cqi.olaBufs[octave][i] += seq[i]
	}
}
