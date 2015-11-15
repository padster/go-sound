package sounds

import (
	"fmt"
)

// A sliceSound is the slice of samples that back the created Sound.
type sliceSound struct {
	samples []float64
}

// WrapSliceAsSound wraps an already created slice of [-1, 1] as a sound.
func WrapSliceAsSound(samples []float64) Sound {
	data := sliceSound{samples}
	return NewBaseSound(&data, uint64(len(samples)))
}

// Run generates the samples by simply iterating through the provided slice.
func (s *sliceSound) Run(base *BaseSound) {
	for _, sample := range s.samples {
		if !base.WriteSample(sample) {
			break
		}
	}
}

// Stop cleans up the sound, in this case doing nothing.
func (s *sliceSound) Stop() {
	// No-op all stopping is done in base.
}

// Reset resets the sound, in this case doing nothing.
func (s *sliceSound) Reset() {
	// No-op, sound gets reset in Run()
}

// String returns the textual representation
func (s *sliceSound) String() string {
	return fmt.Sprintf("SliceSound[%d samples]", len(s.samples))
}
