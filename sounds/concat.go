package sounds

import (
	"fmt"
)

// A concat is parameters to the algorithm that concatenates multiple sounds
// one after the other, to allow playing them in series.
type concat struct {
	wrapped []Sound
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
	sampleCount := uint64(0)
	for _, child := range wrapped {
		childLength := child.Length()
		if sampleCount+childLength < childLength { // Overflow, so cap out at max.
			childLength = MaxLength
			break
		} else {
			sampleCount += childLength
		}
	}

	data := concat{
		wrapped,
	}

	return NewBaseSound(&data, sampleCount)
}

// Run generates the samples by copying each wrapped sound in turn.
func (s *concat) Run(base *BaseSound) {
	cease := false

	// TODO(padster): The trivial implementation leads to bad sounds at the changeover points.
	// The sounds should be merged together more cleanly to avoid this.
	for _, wrapped := range s.wrapped {
		wrapped.Start()
		for sample := range wrapped.GetSamples() {
			if !base.WriteSample(sample) {
				cease = true
				break
			}
		}
		wrapped.Stop()
		if cease {
			break
		}
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
}

// String returns the textual representation
func (s *concat) String() string {
	result := "Concat["
	for i, wrapped := range s.wrapped {
		if i > 0 {
			result += " + "
		}
		result += fmt.Sprintf("%s", wrapped)
	}
	return result + "]"
}
