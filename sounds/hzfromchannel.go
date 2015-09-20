package sounds

import (
  "math"
)

// A hzFromChannel is parameters to the algorithm that generates a variable tone.
type hzFromChannel struct {
  wrapped      <-chan float64
  wrappedWithAmplitute <-chan []float64
}

// NewHzFromChannel takes stream of hz values, and generates a tone that sounds
// like those values over time. For a fixed tone, see NewSineWave.
func NewHzFromChannel(wrapped <-chan float64) Sound {
  return NewBaseSound(&hzFromChannel{
    wrapped,
    nil,
  }, MaxLength)
}

func NewHzFromChannelWithAmplitude(wrappedWithAmplitute <-chan []float64) Sound {
  return NewBaseSound(&hzFromChannel{
    nil,
    wrappedWithAmplitute,
  }, MaxLength)
}

// Run generates the samples by adding the wrapped samples to a delayed version of the channel.
func (s *hzFromChannel) Run(base *BaseSound) {
  timeAt := 0.0
  TAU := 2.0 * math.Pi

  if s.wrapped != nil {
    for currentHz := range s.wrapped {
      timeDelta := TAU * (currentHz * SecondsPerCycle)
      timeAt = math.Mod(timeAt + timeDelta, TAU)
      if !base.WriteSample(math.Sin(timeAt)) {
        return
      }
    }
  } else {
    for hzAndAmplitude := range s.wrappedWithAmplitute {
      currentHz := hzAndAmplitude[0]
      amplitude := hzAndAmplitude[1]
      timeDelta := TAU * (currentHz * SecondsPerCycle)
      timeAt = math.Mod(timeAt + timeDelta, TAU)
      if !base.WriteSample(amplitude * math.Sin(timeAt)) {
        return
      }
    }
  }
}

// Stop cleans up the sound by stopping the underlying sound.
func (s *hzFromChannel) Stop() {
  // NO-OP
}

// Reset resets the underlying sound, and clears out the buffer state.
func (s *hzFromChannel) Reset() {
  panic("Can't reset a stream-based sound.")
}

// String returns the textual representation
func (s *hzFromChannel) String() string {
  return "HzFromChannel"
}
