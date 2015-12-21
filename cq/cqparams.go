package cq

type Window int

const (
	SqrtBlackmanHarris Window = iota
	SqrtBlackman
	SqrtHann
	BlackmanHarris
	Blackman
	Hann
)

type CQParams struct {
	sampleRate    float64
	Octaves       int
	minFrequency  float64
	BinsPerOctave int

	// Spectral atom bandwidth scaling;1 is optimal for reconstruction,
	// q < 1 increases smearing in frequency domain but improves the time domain.
	q float64

	// Hop size between temporal atoms; 1 == no overlap, smaller = overlapping.
	atomHopFactor float64

	// Values smaller than this are zeroed in the kernel.
	threshold float64

	// Window shape for kernal atoms.
	window Window
}

func NewCQParams(sampleRate float64, octaves int, minFreq float64, binsPerOctave int) CQParams {
	if minFreq < 0 || octaves < 1 {
		panic("Requires frequencies 0 < min, octaves ")
	}

	return CQParams{
		sampleRate,
		octaves,
		minFreq,
		binsPerOctave,
		1.0,                /* Q scaling factor */
		0.25,               /* hop size of shortest temporal atom. */
		0.0005,             /* sparcity threshold for resulting kernal. */
		SqrtBlackmanHarris, /* window shape */
	}
}
