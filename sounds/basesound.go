package sounds

import (
	"time"
)

// A SoundDefinition represents the simplified requirements that BaseSound converts into a Sound
type SoundDefinition interface {
	// Run executes the normal logic of the sound, writing to base.WriteSample until it is false.
	Run(base *BaseSound)

	// Stop cleans up at the end of the Sound.
	Stop()

	// Reset rewrites all state to the same as before Run()
	Reset()
}

// A BaseSound manages state around the definition, and adapts all the Sound methods.
type BaseSound struct {
	samples     chan float64
	running     bool
	sampleCount uint64
	duration    time.Duration
	definition  SoundDefinition
}

// NewBaseSound takes a simpler definition of a sound, plus a duration, and
// converts them into something that implements the Sound interface.
func NewBaseSound(def SoundDefinition, sampleCount uint64) Sound {
	duration := SamplesToDuration(sampleCount)

	ret := BaseSound{
		nil, /* samples */
		false, /* running */
		sampleCount,
		duration,
		def,
	}
	return &ret
}

// GetSamples returns the samples for this sound, valid between a Start() and Stop()
func (s *BaseSound) GetSamples() <-chan float64 {
	return s.samples
}

// Length returns the provided number of samples for this sound.
func (s *BaseSound) Length() uint64 {
	return s.sampleCount
}

// Duration returns the duration of time the sound runs for.
func (s *BaseSound) Duration() time.Duration {
	return s.duration
}

// Start begins the Sound by initialzing the channel, running the definition
// on a separate goroutine, and cleaning up once it has finished.
func (s *BaseSound) Start() {
	s.running = true
	s.samples = make(chan float64)

	// NOTE: It may make sense to move things to the other side of this goroutine boundary.
	// e.g. Whether to start/stop child sounds are inside the goroutine, but can be moved
	// outside if Run() is split into two calls, one in and one out.
	go func() {
		s.definition.Run(s)
		s.Stop()
		close(s.samples)
	}()
}

// Stop ends the sound, preventing any more samples from being written.
func (s *BaseSound) Stop() {
	s.running = false
	s.definition.Stop()
}

// Reset clears all the state in a stopped sound back to pre-Start values.
func (s *BaseSound) Reset() {
	if s.running {
		panic("Must call Stop before reset!")
	}
	s.definition.Reset()
}

// WriteSample appends a sample to the channel, returning whether the write was successful.
func (s *BaseSound) WriteSample(sample float64) bool {
	if s.running {
		s.samples <- sample
	}
	return s.running
}

// Running returns whether Sound is still generating samples.
func (s *BaseSound) Running() bool {
	return s.running
}
