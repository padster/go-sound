// Adds sounds together in parallel, and normalizes to (-1, 1) by dividing by the input count.
package sounds

type NormalSum struct {
  samples chan float64
  wrapped []Sound

  running bool
}

func SumSounds(wrapped ...Sound) *NormalSum {
  ret := NormalSum{
    make(chan float64),
    wrapped,
    false, /* running */
  }
  return &ret
}

func (s *NormalSum) GetSamples() <-chan float64 {
  return s.samples
}

func (s *NormalSum) Start() {
  s.running = true

  for _, wrapped := range s.wrapped {
    wrapped.Start()
  }

  if len(s.wrapped) > 0 {
    normScalar := 1.0 / float64(len(s.wrapped))

    go func() {
      for s.running {
        sum := 0.0
        for _, wrapped := range s.wrapped {
          sample, stream_ok := <-wrapped.GetSamples()
          if !stream_ok || !s.running {
            s.running = false
            break
          }
          sum += sample
        }

        if s.running {
          s.samples <- sum * normScalar
        }
      }

      s.running = false
      close(s.samples)
    }()
  }
}

func (s *NormalSum) Stop() {
  if s.running {
    s.running = false
    for _, wrapped := range s.wrapped {
      wrapped.Stop()
    }
    // close(s.samples)
  }
}

// TODO - implement properly (properly handle immediate changes while running)
func (s *NormalSum) Reset() {
  s.running = true
  for _, wrapped := range s.wrapped {
    wrapped.Reset()
  }
}
