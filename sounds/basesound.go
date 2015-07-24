// Sound implementation that is a single pure sine wave.
package sounds

type SoundDefinition interface {
	// TODO - add documentation and requirements.
	Run(base *BaseSound)
	Stop()
	Reset()
}

type BaseSound struct {
	samples    chan float64
	running    bool
	durationMs uint64
	definition SoundDefinition
}

// TODO - explain
func NewBaseSound(def SoundDefinition, durationMs uint64) *BaseSound {
	ret := BaseSound{
		// TODO - make the channel lazily
		make(chan float64),
		false, /* running */
		durationMs,
		def,
	}
	return &ret
}

func (s *BaseSound) GetSamples() <-chan float64 {
	return s.samples
}

func (s *BaseSound) DurationMs() uint64 {
	return s.durationMs
}

func (s *BaseSound) Start() {
	s.running = true
	go func() {
		s.definition.Run(s)
		s.Stop()
		s.definition.Stop()
		close(s.samples)
	}()
}

func (s *BaseSound) Stop() {
	s.running = false
}

func (s *BaseSound) Reset() {
	// TODO: Debug why this fails (see demo)
	// if s.running {
	// panic("Must call Stop before reset!")
	// }

	s.definition.Reset()
	s.samples = make(chan float64)
	s.running = true
}

func (s *BaseSound) WriteSample(sample float64) bool {
	if s.running {
		s.samples <- sample
	}
	return s.running
}

func (s *BaseSound) Running() bool {
	return s.running
}
