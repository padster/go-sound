// Runs multiple non-infinite sounds, one after the other.
package sounds

import (
	"math"
)

type Concat struct {
	wrapped []Sound

	indexAt int
}

func ConcatSounds(wrapped ...Sound) Sound {
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

	concat := Concat{
		wrapped,
		0, /* indexAt */
	}
	return NewBaseSound(&concat, durationMs)
}

func (s *Concat) Run(base *BaseSound) {
	for s.indexAt < len(s.wrapped) {
		s.wrapped[s.indexAt].Start()
		// TODO - merge with range statement?
		samples := s.wrapped[s.indexAt].GetSamples()

		cease := false
		for sample := range samples {
			if !base.WriteSample(sample) {
				cease = true
				break
			}
		}
		s.wrapped[s.indexAt].Stop()
		s.indexAt++

		if cease {
			break
		}
	}
}

func (s *Concat) Stop() {
	// No-op, handled inside Run
}

func (s *Concat) Reset() {
	for _, wrapped := range s.wrapped {
		// TODO - needed? wrapped.Stop()
		wrapped.Reset()
	}
	s.indexAt = 0
}
