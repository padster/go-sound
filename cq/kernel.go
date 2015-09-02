package cq 

import (
	// "fmt"
	"math"
	"math/cmplx"
)

type Properties struct {
	sampleRate float64
    maxFrequency float64
    minFrequency float64
    binsPerOctave int
    fftSize int
    fftHop int
    atomsPerFrame int
    atomSpacing int
    firstCentre int
    lastCentre int
    Q float64
}

type Kernel struct {
	origin []int
	data [][]complex128
}

type CQKernel struct {
	properties Properties
	kernel Kernel
}


// TODO - clean up a lot.
func NewCQKernel(params CQParams) CQKernel {
	// Constructor
	p := Properties{}
	p.sampleRate = params.sampleRate
	p.maxFrequency = params.maxFrequency
	p.binsPerOctave = params.binsPerOctave

	// GenerateKernel
	q := params.q
	atomHopFactor := params.atomHopFactor
	thresh := params.threshold
	bpo := params.binsPerOctave

	p.minFrequency = float64(math.Pow(2, 1 / float64(bpo)) * float64(params.maxFrequency) / 2.0)
	p.Q = q / (math.Pow(2, 1.0 / float64(bpo)) - 1.0)

	maxNK := float64(int(math.Floor(p.Q * p.sampleRate / p.minFrequency + 0.5)))
	minNK := float64(int(math.Floor(p.Q * p.sampleRate / 
		(p.minFrequency * math.Pow(2.0, (float64(bpo) - 1.0) / float64(bpo))) + 0.5)))

	if (minNK == 0 || maxNK == 0) {
		panic("Kernal minNK or maxNK is 0, can't make kernel")
	}

	p.atomSpacing = int(minNK * atomHopFactor + 0.5)
	p.firstCentre = p.atomSpacing * int(math.Ceil(math.Ceil(maxNK / 2.0) / float64(p.atomSpacing)))
	p.fftSize = NextPowerOf2(p.firstCentre + int(math.Ceil(maxNK / 2.0)))
	p.atomsPerFrame = int(math.Floor(
		1.0 + (float64(p.fftSize) - math.Ceil(maxNK / 2.0) - float64(p.firstCentre)) / float64(p.atomSpacing)))
	p.lastCentre = p.firstCentre + (p.atomsPerFrame - 1) * p.atomSpacing
	p.fftHop = (p.lastCentre + p.atomSpacing) - p.firstCentre

	// POIUY
	// FFT := New FFT(p.fftSize)

	dataSize := p.binsPerOctave * p.atomsPerFrame

	kernel := Kernel{
		make([]int, 0, dataSize),
		make([][]complex128, 0, dataSize),
	}

	for k := 1; k <= p.binsPerOctave; k++ {
		nk := int(p.Q * p.sampleRate / (p.minFrequency * math.Pow(2, ((float64(k) - 1.0) / float64(bpo)))) + 0.5)

		win := makeWindow(nk)

		fk := float64(p.minFrequency * math.Pow(2, ((float64(k) - 1.0) / float64(bpo))))

		reals, imags := make([]float64, nk, nk), make([]float64, nk, nk)
		for i := 0; i < nk; i++ {
			arg := float64((2.0 * math.Pi * fk * float64(i)) / p.sampleRate)
			reals[i] = win[i] * math.Cos(arg);
			imags[i] = win[i] * math.Sin(arg);
		}

		atomOffset := int(p.firstCentre - int(math.Ceil(float64(nk) / 2.0)))

		for i := 0; i < p.atomsPerFrame; i++ {
			shift := atomOffset + (i * p.atomSpacing);

			// TODO - verify these are zero outside.
			rin, iin := make([]float64, p.fftSize, p.fftSize), make([]float64, p.fftSize, p.fftSize)
			for j := 0; j < nk; j++ {
				rin[j + shift] = reals[j];
				iin[j + shift] = imags[j];
			}

			rout, iout := make([]float64, p.fftSize, p.fftSize), make([]float64, p.fftSize, p.fftSize)

			// TODO - process FFT
			// m_fft->process(false, rin.data(), iin.data(), rout.data(),
					// iout.data());

			// Keep this dense for the moment (until after normalisation calculations)
			row := make([]complex128, p.fftSize, p.fftSize)
			for j := 0; j < p.fftSize; j++ {
				// TODO - simplify
				if math.Sqrt(rout[j] * rout[j] + iout[j] * iout[j]) < thresh {
					row[j] = complex(0, 0)
				} else {
					row[j] = complex(rout[j] / float64(p.fftSize), iout[j] / float64(p.fftSize))
				}
			}


			kernel.origin = append(kernel.origin, 0)
			kernel.data = append(kernel.data, row)
		}
	}

	// finalizeKernel

	// calculate weight for normalisation
	wx1 := maxidx(kernel.data[0]);
	wx2 := maxidx(kernel.data[len(kernel.data) - 1]);

	subset := make([][]complex128, len(kernel.data), len(kernel.data))
	for i := 0; i < len(kernel.data); i++ {
		subset[i] = make([]complex128, 0, wx2 - wx1 + 1)
	}
	for j := wx1; j <= wx2; j++ {
		for i := 0; i < len(kernel.data); i++ {
			subset[i] = append(subset[i], kernel.data[i][j])
		}
	}

	// Massive hack - precalculate above instead :(
	nrows, ncols := len(subset), len(subset[0])

	square := make([][]complex128, ncols, ncols) // conjugate transpose of subset * subset
	for i := 0; i < ncols; i++ {
		square[i] = make([]complex128, ncols, ncols)
	}

	for j := 0; j < ncols; j++ {
		for i := 0; i < ncols; i++ {
			v := complex(0, 0)
			for k := 0; k < nrows; k++ {
				v += subset[k][i] * cmplx.Conj(subset[k][j])
			}
			square[i][j] = v
		}
	}

	wK := []float64{}
	for i := int(1.0 / q + 0.5); i < ncols - int(1.0 / q + 0.5) - 2; i++ {
		wK = append(wK, cmplx.Abs(square[i][i]))
	}

	weight := float64(p.fftHop) / float64(p.fftSize)
	if len(wK) > 0 {
		weight /= mean(wK)
	}
	weight = math.Sqrt(weight)


	// apply normalisation weight, make sparse, and store conjugate
	// (we use the adjoint or conjugate transpose of the kernel matrix
	// for the forward transform, the plain kernel for the inverse
	// which we expect to be less common)

	sk := Kernel{
		make([]int, len(kernel.data), len(kernel.data)),
		make([][]complex128, len(kernel.data), len(kernel.data)),
	}
	for i := 0; i < len(kernel.data); i++ {
		sk.origin[i] = 0;
		sk.data[i] = []complex128{}

		lastNZ := 0;
		for j := len(kernel.data[i]) - 1; j >= 0; j-- {
			if cmplx.Abs(kernel.data[i][j]) != 0 {
				lastNZ = j
				break;
			}
		}

		haveNZ := false
		for j := 0; j <= lastNZ; j++ {
			if haveNZ || cmplx.Abs(kernel.data[i][j]) != 0 {
				if !haveNZ {
					sk.origin[i] = j;
				}
				haveNZ = true
				sk.data[i] = append(sk.data[i], 
					complexTimes(cmplx.Conj(kernel.data[i][j]), weight))
			}
		}
	}

	return CQKernel {p, sk}
}

