package sounds

import (
	"time"
)

// A timedSound is parameters to the algorithm that limits a sound to a given duration.
type timedSound struct {
	wrapped Sound
	sampleCount uint64
	sampleAt uint64
}

// NewSilence wraps an existing sound as something that stops after a given duration.
//
// For example, to create a sound of middle C that lasts a second:
//	s := sounds.NewTimedSound(sounds.NewSineWave(261.63), 1000)
func NewTimedSound(wrapped Sound, durationMs float64) Sound {
	// NOTE: duration is Ms - time.Duration is possible, but likely more verbose.
	duration := time.Duration(int64(durationMs * 1e6)) * time.Nanosecond
	sampleCount := DurationToSamples(duration)

	if wrapped.Length() < sampleCount {
		// TODO - perhaps pad with timed silence?
		panic("Can't time a sound longer than it starts out.")
	}

	data := timedSound{
		wrapped,
		sampleCount,
		0, /* sampleAt */
	}

	return NewBaseSound(&data, sampleCount)
}

// Run generates the samples by copying the wrapped sound, stopping after the set time.
func (s *timedSound) Run(base *BaseSound) {
	s.wrapped.Start()
	for ; s.sampleAt < s.sampleCount; s.sampleAt++ {
		value, ok := <-s.wrapped.GetSamples()
		if !ok {
			// NOTE: should not happen.
			break
		}
		if !base.WriteSample(value) {
			break
		}
	}
}

// Stop cleans up the sound by stopping the underlying sound.
func (s *timedSound) Stop() {
	s.wrapped.Stop()
}

// Reset resets the underlying sound, and restarts the sample tracker.
func (s *timedSound) Reset() {
	s.wrapped.Reset()
	s.sampleAt = 0
}
