package sounds

import (
	"math"
)

// A silence is parameters to the algorithm that generates silence.
type silence struct{}

// NewSilence creates an unending sound that is inaudible.
func NewSilence() Sound {
	data := silence{}
	return NewBaseSound(&data, math.MaxUint64)
}

// NewTimedSilence creates a silence that lasts for a given duration.
//
// For example, Cage's 4'33" can be generated using:
//  s := sounds.NewTimedSilence(273000)
func NewTimedSilence(durationMs float64) Sound {
	return NewTimedSound(NewSilence(), durationMs)
}

// Run generates the samples by continuously writing 0 (silence).
func (s *silence) Run(base *BaseSound) {
	for base.WriteSample(0) {
		// No-op
	}
}

// Stop cleans up the silence, in this case doing nothing.
func (s *silence) Stop() {
	// No-op
}

// Reset does nothing in the case of silence, as there is no state.
func (s *silence) Reset() {
	// No-op
}
