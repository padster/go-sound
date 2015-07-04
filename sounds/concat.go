// Runs multiple non-infinite sounds, one after the other.
package sounds

import (
	"math"
)

type Concat struct {
	samples chan float64
	wrapped []Sound

	durationMs uint64
	indexAt    int
	running    bool
}

func ConcatSounds(wrapped ...Sound) *Concat {
	durationMs := uint64(0)
	for _, child := range wrapped {
		wrappedLength := child.DurationMs()
		if durationMs+wrappedLength < wrappedLength { // Overflow, so cap out at max.
			durationMs = math.MaxUint64
			break
		} else {
			durationMs += wrappedLength
		}
	}

	ret := Concat{
		make(chan float64),
		wrapped,
		durationMs,
		0,     /* indexAt */
		false, /* running */
	}
	return &ret
}

func (s *Concat) GetSamples() <-chan float64 {
	return s.samples
}

func (s *Concat) DurationMs() uint64 {
	return s.durationMs
}

func (s *Concat) Start() {
	s.running = true

	if len(s.wrapped) > 0 {
		go func() {
			for s.running && s.indexAt < len(s.wrapped) {
				s.wrapped[s.indexAt].Start()
				samples := s.wrapped[s.indexAt].GetSamples()
				for sample := range samples {
					if !s.running {
						break
					}
					s.samples <- sample
				}
				s.wrapped[s.indexAt].Stop()
				s.indexAt++
			}

			if s.indexAt < len(s.wrapped) {
				s.wrapped[s.indexAt].Stop()
			}
			s.Stop()
			close(s.samples)
		}()
	}
}

func (s *Concat) Stop() {
	s.running = false
}

func (s *Concat) Reset() {
	if s.running {
		panic("Stop before reset!")
	}

	s.samples = make(chan float64)
	for _, wrapped := range s.wrapped {
		wrapped.Stop()
		wrapped.Reset()
	}
	s.indexAt = 0
	s.running = true
}
