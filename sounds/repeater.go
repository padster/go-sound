package sounds

// A repeater is parameters to the algorithm that repeats a sound a given number of times.
type repeater struct {
	wrapped   Sound
	loopCount int

	loopAt int
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
func RepeatSound(wrapped Sound, loopCount int) Sound {
	if loopCount < 0 {
		// TODO - support -1 == infinite loop
		panic("Can't have negative loop count for repeat")
	}

	sampleCount := wrapped.Length()
	if sampleCount != MaxLength {
		sampleCount *= uint64(loopCount)
	}

	data := repeater{
		wrapped,
		loopCount,
		0, /* loopAt */
	}
	return NewBaseSound(&data, sampleCount)
}

// Run generates the samples by copying from the wrapped sound multiple times.
func (s *repeater) Run(base *BaseSound) {
	cease := false

	for !cease && s.loopAt < s.loopCount {
		if s.loopAt > 0 {
			s.wrapped.Reset()
		}
		s.wrapped.Start()

		// TODO - merge with range statement?
		samples := s.wrapped.GetSamples()
		for sample := range samples {
			if !base.WriteSample(sample) {
				cease = true
			}
		}

		s.wrapped.Stop()
		s.loopAt++
	}
}

// Stop cleans up the sound by stopping the underlying sound.
func (s *repeater) Stop() {
	s.wrapped.Stop()
}

// Reset resets the underlying sound, plus the loop count tracking.
func (s *repeater) Reset() {
	s.wrapped.Reset()
	s.loopAt = 0
}
