// API for the Sound type.
package sounds

/**
 * Upcoming:
 *   - add chords
 *   - add attack function
 *   - write recognizeable tune in demo
 *   - write README documentation for running
 *   - migrate pulse to local file?
 *   - remove TODOs
 *   - add documentation
 *   - cleanup and push.
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
