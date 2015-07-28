package sounds

import (
	"fmt"

	"github.com/padster/go-sound/types"
)

// A denseIIR is parameters to an Infinite Impulse Response filter, which generates
// new output samples through a linear combination of previous input and output samples.
type denseIIR struct {
	wrapped   Sound
	inCoef    []float64
	outCoef   []float64
	inBuffer  *types.Buffer
	outBuffer *types.Buffer
}

// NewDenseIIR wrapps a sound in an IIR filter, as specified by the coefficients.
// TODO(padster): Also implement the filter design algorithms, e.g:
//   http://engineerjs.com/?sidebar=docs/iir.html
//   http://www.mikroe.com/chapters/view/73/chapter-3-iir-filters/
//   http://www-users.cs.york.ac.uk/~fisher/mkfilter/
//
// For example, to use a high-pass filter for 800hz+ with sample rate of 44.1k:
//  sound := s.NewDenseIIR(...some sound...,
//    []float64{0.8922, -2.677, 2.677, -0.8922},
//    []float64{2.772, -2.57, 0.7961},
//  )
func NewDenseIIR(wrapped Sound, inCoef []float64, outCoef []float64) Sound {
	// TODO(padster): Verify this is doing what it should...hard to tell just by listening.
	data := denseIIR{
		wrapped,
		inCoef,
		outCoef,
		types.NewBuffer(len(inCoef)),
		types.NewBuffer(len(outCoef)),
	}
	return NewBaseSound(&data, wrapped.Length())
}

// Run generates the samples by applying the convolution of the coefs against input/output buffers.
func (s *denseIIR) Run(base *BaseSound) {
	s.wrapped.Start()
	for sample := range s.wrapped.GetSamples() {
		s.inBuffer.Push(sample)

		value := 0.0
		for iX, coefX := range s.inCoef {
			value += coefX * s.inBuffer.GetFromEnd(iX)
		}
		for iY, coefY := range s.outCoef {
			value += coefY * s.outBuffer.GetFromEnd(iY)
		}
		if !base.WriteSample(value) {
			break
		}
	}
}

// Stop cleans up the sound by stopping the underlying sound.
func (s *denseIIR) Stop() {
	s.wrapped.Stop()
}

// Reset resets the underlying sound, and clears the buffer state.
func (s *denseIIR) Reset() {
	s.wrapped.Reset()
	s.inBuffer.Clear()
	s.outBuffer.Clear()
}

// String returns the textual representation
func (s *denseIIR) String() string {
	// TODO(padster): Pass in and use e.g. "Lowpass" etc instead.
	return fmt.Sprintf("DenseIIR[%s]", s.wrapped) // Coefs omitted for brevity
}
