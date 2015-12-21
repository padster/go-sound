package sounds

import (
  "fmt"
  "math"
  "math/rand"

  "github.com/padster/go-sound/types"
)

// A karplusStrong is parameters to the algorithm that synthesizes a string sound
// by playing random noise at a repeated sampling rate, but averaging the samples over time.
type karplusStrong struct {
  hz              float64
  sampleOverhang  float64
  buffer         *types.Buffer
  // 1.0 = never gets quiter / flater (just repeated noise), 0.0 = immediately flat.
  sustain         float64
}

// NewKarplusStrong creates a note at a given frequency by starting with white noise
// then feeding that back into itself with a delay, which ends up sounding like a string.
// See http://music.columbia.edu/cmc/MusicAndComputers/chapter4/04_09.php
//
// For example, to create a string sound at 440hz that never gets quieter:
//  stringA := sounds.NewKarplusStrong(440.0, 1.0)
func NewKarplusStrong(hz float64, sustain float64) Sound {
  if 0.0 > sustain || sustain > 1.0 {
    panic("Sustain must be [0, 1]")
  }

  samplesPerCycle := CyclesPerSecond / hz
  bufferSize := int(math.Ceil(samplesPerCycle))

  buffer := types.NewBuffer(bufferSize)
  for i := 0; i < bufferSize; i++ {
    buffer.Push(rand.Float64() * 2.0 - 1.0)
  }

  data := karplusStrong{
    hz,
    float64(bufferSize) - samplesPerCycle,
    buffer,
    sustain,
  }
  return NewBaseSound(&data, MaxLength)
}

// Run cycles through the buffer and keep adding it to itself, linearly interpolating
// off the end to make sure to get the right cycle rate.
func (s *karplusStrong) Run(base *BaseSound) {
  lastValue := 0.0

  for at := float64(0.0); true; at += 1.0 {
    // Linterpolate off the end.
    lastIndex := s.buffer.Size() - 1
    nextValue := s.sampleOverhang * s.buffer.GetFromEnd(lastIndex) +
          (1.0 - s.sampleOverhang) * s.buffer.GetFromEnd(lastIndex -1)

    // This is the important part, smoothing the new value with the previous one.
    thisValue := s.sustain * nextValue + (1.0 - s.sustain) * lastValue
    if !base.WriteSample(thisValue) {
      break
    }
    s.buffer.Push(thisValue)
    lastValue = thisValue
  }
}

// Stop cleans up the sound by stopping the underlying sound.
func (s *karplusStrong) Stop() {
  // No-op
}

// Reset clears the buffer back to a white-noise state.
func (s *karplusStrong) Reset() {
  for i := 0; i < s.buffer.Size(); i++ {
    s.buffer.Push(rand.Float64() * 2.0 - 1.0)
  }
}

// String returns the textual representation
func (s *karplusStrong) String() string {
  return fmt.Sprintf("KarplusStrong[%.2fhz]", s.hz)
}
