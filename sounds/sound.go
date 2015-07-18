// API for the Sound type.
package sounds

/**
 * Upcoming:
 *   - sound forker
 *   - remove TODOs
 *   - move screen renderer into this repo.
 *   - simplified API to manage lifecycle.
 *   - write examples for each type, and one overall demo with tune using multiple
 *   - add documentation
 *   - proper README
 *   - golang cleanup
 *   - push to github, make public and announce
 *   - implement some of these: https://www.youtube.com/channel/UCchjpg1aaY91WubqAYRcNsg
 *   - implement instrument synthesizers
 *   - midi / wave input
 *   - reverb: https://christianfloisand.wordpress.com/2012/09/04/digital-reverberation
 */
type Sound interface {
	/* Sound wave samples for the sound */
	GetSamples() <-chan float64

	/* How many milliseconds this sound goes for, or math.MaxUint64 if 'infinite'. */
	DurationMs() uint64 // PICK: use time.Duration?

	/* Being writing the soundwave to the samples channel. */
	Start()

	/* Stop writing samples, and close the channel. */
	Stop()

	/* Reset back to the pre-start state. */
	Reset()

	// TODO - Consider adding Pause()
}

// Global constant for the sample rate of each sound stream.
func CyclesPerSecond() float64 {
	return 44100.0
}

// Global constant for the inverse sample rate.
func SecondsPerCycle() float64 {
	return 1.0 / CyclesPerSecond()
}
