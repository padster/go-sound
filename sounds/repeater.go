package sounds

import (
	"math"
)

// A repeater is parameters to the algorithm that repeats a sound a given number of times.
type repeater struct {
	wrapped   Sound
	loopCount int32
}

// RepeatSound forms a sound by repeating a given sound a number of times in series.
//
// For example, for the cello part of Pachelbel's Canon in D:
//	sound := s.RepeatSound(s.ConcatSounds(
//		s.NewTimedSound(s.MidiToSound(50), 800),
//		s.NewTimedSound(s.MidiToSound(45), 800),
//		s.NewTimedSound(s.MidiToSound(47), 800),
//		s.NewTimedSound(s.MidiToSound(42), 800),
//		s.NewTimedSound(s.MidiToSound(43), 800),
//		s.NewTimedSound(s.MidiToSound(38), 800),
//		s.NewTimedSound(s.MidiToSound(43), 800),
//		s.NewTimedSound(s.MidiToSound(45), 800),
//	), -1 /* repeat indefinitely */)
func RepeatSound(wrapped Sound, loopCount int32) Sound {
	// Negative loop count == loop indefinitely
	if loopCount < 0 {
		loopCount = math.MaxInt32
	}

	sampleCount := wrapped.Length()
	if sampleCount != MaxLength {
		sampleCount *= uint64(loopCount)
	}

	data := repeater{
		wrapped,
		loopCount,
	}
	return NewBaseSound(&data, sampleCount)
}

// Run generates the samples by copying from the wrapped sound multiple times.
func (s *repeater) Run(base *BaseSound) {
	cease := false

	for loopAt := int32(0); !cease && loopAt < s.loopCount; loopAt++ {
		s.wrapped.Start()

		for sample := range s.wrapped.GetSamples() {
			if !base.WriteSample(sample) {
				cease = true
			}
		}

		s.wrapped.Stop()
		if !cease {
			s.wrapped.Reset()
		}
	}
}

// Stop cleans up the sound by stopping the underlying sound.
func (s *repeater) Stop() {
	s.wrapped.Stop()
}

// Reset resets the underlying sound, plus the loop count tracking.
func (s *repeater) Reset() {
	s.wrapped.Reset()
}
