// API for the Sound type.
package sounds

/**
 * Upcoming:
 *   - write examples for each type, and one overall demo with tune using multiple
 *   - remove TODOs and PICKs
 *   - add documentation
 *   - proper README
 *   - fix static in wav output
 *   - golang cleanup - gofmt, toString, fix exported set, final variables?
 *   - push to github, make public and announce
 *   - sound forker
 *   - implement some of these: https://www.youtube.com/channel/UCchjpg1aaY91WubqAYRcNsg
 *   - implement instrument synthesizers
 *   - midi / wave input
 *   - reverb: https://christianfloisand.wordpress.com/2012/09/04/digital-reverberation
 */
type Sound interface {
	/* Sound wave samples for the sound */
	GetSamples() <-chan float64

	/* How many milliseconds this sound goes for, or math.MaxUint64 if 'infinite'. */
	DurationMs() uint64 // TODO: use time.Duration

	/* Being writing the soundwave to the samples channel. */
	Start()

	/* Stop writing samples, and close the channel. */
	Stop()

	/* Reset back to the pre-start state. */
	Reset()

	// PICK - Consider adding Pause()
}

// Global constant for the sample rate of each sound stream.
func CyclesPerSecond() float64 {
	return 44100.0
}

// Global constant for the inverse sample rate.
func SecondsPerCycle() float64 {
	return 1.0 / CyclesPerSecond()
}
