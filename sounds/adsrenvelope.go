/* 
Envelope based on Attack, Decay, Sustain, Release config.
https://en.wikipedia.org/wiki/Synthesizer#ADSR_envelope 
*/
package sounds

type ADSREnvelope struct {
  samples chan float64
  wrapped Sound

  msAt float64
  attackMs float64
  sustainStartMs float64
  sustainEndMs float64
  sustainLevel float64

  durationMs   uint64
  running      bool
}

func NewADSREnvelope(wrapped Sound, 
  attackMs float64, delayMs float64, sustainLevel float64, releaseMs float64) *ADSREnvelope {

  durationMs := wrapped.DurationMs()

  ret := ADSREnvelope{
    make(chan float64),
    wrapped,
    0,                               /* msAt */
    attackMs,                        /* attackMs */
    attackMs + delayMs,              /* sustainStartMs */
    float64(durationMs) - releaseMs, /* sustainEndMs */
    sustainLevel,
    durationMs,
    false,                           /* running */
  }
  return &ret
}

func (s *ADSREnvelope) GetSamples() <-chan float64 {
  return s.samples
}

func (s *ADSREnvelope) DurationMs() uint64 {
  return s.durationMs
}

func (s *ADSREnvelope) Start() {
  s.running = true
  s.wrapped.Start()

  floatDuration := float64(s.durationMs)
  go func() {
    for s.msAt < floatDuration && s.running {
      scale := float64(0) // [0, 1]

      switch {
        case s.msAt < s.attackMs:
          scale = s.msAt / s.attackMs
        case s.msAt < s.sustainStartMs:
          scale = 1 - (1 - s.sustainLevel) * (s.msAt - s.attackMs) / (s.sustainStartMs - s.attackMs)
        case s.msAt < s.sustainEndMs:
          scale = s.sustainLevel
        default:
          scale = s.sustainLevel * (s.msAt - floatDuration) / (s.sustainEndMs - floatDuration)
      }

      next := <-s.wrapped.GetSamples()
      s.samples <- next * scale
      s.msAt += SecondsPerCycle() * 1000.0
    }

    s.Stop()
    close(s.samples)
  }()
}

func (s *ADSREnvelope) Stop() {
  s.wrapped.Stop()
  s.running = false
}

func (s *ADSREnvelope) Reset() {
  if s.running {
    panic("Stop before reset!")
  }

  s.samples = make(chan float64)
  s.msAt = float64(0)
  s.wrapped.Reset()
  s.running = true
}
