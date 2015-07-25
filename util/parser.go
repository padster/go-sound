// Parser provides a collection of String -> Sound conversions
package util

import (
	"fmt"
	"strconv"

	s "github.com/padster/go-sound/sounds"
)

// Frequencies of notes in scale 0 - see:
// http://www.phy.mtu.edu/~suits/notefreqs.html
var freq = []float64{
	16.35, // C
	17.32, // C#/Db
	18.35, // D
	19.45, // D#/Eb
	20.60, // E
	21.83, // F
	23.12, // F#/Gb
	24.50, // G
	25.96, // G#/Ab
	27.50, // A
	29.14, // A#/Bb
	30.87, // B
}

// noteToMidi takes an offset into a string, parses a note, and returns
// both the note's base midi value plus the end offset of the note.
func noteToMidi(note string, offset int) (int, int) {
	resultSemi := 0

	switch note[offset] {
	case 'A':
		resultSemi = 9
	case 'B':
		resultSemi = 11
	case 'C':
		resultSemi = 0
	case 'D':
		resultSemi = 2
	case 'E':
		resultSemi = 4
	case 'F':
		resultSemi = 5
	case 'G':
		resultSemi = 7
	default:
		panic("Unknown note " + note)
	}
	offset++

	// TODO(padster): Support bb, ##, etc if useful.
	if offset < len(note) {
		switch note[offset] {
		case 'b':
			resultSemi--
			offset++
		case '#':
			resultSemi++
			offset++
		}
	}

	return resultSemi, offset
}

// midiToHz returns the Hz of a given midi note.
func midiToHz(midiNote int) float64 {
	// Assuming C0 hz == 12 midi
	octave := midiNote/12 - 1
	semitone := midiNote % 12
	scale := 1 << uint(octave)
	return freq[semitone] * float64(scale)
}

// MidiToSound converts a midi note into a sound that plays its pitch.
func MidiToSound(midiNote int) s.Sound {
	return s.NewSineWave(midiToHz(midiNote))
}

// noteToHz reads a note starting at an offset, and returns the hz and the end offset.
func noteToHz(note string, offset int, base uint) (float64, int) {
	midi, next := noteToMidi(note, offset)
	return midiToHz(midi + 12*int(base+1)), next
}

// noteToHz reads a note starting at an offset, and returns its Sound and the end offset.
func noteToSound(note string, offset int, base uint) (s.Sound, int) {
	baseHz, next := noteToHz(note, offset, base)
	return s.NewSineWave(baseHz), next
}

// ParseNotesToChord takes a collection of notes (e.g. "CEG") plus the base octave
// and returns a sound of them all being played together.
func ParseNotesToChord(notes string, base uint) s.Sound {
	asSounds := make([]s.Sound, 0, len(notes))
	var sound s.Sound
	for at := 0; at < len(notes); {
		sound, at = noteToSound(notes, at, base)
		asSounds = append(asSounds, sound)
	}
	return s.SumSounds(asSounds...)
}

// ParseChord converts a chord string (e.g. "G#sus4") and base octave into a
// Sound that contains the notes in the chord.
func ParseChord(chord string, base uint) s.Sound {
	baseMidi, at := noteToMidi(chord, 0)
	baseMidi += 12 * int(base+1)

	modifier := chord[at:]
	var offsets []int

	// A subset of these, converted to integer notation:
	// https://en.wikibooks.org/wiki/Music_Theory/Complete_List_of_Chord_Patterns
	switch modifier {
	case "":
		offsets = []int{0, 4, 7}
	case "5":
		offsets = []int{0, 7}
	case "m":
		offsets = []int{0, 3, 7}
	case "dom":
		fallthrough
	case "7":
		offsets = []int{0, 4, 7, 10}
	case "M7":
		offsets = []int{0, 4, 7, 11}
	case "m7":
		offsets = []int{0, 3, 7, 10}
	case "6":
		offsets = []int{0, 4, 7, 9}
	case "m6":
		offsets = []int{0, 3, 7, 9}
	case "dim":
		offsets = []int{0, 3, 6}
	case "sus4":
		offsets = []int{0, 5, 7}
	case "sus2":
		offsets = []int{0, 2, 7}
	case "aug":
		offsets = []int{0, 4, 8}

	default:
		panic("Unsupported chord modifier: " + modifier)
	}

	asSounds := make([]s.Sound, len(offsets), len(offsets))
	for i, offset := range offsets {
		asSounds[i] = MidiToSound(offset + baseMidi)
	}
	return s.SumSounds(asSounds...)
}

// GuitarChord converts a standard guitar representation (e.g. "2x0232")
// into the sound of those notes being played, assuming standard tuning.
func GuitarChord(chord string) s.Sound {
	// Standard guitar tuning: EADGBE
	stringMidi := []int{40, 45, 50, 55, 59, 64}
	noteMidi := []int{}

	for i, fret := range chord {
		if '0' <= fret && fret <= '9' {
			offset, _ := strconv.Atoi(fmt.Sprintf("%c", fret))
			noteMidi = append(noteMidi, stringMidi[i]+offset)
		}
	}

	asSounds := make([]s.Sound, len(noteMidi), len(noteMidi))
	for i, offset := range noteMidi {
		asSounds[i] = MidiToSound(offset)
	}
	return s.SumSounds(asSounds...)
}
