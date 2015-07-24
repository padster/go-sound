package sounds

import (
	"math"
)

// A sineWave is parameters to the algorithm that generates a sound wave at a given pitch.
type sineWave struct {
	hz        float64
	timeDelta float64

	timeAt float64
}

// NewSineWave creates an unending sound at a given pitch (in hz).
//
// For example, to create a sound represeting A440:
//	s := sounds.NewSineWave(440)
func NewSineWave(hz float64) Sound {
	timeDelta := (2.0 * math.Pi) * (hz * SecondsPerCycle())

	data := sineWave{
		hz,
		timeDelta,
		0, /* timeAt */
	}

	return NewBaseSound(&data, math.MaxUint64)
}

// Run generates the samples by creating a sine wave at the desired frequency.
func (s *sineWave) Run(base *BaseSound) {
	for {
		if !base.WriteSample(math.Sin(s.timeAt)) {
			return
		}
		// NOTE: Will overflow if run too long.
		s.timeAt += s.timeDelta
	}
}

// Stop cleans up the sound, in this case doing nothing.
func (s *sineWave) Stop() {
	// No-op
}

// Reset sets the offset in the wavelength back to zero.
func (s *sineWave) Reset() {
	s.timeAt = 0
}
