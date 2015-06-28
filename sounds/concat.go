// Runs multiple non-infinite sounds, one after the other.
package sounds

import (
	"fmt"
)

type Concat struct {
	samples chan float64
	wrapped []Sound

	indexAt int
	running bool
}

func ConcatSounds(wrapped ...Sound) *Concat {
	ret := Concat{
		make(chan float64),
		wrapped,
		0,     /* indexAt */
		false, /* running */
	}
	return &ret
}

func (s *Concat) GetSamples() <-chan float64 {
	return s.samples
}

func (s *Concat) Start() {
	s.running = true

	if len(s.wrapped) > 0 {
		go func() {
			for s.running && s.indexAt < len(s.wrapped) {
				fmt.Printf("Starting concat sound at %v\n", s.indexAt)
				s.wrapped[s.indexAt].Start()
				samples := s.wrapped[s.indexAt].GetSamples()
				for sample := range samples {
					if !s.running {
						break
					}
					s.samples <- sample
				}
				fmt.Printf("Finished concat sound at %v\n", s.indexAt)
				s.wrapped[s.indexAt].Stop()
				s.indexAt++
			}
			fmt.Printf("CONCAT FINISHED, closing samples\n")
			s.running = false
			close(s.samples)
		}()
	}
}

func (s *Concat) Stop() {
	if s.running {
		s.running = false
		s.wrapped[s.indexAt].Stop()
		s.indexAt = 0
		close(s.samples)
	}
}

// TODO - implement properly (properly handle immediate changes while running)
func (s *Concat) Reset() {
	s.running = true
	s.wrapped[s.indexAt].Stop()
	s.indexAt = 0
}
