// Runs a particular sound for a set amount of time.
package sounds

import "fmt"

type TimedSound struct {
	wrapped Sound

	durationMs   uint64
	durationLeft float64
}

func NewTimedSound(wrapped Sound, durationMs float64) Sound {
	timedSound := TimedSound{
		wrapped,
		uint64(durationMs),
		durationMs, /* durationLeft */
	}

	return NewBaseSound(&timedSound, timedSound.durationMs)
}

func (s *TimedSound) Run(base *BaseSound) {
	s.wrapped.Start()
	for s.durationLeft > 0.0 {
		if !base.WriteSample(<-s.wrapped.GetSamples()) {
			break
		}
		s.durationLeft -= SecondsPerCycle() * 1000.0
	}
}

func (s *TimedSound) Stop() {
	fmt.Println("Stop timed, about to stop child")
	s.wrapped.Stop()
	fmt.Println("Stopped child of timed! = sum?")
}

func (s *TimedSound) Reset() {
	fmt.Println("Reset timed")
	s.wrapped.Reset()
	s.durationLeft = float64(s.durationMs)
}
