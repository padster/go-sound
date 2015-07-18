// Demo usage of the go-sound Sounds library.
package main

import (
	"github.com/padster/go-sound/output"
	s "github.com/padster/go-sound/sounds"
)

func main() {
	// TODO - fix it so it works at lower values (e.g. 215)
	b := float64(500)
	// var chordC s.Sound = s.ConcatSounds(
		// s.NewTimedSound(s.NewSineWave(523.25), b), // C
		// s.NewTimedSound(s.NewSineWave(659.25), b), // E
		// s.NewTimedSound(s.NewSineWave(783.99), b), // G
	// )
	// chordC = s.NewADSREnvelope(chordC, 50, 250, 0.3, 100)
	// chordA := s.SumSounds(
	// s.NewSineWave(440), s.NewSineWave(523.25), s.NewSineWave(659.25))
	// sound := s.ConcatSounds(
	// chordC,
	// s.NewTimedSound(chordA, b * 2),
	// )

	sound := s.ConcatSounds(
		s.NewADSREnvelope(
			s.NewTimedSound(s.NewSineWave(523.25), b), 25, 200, 0.3, 100),
		s.NewADSREnvelope(
			s.NewTimedSound(s.NewSineWave(659.25), b), 25, 200, 0.3, 100),
		s.NewADSREnvelope(
			s.NewTimedSound(s.NewSineWave(783.99), b), 25, 200, 0.3, 100),
	)

	// base := s.NewADSREnvelope(
	// s.NewTimedSound(s.NewSineWave(440), b), 25, 200, 0.3, 100)
	// sound := s.RepeatSound(chordC, 2)
	// sound := s.ConcatSounds(base,
	// s.NewADSREnvelope(s.NewTimedSound(s.NewSineWave(783.99), b), 25, 200, 0.3, 100))

	// chord := s.SumSounds(
		// s.LoadWavAsSound("piano.wav", 0),
		// s.NewSineWave(392.00),
	// )

	// sound := s.RepeatSound(s.NewTimedSound(chord, b), 3)

	// output.Play(sound)

	renderer := output.NewScreen(1500, 400)
	renderer.Render(sound)

	// TODO - modem faker: 440 for 0.5s, pause for 1s, 440 for 0.5s, pause for 1s, 880 for 1.5s
}
