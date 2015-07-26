package sounds

import (
	"fmt"
	"time"
)

// A timedSound is parameters to the algorithm that limits a sound to a given duration.
type timedSound struct {
	wrapped     Sound
	sampleCount uint64
}

// NewSilence wraps an existing sound as something that stops after a given duration.
//
// For example, to create a sound of middle C that lasts a second:
//	s := sounds.NewTimedSound(sounds.NewSineWave(261.63), 1000)
func NewTimedSound(wrapped Sound, durationMs float64) Sound {
	// NOTE: duration is Ms - time.Duration is possible, but likely more verbose.
	duration := time.Duration(int64(durationMs*1e6)) * time.Nanosecond
	sampleCount := DurationToSamples(duration)

	if wrapped.Length() < sampleCount {
		// TODO(padster) - perhaps instead pad out with timed silence?
		panic(fmt.Sprintf(
				"Can't time a sound longer than it starts out, %d < %d\n",
				wrapped.Length(), sampleCount))
	}

	data := timedSound{
		wrapped,
		sampleCount,
	}

	return NewBaseSound(&data, sampleCount)
}

// Run generates the samples by copying the wrapped sound, stopping after the set time.
func (s *timedSound) Run(base *BaseSound) {
	s.wrapped.Start()
	for at := uint64(0); at < s.sampleCount; at++ {
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
}

// String returns the textual representation
func (s *timedSound) String() string {
	ms := float64(SamplesToDuration(s.sampleCount)) / float64(time.Millisecond)
	return fmt.Sprintf("Timed[%s for %.2fms]", s.wrapped, ms)
}
