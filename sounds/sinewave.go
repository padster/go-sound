// Sound implementation that is a single pure sine wave.
package sounds

import (
	"math"
)

type SineWave struct {
	samples chan float64
	hz      float64

	timeAt  float64
	running bool
}

// NewSineWave creates a new sound at a given pitch.
func NewSineWave(hz float64) *SineWave {
	ret := SineWave{
		make(chan float64),
		hz,
		0,     /* timeAt */
		false, /* running */
	}
	return &ret
}

func (s *SineWave) GetSamples() <-chan float64 {
	return s.samples
}

func (s *SineWave) DurationMs() uint64 {
	return math.MaxUint64
}

func (s *SineWave) Start() {
	s.running = true
	timeDelta := s.hz * 2.0 * math.Pi * SecondsPerCycle()

	go func() {
		for s.running {
			s.samples <- math.Sin(s.timeAt)
			s.timeAt += timeDelta
		}
		close(s.samples)
	}()
}

func (s *SineWave) Stop() {
	s.running = false
}

func (s *SineWave) Reset() {
	if s.running {
		panic("Stop before reset!")
	}

	s.timeAt = 0
	s.samples = make(chan float64)
	s.running = true
}
