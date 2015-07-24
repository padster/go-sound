package sounds

import (
	"math"
)

// A concat is parameters to the algorithm that concatenates multiple sounds
// one after the other, to allow playing them in series.
type concat struct {
	wrapped []Sound

	indexAt int
}

// ConcatSounds creates a sound by concatenating multiple sounds in series.
//
// For example, to create the 5-note sequence from Close Enounters:
//	s := sounds.ConcatSounds(
//		sounds.NewTimedSound(sounds.MidiToSound(74), 400),
//		sounds.NewTimedSound(sounds.MidiToSound(76), 400),
//		sounds.NewTimedSound(sounds.MidiToSound(72), 400),
//		sounds.NewTimedSound(sounds.MidiToSound(60), 400),
//		sounds.NewTimedSound(sounds.MidiToSound(67), 1200),
//	)
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

	data := concat{
		wrapped,
		0, /* indexAt */
	}

	return NewBaseSound(&data, durationMs)
}

// Run generates the samples by copying each wrapped sound in turn.
func (s *concat) Run(base *BaseSound) {
	cease := false

	for !cease && s.indexAt < len(s.wrapped) {
		s.wrapped[s.indexAt].Start()
		// TODO - merge with range statement?
		samples := s.wrapped[s.indexAt].GetSamples()

		for sample := range samples {
			if !base.WriteSample(sample) {
				cease = true
				break
			}
		}
		s.wrapped[s.indexAt].Stop()
		s.indexAt++
	}
}

// Stop cleans up the sound by stopping all underlying sounds.
func (s *concat) Stop() {
	for _, wrapped := range s.wrapped {
		wrapped.Stop()
	}
}

// Reset resets the all underlying sounds, and lines up playing the first sound next.
func (s *concat) Reset() {
	for _, wrapped := range s.wrapped {
		wrapped.Reset()
	}
	s.indexAt = 0
}
