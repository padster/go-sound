// Runs a particular sound for a set amount of time.
package sounds

import (
	// "fmt"
)

type TimedSound struct {
	samples chan float64
	wrapped Sound

	durationMs   uint64
	durationLeft float64
	running      bool
}

func NewTimedSound(wrapped Sound, durationMs float64) *TimedSound {
	ret := TimedSound{
		make(chan float64),
		wrapped,
		uint64(durationMs),
		durationMs, /* durationLeft */
		false,      /* running */
	}
	return &ret
}

func (s *TimedSound) GetSamples() <-chan float64 {
	return s.samples
}

func (s *TimedSound) DurationMs() uint64 {
	return s.durationMs
}

func (s *TimedSound) Start() {
	s.running = true
	s.wrapped.Start()
	go func() {
		for s.running && s.durationLeft > 0.0 {
			next := <-s.wrapped.GetSamples()
			s.samples <- next
			s.durationLeft -= SecondsPerCycle() * 1000.0
		}

		// TODO: BIG HACK?
		if s.running {
			s.running = false
			s.wrapped.Stop()
			close(s.samples)
		}
	}()
}

func (s *TimedSound) Stop() {
	if s.running {
		s.running = false
		s.wrapped.Stop()
	}
}

func (s *TimedSound) Reset() {
	s.durationLeft = float64(s.durationMs)
	s.running = true
	s.wrapped.Reset()
	s.samples = make(chan float64) // HACK - need to recreate, otherwise it may already be closed.
}
