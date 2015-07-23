// Sound implementation that is a single pure sine wave.
package sounds

import (
	"math"
)

type SineWave struct {
	hz        float64
	timeDelta float64

	timeAt float64
}

// NewSineWave creates a new sound at a given pitch.
func NewSineWave(hz float64) Sound {
	timeDelta := hz * 2.0 * math.Pi * SecondsPerCycle()
	sineWave := SineWave{
		hz,
		timeDelta,
		0, /* timeAt */
	}

	return NewBaseSound(&sineWave, math.MaxUint64)
}

func (s *SineWave) Run(base *BaseSound) {
	for {
		if !base.WriteSample(math.Sin(s.timeAt)) {
			return
		}
		s.timeAt += s.timeDelta
	}
}

func (s *SineWave) Stop() {
	// No-op
}

func (s *SineWave) Reset() {
	s.timeAt = 0
}
