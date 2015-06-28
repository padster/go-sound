// Demo usage of the go-sound Sounds library.
package main

import (
	"github.com/padster/go-sound/output"
	s "github.com/padster/go-sound/sounds"
)

func main() {
	// TODO - fix it so it works at lower values (e.g. 215)
	b := int32(1000)
	chordC := s.SumSounds(
		s.NewTimedSound(s.NewSineWave(523.25), b), // C
		s.NewTimedSound(s.NewSineWave(659.25), b), // E
		s.NewTimedSound(s.NewSineWave(783.99), b), // G
	)
	chordA := s.SumSounds(
		s.NewSineWave(440), s.NewSineWave(523.25), s.NewSineWave(659.25))
	sound := s.ConcatSounds(
		chordC,
		s.NewTimedSound(chordA, b * 2),
	)
	// renderer := output.NewScreen(1200, 300)
	// renderer.Render(sound)
	output.Play(sound)
}
