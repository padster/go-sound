// Runs a single sound multiple times.
package sounds

import (
	"fmt"
	"math"
)

type Repeater struct {
	wrapped Sound

	loops  int
	loopAt int
}

func RepeatSound(wrapped Sound, loops int) Sound {
	// TODO - support -1 == infinite loop
	durationMs := wrapped.DurationMs()
	if durationMs != math.MaxUint64 {
		durationMs *= uint64(loops)
	}

	repeater := Repeater{
		wrapped,
		loops,
		0, /* loopAt */
	}
	return NewBaseSound(&repeater, durationMs)
}

func (s *Repeater) Run(base *BaseSound) {
RepeatLoop:
	for s.loopAt < s.loops {
		s.wrapped.Start()
		// TODO - merge with range statement?
		samples := s.wrapped.GetSamples()

		for sample := range samples {
			if !base.WriteSample(sample) {
				break RepeatLoop
			}
		}

		// TODO - figure out semantics of stopping wrapped children here vs. in base sound.
		s.wrapped.Stop()
		s.wrapped.Reset()
		s.loopAt++
	}
}

func (s *Repeater) Stop() {
	fmt.Println("Stop repeat")
	s.wrapped.Stop()
}

func (s *Repeater) Reset() {
	fmt.Println("Reset repeat")
	s.wrapped.Reset()
	s.loopAt = 0
}
