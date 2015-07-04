// Runs a single sound multiple times.
package sounds

import (
	"fmt"
	"math"
)

type Repeater struct {
	samples chan float64
	wrapped Sound

	loops      int
	loopAt     int
	durationMs uint64
	running    bool
}

func RepeatSound(wrapped Sound, loops int) *Repeater {
	// TODO - support -1 == infinite loop
	durationMs := wrapped.DurationMs()
	if durationMs != math.MaxUint64 {
		durationMs *= uint64(loops)
	}

	ret := Repeater{
		make(chan float64),
		wrapped,
		loops,
		0, /* loopAt */
		durationMs,
		false, /* running */
	}
	return &ret
}

func (s *Repeater) GetSamples() <-chan float64 {
	return s.samples
}

func (s *Repeater) DurationMs() uint64 {
	return s.durationMs
}

func (s *Repeater) Start() {
	s.running = true

	go func() {
		for s.running && s.loopAt < s.loops {
			fmt.Printf("Loop %v of %v\n", s.loopAt, s.loops)
			s.wrapped.Start()
			samples := s.wrapped.GetSamples()
			for sample := range samples {
				if !s.running {
					break
				}
				s.samples <- sample
			}
			s.wrapped.Stop()
			s.wrapped.Reset()
			s.loopAt++
		}

		s.wrapped.Stop()
		s.Stop()
		close(s.samples)
	}()
}

func (s *Repeater) Stop() {
	s.wrapped.Stop()
	s.running = false
}

func (s *Repeater) Reset() {
	if s.running {
		panic("Stop before reset!")
	}

	s.samples = make(chan float64)
	s.loopAt = 0
	s.wrapped.Reset()
	s.running = true
}
