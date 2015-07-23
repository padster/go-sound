// Adds sounds together in parallel, and normalizes to (-1, 1) by dividing by the input count.
package sounds

import (
	"fmt"
	"math"
)

type NormalSum struct {
	wrapped    []Sound
	normScalar float64
}

func SumSounds(wrapped ...Sound) Sound {
	if len(wrapped) == 0 {
		panic("NormalSum can't take no sounds")
	}

	// TODO - durationMs := math.MaxUint64 ?
	var durationMs uint64 = math.MaxUint64
	for _, child := range wrapped {
		childDurationMs := child.DurationMs()
		if childDurationMs < durationMs {
			durationMs = childDurationMs
		}
	}

	sum := NormalSum{
		wrapped,
		1.0 / float64(len(wrapped)), /* normScalar */
	}
	return NewBaseSound(&sum, durationMs)
}
func (s *NormalSum) Run(base *BaseSound) {
	// TODO - start children in calling thread or running thread?
	for _, wrapped := range s.wrapped {
		wrapped.Start()
	}

	fmt.Println("Running sum...")
	for {
		sum := 0.0
		for _, wrapped := range s.wrapped {
			sample, stream_ok := <-wrapped.GetSamples()
			if !stream_ok || !base.Running() {
				base.Stop()
				break
			}
			sum += sample
		}

		if !base.WriteSample(sum * s.normScalar) {
			fmt.Println("Breaking sum")
			break
		}
	}

	fmt.Println("Sum stopping...")
}

func (s *NormalSum) Stop() {
	fmt.Println("Stop sum")
	for _, wrapped := range s.wrapped {
		wrapped.Stop()
	}
}

func (s *NormalSum) Reset() {
	fmt.Println("Reset sum")
	for _, wrapped := range s.wrapped {
		wrapped.Reset()
	}
}
