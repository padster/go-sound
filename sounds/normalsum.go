package sounds

import (
	"fmt"
)

// A normalSum is parameters to the algorithm that adds together sounds in parallel,
// normalized to avoid going outside [-1, 1]
type normalSum struct {
	wrapped    []Sound
	normScalar float64
}

// SumSounds creates a sound by adding multiple sounds in parallel, playing them
// at the same time and normalizing their volume.
//
// For example, to play a G7 chord for a second:
//	s := sounds.SumSounds(
//		sounds.NewTimedSound(sounds.MidiToSound(55), 1000),
//		sounds.NewTimedSound(sounds.MidiToSound(59), 1000),
//		sounds.NewTimedSound(sounds.MidiToSound(62), 1000),
//		sounds.NewTimedSound(sounds.MidiToSound(65), 1000),
//		sounds.NewTimedSound(sounds.MidiToSound(67), 1000),
//	)
func SumSounds(wrapped ...Sound) Sound {
	if len(wrapped) == 0 {
		panic("SumSounds can't take no sounds")
	}

	sampleCount := MaxLength
	for _, child := range wrapped {
		childLength := child.Length()
		if childLength < sampleCount {
			sampleCount = childLength
		}
	}

	data := normalSum{
		wrapped,
		1.0 / float64(len(wrapped)), /* normScalar */
	}

	return NewBaseSound(&data, sampleCount)
}

// Run generates the samples by summing all the wrapped samples and normalizing.
func (s *normalSum) Run(base *BaseSound) {
	for _, wrapped := range s.wrapped {
		wrapped.Start()
	}

	for {
		sum := 0.0
		for _, wrapped := range s.wrapped {
			sample, stream_ok := <-wrapped.GetSamples()
			if !stream_ok || !base.Running() {
				base.Stop()
				break
			}
			sum += sample
		}

		if !base.WriteSample(sum * s.normScalar) {
			break
		}
	}
}

// Stop cleans up the sound by stopping all underlyings sound.
func (s *normalSum) Stop() {
	for _, wrapped := range s.wrapped {
		wrapped.Stop()
	}
}

// Reset resets all underlying sounds.
func (s *normalSum) Reset() {
	for _, wrapped := range s.wrapped {
		wrapped.Reset()
	}
}

// String returns the textual representation
func (s *normalSum) String() string {
	result := "Sum["
	for i, wrapped := range s.wrapped {
		if i > 0 {
			result += " & "
		}
		result += fmt.Sprintf("%s", wrapped)
	}
	return result + "]"
}
