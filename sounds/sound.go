// API for the Sound type.
package sounds

/**
 * Upcoming:
 *   - get renderer working again
 *   - repeater
 *   - add chords
 *   - write recognizeable tune in demo
 *   - clean up README documentation for running
 *   - remove TODOs
 *   - add documentation
 *   - cleanup and push.
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

	Start()
	Stop()

	// TODO - Does reset also restart? Have separate, or rename?
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
