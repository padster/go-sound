package sounds

import (
	"fmt"
	"time"

	"github.com/padster/go-sound/types"
)

// A delay is parameters to the algorithm that adds a sound to a delayed version of itself.
type delay struct {
	wrapped      Sound
	delaySamples uint64
	buffer       *types.Buffer
}

// AddDelay takes a sound, and adds it with a delayed version of itself after a given duration.
//
// For example, to have a three note progression with a delay of 123ms:
//  s.AddDelay(s.ConcatSounds(
//    s.NewTimedSound(u.MidiToSound(55), 678),
//    s.NewTimedSound(u.MidiToSound(59), 678),
//    s.NewTimedSound(u.MidiToSound(62), 678),
//  ), 123)
func AddDelay(wrapped Sound, delayMs float64) Sound {
	delayDuration := time.Duration(int64(delayMs*1e6)) * time.Nanosecond
	delaySamples := DurationToSamples(delayDuration)

	data := delay{
		wrapped,
		delaySamples,
		types.NewBuffer(int(delaySamples)),
	}

	return NewBaseSound(&data, wrapped.Length())
}

// Run generates the samples by adding the wrapped samples to a delayed version of the channel.
func (s *delay) Run(base *BaseSound) {
	s.wrapped.Start()
	for sample := range s.wrapped.GetSamples() {
		// Add to buffer, and read the delayed version.
		delayed := s.buffer.Push(sample)

		value := (sample + delayed) * 0.5
		if !base.WriteSample(value) {
			break
		}
	}
}

// Stop cleans up the sound by stopping the underlying sound.
func (s *delay) Stop() {
	s.wrapped.Stop()
}

// Reset resets the underlying sound, and clears out the buffer state.
func (s *delay) Reset() {
	s.wrapped.Reset()
	s.buffer.Clear()
}

// String returns the textual representation
func (s *delay) String() string {
	ms := float64(SamplesToDuration(s.delaySamples)) / float64(time.Millisecond)
	return fmt.Sprintf("Delay[%s with delay %.2fms]", s.wrapped, ms)
}
