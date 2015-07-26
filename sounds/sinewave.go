package sounds

import (
	"fmt"
	"math"
)

// A sineWave is parameters to the algorithm that generates a sound wave at a given pitch.
type sineWave struct {
	hz        float64
	timeDelta float64
}

// NewSineWave creates an unending sound at a given pitch (in hz).
//
// For example, to create a sound represeting A440:
//	s := sounds.NewSineWave(440)
func NewSineWave(hz float64) Sound {
	timeDelta := (2.0 * math.Pi) * (hz * SecondsPerCycle)

	data := sineWave{
		hz,
		timeDelta,
	}

	return NewBaseSound(&data, MaxLength)
}

// Run generates the samples by creating a sine wave at the desired frequency.
func (s *sineWave) Run(base *BaseSound) {
	// NOTE: Will overflow if run too long.
	for timeAt := float64(0); true; timeAt += s.timeDelta {
		if !base.WriteSample(math.Sin(timeAt)) {
			return
		}
	}
}

// Stop cleans up the sound, in this case doing nothing.
func (s *sineWave) Stop() {
	// No-op
}

// Reset sets the offset in the wavelength back to zero.
func (s *sineWave) Reset() {
	// No-op
}

// String returns the textual representation
func (s *sineWave) String() string {
	return fmt.Sprintf("Hz[%.2f]", s.hz)
}
