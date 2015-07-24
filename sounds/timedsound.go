package sounds

// A timedSound is parameters to the algorithm that limits a sound to a given duration.
type timedSound struct {
	wrapped Sound
	// TODO - switch to sample count.
	durationMs uint64

	durationLeft float64
}

// NewSilence wraps an existing sound as something that stops after a given duration.
//
// For example, to create a sound of middle C that lasts a second:
//	s := sounds.NewTimedSound(sounds.NewSineWave(261.63), 1000)
func NewTimedSound(wrapped Sound, durationMs float64) Sound {
	if float64(wrapped.DurationMs()) < durationMs {
		// TODO - perhaps pad with timed silence?
		panic("Can't time a sound longer than it starts out.")
	}

	data := timedSound{
		wrapped,
		uint64(durationMs),
		durationMs, /* durationLeft */
	}

	return NewBaseSound(&data, data.durationMs)
}

// Run generates the samples by copying the wrapped sound, stopping after the set time.
func (s *timedSound) Run(base *BaseSound) {
	s.wrapped.Start()
	for s.durationLeft > 0.0 {
		value, ok := <-s.wrapped.GetSamples()
		if !ok {
			// NOTE: should not happen.
			break
		}
		if !base.WriteSample(value) {
			break
		}
		s.durationLeft -= SecondsPerCycle() * 1000.0
	}
}

// Stop cleans up the sound by stopping the underlying sound.
func (s *timedSound) Stop() {
	s.wrapped.Stop()
}

// Reset resets the underlying sound, and restarts the duration tracker.
func (s *timedSound) Reset() {
	s.wrapped.Reset()
	s.durationLeft = float64(s.durationMs)
}
