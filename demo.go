// Demo usage of the go-sound Sounds library, to play Clair de Lune.
package main

import (
	"fmt"
	"math"
	"runtime"

	// "github.com/padster/go-sound/cq"
	// file "github.com/padster/go-sound/file"
	"github.com/padster/go-sound/output"
	s "github.com/padster/go-sound/sounds"
	"github.com/padster/go-sound/util"

	pm "github.com/rakyll/portmidi"
)

const (
	// ~65 bpm ~= 927 ms/b ~= 309 ms/quaver (in 9/8)
	q = float64(309)
)

// Notes in a treble clef, centered on B (offset 8)
var trebleMidi = [...]int{57, 59, 60, 62, 64, 65, 67, 69, 71, 72, 74, 76, 77, 79, 81, 83, 84}

// Key is D♭-Major = five flats: A♭, B♭, D♭, E♭, G♭
var trebleKeys = [...]int{-1, -1, 00, -1, -1, 00, -1, -1, -1, 00, -1, -1, 00, -1, -1, -1, 00}

func main() {
	// NOTE: Not required, but shows how this can run on multiple threads.
	runtime.GOMAXPROCS(8)

	// Switch on whichever demo you'd like here:
	if false {
		renderMidi()
	} else {
		playClairDeLune()
	}
}

// renderMidi reads a midi device, converts it to a live sound, and renders the waveform to screen.
func renderMidi() {
	fmt.Println("Loading portmidi...")
	pm.Initialize()

	midi := s.NewMidiInput(pm.DeviceId(3)) // ++See below
	fmt.Printf("Rendering midi...\n")

	// Render the generated sine wave to screen:
	output.Render(midi, 2000, 500)
	// ...or, play it live:
	// output.Play(midi)

	/*
		++ Note: to find your device id, use a version of this:
		for i := 1; i <= pm.CountDevices(); i++ {
			fmt.Printf("Reading device: %d\n", i)
			di := pm.GetDeviceInfo(pm.DeviceId(i))
			fmt.Printf("Info = %v\n", di)
		}
	*/
}

// playClairDeLune builds then plays Clair de Lune (Debussey)
// music from http://www.piano-midi.de/noten/debussy/deb_clai.pdf
func playClairDeLune() {
	fmt.Println("Building sound.")

	/*
	finalNoteLength := float64(3 + 6) // 6 extra beats, just for effect

	// Left-hand split for a bit near the end.
	rh1 := s.ConcatSounds(
		notesT(7, fs(1)),
		notesTQRun(0, 1, 0, 3, 0, -1, 0),
		notesT(2, fs(-1)), notesTQRun(-2, -1), notesT(3, fs(-2)), notesT(finalNoteLength, fs(-3)),
	)
	rh2 := s.ConcatSounds(
		notesT(6, fs(-1)),
		notesT(6, fs(-2)), notesT(3, fs(-2)),
		notesT(6, fs(-4)), notesT(finalNoteLength, fs(-4)),
	)

	// Split of couplets over long Bb
	couplets := s.SumSounds(
		s.ConcatSounds(notesT(1.5, fs(2)), notesT(3, fs(4)), notesT(2.5, fs(2))),
		notesT(7, fs(0)),
	)

	// Top half of the score:
	rightHand := s.ConcatSounds(
		rest(2), notesT(4, fs(4, 6)), notesT(4, fs(2, 4)),
		notesT(1, fs(1, 3)), notesT(1, fs(2, 4)), notesT(7, fs(1, 3)),
		notesT(1, fs(0, 2)), notesT(1, fs(1, 3)), couplets,
		notesT(1, fs(-1, 1)), notesT(1, fs(0, 2)), s.SumSounds(rh1, rh2),
	)

	// Bottom half.
	leftHand := s.ConcatSounds(
		rest(1), notesT(8, fs(-1, -3)),
		notesT(9, fs(-0.5, -2)),
		notesT(9, fs(-1, -3)),
		notesT(9, fs(-2, -4)),
		notesT(6, fs(-4, -5)),
		notesT(3, fs(-4, -6)),
		notesT(6, fs(-5, -7)), // HACK: Actually in bass clef, but rewritten in treble for these two chords.
		notesT(finalNoteLength, fs(-6, -7.5)),
	)

	clairDeLune := s.SumSounds(leftHand, rightHand)
	*/
	// toPlay := s.NewDenseIIR(clairDeLune,
		// []float64{0.8922, -2.677, 2.677, -0.8922},
		// []float64{2.772, -2.57, 0.7961},
	// )
	
	// hz := 440.0
	// toPlay := s.SumSounds(
		// s.NewSineWave(hz),
		// s.NewSquareWave(hz),
		// s.NewSawtoothWave(hz),
		// s.NewTriangleWave(hz),
	// )
	toPlay := s.NewJackInput("go-sound-in")
	// toPlay := s.NewTimedSound(s.NewSineWave(500), 1000)
	// toPlay := s.SumSounds(s1, s2)

	// toPlay := s.NewTimedSound(shephardTones(), 10000)
	// toPlay := file.Read("greatgig.flac")
	// file.Write(toPlay, "gg.wav")
	// fmt.Printf("Playing: \n\t%s\n", toPlay)
	// output.Render(toPlay, 2000, 400)
	// output.PlayJack(toPlay)
	output.Play(toPlay)

	// output.Play(s.LoadFlacAsSound("toneslide.flac"))

	// Optional: Write to a .wav file:
	// clairDeLune.Reset()
	// fmt.Println("Writing sound to file.")
	// file.Write(clairDeLune, "clairdelune.wav")

	// Optional: Draw to screen:
	// clairDeLune.Reset()
	// fmt.Println("Drawing sound to screen.")
	// output.Render(clairDeLune, 2000, 400)
}

