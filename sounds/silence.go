// The sound of silence.
package sounds

import (
  "math"
)

type Silence struct {

}

func NewSilence() Sound {
  silence := Silence{}
  return NewBaseSound(&silence, math.MaxUint64)
}

func (s *Silence) Run(base *BaseSound) {
  for base.WriteSample(0) {
    // No-op
  }
}

func (s *Silence) Stop() {
  // No-op
}

func (s *Silence) Reset() {
  // No-op
}

func NewTimedSilence(durationMs float64) Sound {
  return NewTimedSound(NewSilence(), durationMs)
}
