package sounds

// An adsrSound is parameters to the algorithm that applies an
// Attack/Decay/Sustain/Release envelope over a sound.
//
// See https://en.wikipedia.org/wiki/Synthesizer#ADSR_envelope
type adsrEnvelope struct {
	wrapped        Sound
	floatDuration  float64
	attackMs       float64
	sustainStartMs float64
	sustainEndMs   float64
	sustainLevel   float64

	msAt float64
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

	durationMs := wrapped.DurationMs()

	data := adsrEnvelope{
		wrapped,
		float64(durationMs),             /* floatDuration */
		attackMs,                        /* attackMs */
		attackMs + delayMs,              /* sustainStartMs */
		float64(durationMs) - releaseMs, /* sustainEndMs */
		sustainLevel,
		0, /* msAt */
	}

	return NewBaseSound(&data, durationMs)
}

// Run generates the samples by scaling the wrapped sound by the relevant envelope part.
func (s *adsrEnvelope) Run(base *BaseSound) {
	s.wrapped.Start()

	for s.msAt < s.floatDuration {
		scale := float64(0) // [0, 1]

		// PICK: pre-calculate denominators?
		switch {
		case s.msAt < s.attackMs:
			scale = s.msAt / s.attackMs
		case s.msAt < s.sustainStartMs:
			scale = 1 - (1-s.sustainLevel)*
				(s.msAt-s.attackMs)/(s.sustainStartMs-s.attackMs)
		case s.msAt < s.sustainEndMs:
			scale = s.sustainLevel
		default:
			scale = s.sustainLevel *
				(s.msAt - s.floatDuration) / (s.sustainEndMs - s.floatDuration)
		}

		next, ok := <-s.wrapped.GetSamples()
		if !ok || !base.WriteSample(next*scale) {
			return
		}
		s.msAt += SecondsPerCycle() * 1000.0
	}
}

// Stop cleans up the sound by stopping the underlying sound.
func (s *adsrEnvelope) Stop() {
	s.wrapped.Stop()
}

// Reset resets the underlying sound, and zeroes the position in the envelope.
func (s *adsrEnvelope) Reset() {
	s.wrapped.Reset()
	s.msAt = float64(0)
}
