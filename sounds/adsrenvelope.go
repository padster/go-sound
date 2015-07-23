/*
Envelope based on Attack, Decay, Sustain, Release config.
https://en.wikipedia.org/wiki/Synthesizer#ADSR_envelope
*/
package sounds

import "fmt"

type ADSREnvelope struct {
	wrapped Sound

	floatDuration  float64
	msAt           float64
	attackMs       float64
	sustainStartMs float64
	sustainEndMs   float64
	sustainLevel   float64
}

func NewADSREnvelope(wrapped Sound,
	attackMs float64, delayMs float64, sustainLevel float64, releaseMs float64) Sound {

	durationMs := wrapped.DurationMs()

	envelope := ADSREnvelope{
		wrapped,
		float64(durationMs),             /* floatDuration */
		0,                               /* msAt */
		attackMs,                        /* attackMs */
		attackMs + delayMs,              /* sustainStartMs */
		float64(durationMs) - releaseMs, /* sustainEndMs */
		sustainLevel,
	}

	return NewBaseSound(&envelope, durationMs)
}

func (s *ADSREnvelope) Run(base *BaseSound) {
	s.wrapped.Start()

	for s.msAt < s.floatDuration {
		scale := float64(0) // [0, 1]

		switch {
		case s.msAt < s.attackMs:
			scale = s.msAt / s.attackMs
		case s.msAt < s.sustainStartMs:
			scale = 1 - (1-s.sustainLevel)*(s.msAt-s.attackMs)/(s.sustainStartMs-s.attackMs)
		case s.msAt < s.sustainEndMs:
			scale = s.sustainLevel
		default:
			scale = s.sustainLevel * (s.msAt - s.floatDuration) / (s.sustainEndMs - s.floatDuration)
		}

		next := <-s.wrapped.GetSamples()
		if !base.WriteSample(next * scale) {
			return
		}
		s.msAt += SecondsPerCycle() * 1000.0
	}
}

func (s *ADSREnvelope) Stop() {
	fmt.Println("Stop adsr")
	s.wrapped.Stop()
}

func (s *ADSREnvelope) Reset() {
	fmt.Println("Reset adsr")
	s.wrapped.Reset()
	s.msAt = float64(0)
}