// fs is a short way to write an array of floats.
func fs(fs ...float64) []float64 {
	return fs
}

// The Sound of silence for quaverCount quavers
func rest(quaverCount float64) s.Sound {
	return s.NewTimedSilence(q * quaverCount)
}

// A chord of notes in the treble clef, 0 = B, then notes up and down (e.g. -4 = E, 4 = F)
// in the proper key (Db major), with +/- 0.5 signifying a sharp or flat.
func notesT(quaverCount float64, notes []float64) s.Sound {
	sounds := make([]s.Sound, len(notes), len(notes))
	for i, note := range notes {
		sounds[i] = noteTMidi(note, quaverCount)
	}
	return s.SumSounds(sounds...)
}

// A run of quavers in the treble clef
func notesTQRun(notes ...float64) s.Sound {
	sounds := make([]s.Sound, len(notes), len(notes))
	for i, note := range notes {
		sounds[i] = noteTMidi(note, 1.0)
	}
	return s.ConcatSounds(sounds...)
}

// Converts a treble note offset to a midi offset
func noteTMidi(note float64, quaverCount float64) s.Sound {
	// NOTE: Only [-8, 8] allowed for 'note'.
	bFloat, sharp := math.Modf(note)
	base := int(bFloat)
	if sharp < 0 {
		sharp += 1.0
		base--
	}

	// 0 = B = offset 8
	midi := trebleMidi[base+8] + trebleKeys[base+8]
	if sharp > 0.1 {
		midi++
	}
	midiToSound := s.NewTimedSound(util.MidiToSound(midi), quaverCount*q)
	return s.NewADSREnvelope(midiToSound, 15, 50, 0.5, 20)
}

// Shephard tones
func shephardTones() s.Sound {
	octaves := 5
	base, mid := 110.0, 155.563491861

	tones := 2 * octaves
	bases := make([]float64, tones, tones)
	for i := 0; i < octaves; i++ {
		bases[2 * i] = base * float64(unsafeShift(i))
		bases[2*i+1] = mid * float64(unsafeShift(i))
	}
	secondsPerOctave := 10

	maxHz := bases[0] * float64(unsafeShift(octaves))
	downOctaves := 1.0 / float64(unsafeShift(octaves))

	samplesPerOctave := int(secondsPerOctave * s.CyclesPerSecond)
	octavesPerSample := 1.0 / float64(samplesPerOctave)

	channels := make([]chan []float64, tones, tones)
	for i := 0; i < tones; i++ {
		channels[i] = make(chan []float64)
	}
	go func() {
		for {
			for sample := 0; sample < octaves * samplesPerOctave; sample++ {
				for i := 0; i < tones; i++ {
					hz := bases[i] * math.Pow(2.0, float64(sample) * octavesPerSample)
					if hz >= maxHz {
						hz *= downOctaves
					}
					channels[i] <- []float64{hz, gaussianAmplitude(hz, bases[0], maxHz)}
				}
			}
		}
	}()

	sounds := make([]s.Sound, tones, tones)
	for i, v := range channels {
		sounds[i] = s.NewHzFromChannelWithAmplitude(v)
	}
	return s.SumSounds(sounds...)
}

func gaussianAmplitude(at float64, minHz float64, maxHz float64) float64 {
	lHalf := 0.5 * (math.Log(minHz) + math.Log(maxHz))
	diff := (math.Log(at) - lHalf)
	return math.Exp(-1.0 * diff * diff)
}

func unsafeShift(s int) int {
	return 1 << uint(s)
}
