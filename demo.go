// Demo usage of the go-sound Sounds library.
package main

import (
	"github.com/padster/go-sound/output"
	s "github.com/padster/go-sound/sounds"
)

func main() {
	// TODO - fix it so it works at lower values (e.g. 215)
	b := int32(500)
	sound := s.ConcatSounds(
		s.NewTimedSound(s.NewSineWave(440), b),
		s.NewTimedSound(s.NewSineWave(880), b),
		s.NewTimedSound(s.NewSineWave(440), b),
	)
	// renderer := output.NewScreen(1200, 300)
	// renderer.Render(sound)
	output.Play(sound)
}
