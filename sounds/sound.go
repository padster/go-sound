// API for the Sound type.
package sounds

type Sound interface {
	GetSamples() <-chan float64

	Start()
	Stop()
	Reset()
}

// Global constant for the sample rate of each sound stream.
func CyclesPerSecond() float64 {
	return 44100.0
}
