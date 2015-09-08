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
	minFrequency  float64
	maxFrequency  float64
	binsPerOctave int

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

func NewCQParams(sampleRate float64, minFreq float64, maxFreq float64, binsPerOctave int) CQParams {
	if maxFreq <= minFreq || minFreq <= 0 {
		panic("Requires frequencies 0 < min < max")
	}

	return CQParams{
		sampleRate,
		minFreq,
		maxFreq,
		binsPerOctave,
		1.0,                /* Q scaling factor */
		0.25,               /* hop size of shortest temporal atom. */
		0.0005,             /* sparcity threshold for resulting kernal. */
		SqrtBlackmanHarris, /* window shape */
	}
}
