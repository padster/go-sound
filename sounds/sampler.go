package sounds

import (
	"fmt"
)

// A linearSampler is parameters to the algorithm that forms a sound by
// sampling a second sound, and linearly interpolating that to a different sample rate.
type linearSampler struct {
	wrapped    Sound
	pitchScale float64
}

// LinearSample wraps an existing sound and samples it at a different rate,
// modifying both its pitch and duration.
//
// For example, to modulate a sound up an octave, and make it half as long:
//  s := ...some sound...
//  higher := sounds.LinearSample(s, 2.0)
func LinearSample(wrapped Sound, pitchScale float64) Sound {
	newLength := MaxLength
	if wrapped.Length() < MaxLength {
		newLength = uint64(float64(wrapped.Length()) / pitchScale)
	}

	data := linearSampler{
		wrapped,
		pitchScale,
	}
	return NewBaseSound(&data, newLength)
}

// Run generates the samples by iterating through the origin, and
// resampling at the required rate, linearly interpolating to calculate the new samples.
func (s *linearSampler) Run(base *BaseSound) {
	s.wrapped.Start()

	last := float64(0.0)

	for at := float64(0.0); true; at -= 1.0 {
		current, ok := <-s.wrapped.GetSamples()
		if !ok {
			break
		}

		for ; -1-1e-9 < at && at < 1e-9; at += s.pitchScale {
			// at == -1 -> last, at == 0 -> current, so:
			sample := current + at*(current-last)
			if !base.WriteSample(sample) {
				break
			}
		}

		last = current
	}
}

// Stop cleans up the sound by stopping the underlying sound.
func (s *linearSampler) Stop() {
	s.wrapped.Stop()
}

// Reset resets the underlying sound, and restarts the sample tracker.
func (s *linearSampler) Reset() {
	s.wrapped.Reset()
}

// String returns the textual representation
func (s *linearSampler) String() string {
	return fmt.Sprintf("Sampled[%s at %.2f]", s.wrapped, s.pitchScale)
}
