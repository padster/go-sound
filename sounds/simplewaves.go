package sounds

import (
	"fmt"
	"math"
)

type SimpleSampleMap func(float64) float64

// A simpleWave is parameters to the algorithm that generates a sound wave by cycling a particular periodic
// shape at a given frequency.
type simpleWave struct {
	hz        float64
	timeDelta float64
	mapper    SimpleSampleMap
}


// NewSimpleWave creates an unending repeating sound based on cycles defined by a given mapping function.
// For examples of usage, see sine/square/sawtooth/triangle waves below.
func NewSimpleWave(hz float64, mapper SimpleSampleMap) Sound {
	return NewBaseSound(&simpleWave{hz, hz * SecondsPerCycle, mapper}, MaxLength)
}

// NewSineWave creates an unending sinusoid at a given pitch (in hz).
//
// For example, to create a sound represeting A440:
//	s := sounds.NewSineWave(440)
func NewSineWave(hz float64) Sound {
	return NewSimpleWave(hz, SineMap)
}

// NewSquareWave creates an unending [-1, 1] square wave at a given pitch. 
func NewSquareWave(hz float64) Sound {
	return NewSimpleWave(hz, SquareMap)
}

// NewSawtoothWave creates an unending sawtooth pattern (-1->1 then resets to -1.)
func NewSawtoothWave(hz float64) Sound {
	return NewSimpleWave(hz, SawtoothMap)
}

// NewTriangleWave creates an unending triangle pattern (-1->1->-1 linearly)
func NewTriangleWave(hz float64) Sound {
	return NewSimpleWave(hz, TriangleMap)
}

// Run generates the samples by creating a sine wave at the desired frequency.
func (s *simpleWave) Run(base *BaseSound) {
	// NOTE: Will overflow if run too long.
	for timeAt := float64(0); true; _, timeAt = math.Modf(timeAt + s.timeDelta) {
		if !base.WriteSample(s.mapper(timeAt)) {
			return
		}
	}
}

// Stop cleans up the sound, in this case doing nothing.
func (s *simpleWave) Stop() {
	// No-op
}

// Reset sets the offset in the wavelength back to zero.
func (s *simpleWave) Reset() {
	// No-op
}

// String returns the textual representation
func (s *simpleWave) String() string {
	return fmt.Sprintf("Hz[%.2f]", s.hz)
}

// Below are some sample mappers used for generating various useful shapes.
func SineMap(at float64) float64 {
	return math.Sin(at * 2.0 * math.Pi)
}
func SquareMap(at float64) float64 {
	if at < 0.5 {
		return -1.0
	} else {
		return 1.0
	}
}
func SawtoothMap(at float64) float64 {
	return at * 2.0 - 1.0
}
func TriangleMap(at float64) float64 {
	if at > 0.5 {
		at = 1.0 - at
	}
	return at * 2.0 * 2.0 - 1.0
}