func (k CQKernel) processForward(cv []complex128) []complex128 {
	// straightforward matrix multiply (taking into account m_kernel's
	// slightly-sparse representation)

	if len(k.kernel.data) == 0 {
		panic("Whoops - return empty array? is this even possible?")
	}

	nrows := k.properties.binsPerOctave * k.properties.atomsPerFrame

	rv := make([]complex128, nrows, nrows)
	for i := 0; i < nrows; i++ {
		rv[i] = complex(0, 0)
		for j := 0; j < len(k.kernel.data[i]); j++ {
			rv[i] += cv[j + k.kernel.origin[i]] * k.kernel.data[i][j];
		}
	}
	return rv;
}


func (k CQKernel) processInverse(cv []complex128) []complex128 {
	// matrix multiply by conjugate transpose of m_kernel. This is
	// actually the original kernel as calculated, we just stored the
	// conjugate-transpose of the kernel because we expect to be doing
	// more forward transforms than inverse ones.
	if len(k.kernel.data) == 0 {
		panic("Whoops - return empty array? is this even possible?")
	}

	ncols := k.properties.binsPerOctave * k.properties.atomsPerFrame
	nrows := k.properties.fftSize;

	rv := make([]complex128, nrows, nrows)
	for j := 0; j < ncols; j++ {
		i0 := k.kernel.origin[j];
		i1 := i0 + len(k.kernel.data[j]);
		for i := i0; i < i1; i++ {
			rv[i] += cv[j] * cmplx.Conj(k.kernel.data[j][i - i0])
		}
	}
	return rv;
}

func makeWindow(len int) []float64 {
	// HACK - need to do
	result := make([]float64, len, len)
	return result
}


// Utilities - TODO split?
func mean(fs []float64) float64 {
	s := 0.0
	for _, v := range fs {
		s += v
	}
	return s / float64(len(fs))
}


func maxidx(row []complex128) int {
	idx, max := 0, cmplx.Abs(row[0])
	for i, v := range row {
		vAbs := cmplx.Abs(v)
		if vAbs > max {
			idx, max = i, vAbs
		}
	}
	return idx
}

func complexTimes(c complex128, f float64) complex128 {
	return complex(real(c) * f, imag(c) * f)
}

// IsPowerOf2 returns true if x is a power of 2, else false.
func IsPowerOf2(x int) bool {
	return x&(x-1) == 0
}

// NextPowerOf2 returns the next power of 2 >= x.
func NextPowerOf2(x int) int {
	if IsPowerOf2(x) {
		return x
	}

	return int(math.Pow(2, math.Ceil(math.Log2(float64(x)))))
}
