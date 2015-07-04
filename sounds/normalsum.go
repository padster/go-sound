// Adds sounds together in parallel, and normalizes to (-1, 1) by dividing by the input count.
package sounds

import (
	"math"
)

type NormalSum struct {
	samples chan float64
	wrapped []Sound

	durationMs uint64
	running    bool
}

func SumSounds(wrapped ...Sound) *NormalSum {
	if len(wrapped) == 0 {
		panic("NormalSum can't take no sounds")
	}

	var durationMs uint64 = math.MaxUint64
	for _, child := range wrapped {
		childDurationMs := child.DurationMs()
		if childDurationMs < durationMs {
			durationMs = childDurationMs
		}
	}

	ret := NormalSum{
		make(chan float64),
		wrapped,
		durationMs,
		false, /* running */
	}
	return &ret
}

func (s *NormalSum) GetSamples() <-chan float64 {
	return s.samples
}

func (s *NormalSum) DurationMs() uint64 {
	return s.durationMs
}

func (s *NormalSum) Start() {
	normScalar := 1.0 / float64(len(s.wrapped))

	s.running = true
	for _, wrapped := range s.wrapped {
		wrapped.Start()
	}

	go func() {
		for s.running {
			sum := 0.0
			for _, wrapped := range s.wrapped {
				sample, stream_ok := <-wrapped.GetSamples()
				if !stream_ok || !s.running {
					s.running = false
					break
				}
				sum += sample
			}

			if s.running {
				s.samples <- sum * normScalar
			}
		}

		s.Stop()
		close(s.samples)
	}()
}

func (s *NormalSum) Stop() {
	s.running = false
	for _, wrapped := range s.wrapped {
		wrapped.Stop()
	}
}

func (s *NormalSum) Reset() {
	if s.running {
		panic("Stop before reset!")
	}

	s.samples = make(chan float64)
	for _, wrapped := range s.wrapped {
		wrapped.Reset()
	}
	s.running = true
}
