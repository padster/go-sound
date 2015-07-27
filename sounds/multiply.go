package sounds

import (
  "fmt"
)

// A multiply is parameters to the algorithm that scales the amplitude of a sound.
type multiply struct {
  wrapped     Sound
  factor float64
}

// MultiplyWithClip wraps an existing sound and scales its amplitude by a given factory,
// clipping the result to [-1, 1].
//
// For example, to create a sound half as loud as the default E5 sine wave:
//  s := sounds.MultiplyWithClip(sounds.NewSineWave(659.25), 0.5)
func MultiplyWithClip(wrapped Sound, factor float64) Sound {
  data := multiply{
    wrapped,
    factor,
  }

  return NewBaseSound(&data, wrapped.Length())
}

// Run generates the samples by scaling the wrapped sound's samples, clipping to the valid range.
func (s *multiply) Run(base *BaseSound) {
  s.wrapped.Start()
  for sample := range s.wrapped.GetSamples() {
    scaled := s.factor * sample
    if scaled > 1 {
      scaled = 1.0
    } else if scaled < -1 {
      scaled = -1
    }
    if !base.WriteSample(scaled) {
      break
    }
  }
}

// Stop cleans up the sound by stopping the underlying sound.
func (s *multiply) Stop() {
  s.wrapped.Stop()
}

// Reset resets the underlying sound.
func (s *multiply) Reset() {
  s.wrapped.Reset()
}

// String returns the textual representation
func (s *multiply) String() string {
  return fmt.Sprintf("Multiple[%s scaled by %.2f]", s.wrapped, s.factor)
}
