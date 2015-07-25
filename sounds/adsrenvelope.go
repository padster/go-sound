package sounds

import (
	"time"
)

// An adsrSound is parameters to the algorithm that applies an
// Attack/Decay/Sustain/Release envelope over a sound.
//
// See https://en.wikipedia.org/wiki/Synthesizer#ADSR_envelope
type adsrEnvelope struct {
	wrapped        Sound

	attackSamples uint64
	sustainStartSamples uint64
	sustainEndSamples uint64
	sampleCount uint64
	sustainLevel float64
}

// NewADSREnvelope wraps an existing sound with a parametric envelope.
//
// For details, read https://en.wikipedia.org/wiki/Synthesizer#ADSR_envelope
//
// For example, to create an envelope around an A440 note
//	s := sounds.NewADSREnvelope(
//		sounds.NewTimedSound(sounds.NewSineWave(261.63), 1000),
//		50, 200, 0.5, 100)
func NewADSREnvelope(wrapped Sound,
	attackMs float64, delayMs float64, sustainLevel float64, releaseMs float64) Sound {
	// NOTE: params are is Ms - time.Duration is possible, but likely more verbose.

	sampleCount := wrapped.Length()
	attack := time.Duration(attackMs) * time.Millisecond
	delay := time.Duration(delayMs) * time.Millisecond
	release := time.Duration(releaseMs) * time.Millisecond

	data := adsrEnvelope{
		wrapped,
		DurationToSamples(attack), /* attackSamples */
		DurationToSamples(attack + delay),              /* sustainStartMs */
		sampleCount - DurationToSamples(release), /* sustainEndMs */
		sampleCount,
		sustainLevel,
	}

	return NewBaseSound(&data, sampleCount)
}

// Run generates the samples by scaling the wrapped sound by the relevant envelope part.
func (s *adsrEnvelope) Run(base *BaseSound) {
	s.wrapped.Start()

	attackDelta := 1.0 / float64(s.attackSamples)
	decayDelta := 1.0 / float64(s.sustainStartSamples - s.attackSamples)
	releaseDelta := 1.0 / float64(s.sampleCount - s.sustainEndSamples)

	for at := uint64(0); at < s.sampleCount; at++ {
		scale := float64(0) // [0, 1]

		// NOTE: this could be split into multiple loops but it doesn't seem worth optimizing currently.
		switch {
		case at < s.attackSamples:
			scale = float64(at) * attackDelta
		case at < s.sustainStartSamples:
			scale = 1 - (1-s.sustainLevel) * decayDelta * float64(at-s.attackSamples)
		case at < s.sustainEndSamples:
			scale = s.sustainLevel
		default:
			scale = s.sustainLevel * releaseDelta * float64(s.sampleCount - at)
		}

		next, ok := <-s.wrapped.GetSamples()
		if !ok || !base.WriteSample(next*scale) {
			return
		}
	}
}

// Stop cleans up the sound by stopping the underlying sound.
func (s *adsrEnvelope) Stop() {
	s.wrapped.Stop()
}

// Reset resets the underlying sound, and zeroes the position in the envelope.
func (s *adsrEnvelope) Reset() {
	s.wrapped.Reset()
}
