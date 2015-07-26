// Package sounds provides the basic types for Sounds within this system,
// plus multiple implementations of different sounds that can be used.
package sounds

import (
	"time"
)

const (
	// The sample rate of each sound stream.
	CyclesPerSecond = 44100.0

	// Inverse sample rate.
	SecondsPerCycle = 1.0 / CyclesPerSecond

	// The number of samples in the maximum duration.
	MaxLength = uint64(406750706825295)
	// HACK - Go doesn't allow uint64(float64(math.MaxInt64) * 0.000000001 * CyclesPerSecond) :(

	// Maximum duration, used for unending sounds.
	MaxDuration = time.Duration(int64(float64(MaxLength)*SecondsPerCycle*1e9)) * time.Nanosecond
)

/**
 * Upcoming:
 *   - proper README
 *   - fix static audio in wav output
 *   - golang cleanup - gofmt, toString, fix exported set, final variables, generate godoc, pointer vs. object to baseSound?
 *   - push to github, make public and announce
 *   - sound forker
 *   - implement some of these: https://www.youtube.com/channel/UCchjpg1aaY91WubqAYRcNsg
 *   - implement instrument synthesizers
 *   - midi / wave input
 *   - reverb: https://christianfloisand.wordpress.com/2012/09/04/digital-reverberation
 */
type Sound interface {
	// Sound wave samples for the sound - only valid after Start() and before Stop()
	GetSamples() <-chan float64

	// Number of samples in this sound - MaxLength of unlimited.
	Length() uint64

	// Length of time this goes for. Convenience method, should always be SamplesToDuration(Length())
	Duration() time.Duration

	// Start begins writing the sound wave to the samples channel.
	Start()

	// Running indicates whether a sound has Start()-ed but not yet Stop()-ed
	Running() bool

	// Stop ceases writing samples, and closes the channel.
	Stop()

	// Reset converts the sound back to the pre-Start() state.
	Reset()
}

// SamplesToDuration converts a sample count to a duration of time.
func SamplesToDuration(sampleCount uint64) time.Duration {
	return time.Duration(int64(float64(sampleCount)*1e9*SecondsPerCycle)) * time.Nanosecond
}

func DurationToSamples(duration time.Duration) uint64 {
	return uint64(float64(duration.Nanoseconds()) * 1e-9 * CyclesPerSecond)
}
