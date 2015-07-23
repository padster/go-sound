// Collection of String -> Sound conversions
package sounds

import (
	"fmt"
	"strconv"
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

func midiToHz(midiNote int) float64 {
	// C0 == 12
	octave := midiNote/12 - 1
	scale := 1 << uint(octave)
	semi := midiNote % 12
	return freq[semi] * float64(scale)
}

func MidiToSound(midiNote int) Sound {
	return NewSineWave(midiToHz(midiNote))
}

// Read a note starting at an offset, return the hz and the next offset to start from
func noteToHz(note string, offset int, base uint) (float64, int) {
	midi, offset := noteToMidi(note, offset)
	return midiToHz(midi + 12*int(base+1)), offset
}

func noteToSound(note string, offset int, base uint) (Sound, int) {
	baseHz, next := noteToHz(note, offset, base)
	return NewSineWave(baseHz), next
}

func ParseNotesToChord(notes string, base uint) Sound {
	asSounds := make([]Sound, 0, len(notes))

	var sound Sound
	at := 0
	for at < len(notes) {
		sound, at = noteToSound(notes, at, base)
		asSounds = append(asSounds, sound)
	}

	return SumSounds(asSounds...)
}

func ParseChord(chord string, base uint) Sound {
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

	asSounds := make([]Sound, len(offsets), len(offsets))
	for i, offset := range offsets {
		asSounds[i] = MidiToSound(offset + baseMidi)
	}
	return SumSounds(asSounds...)
}

func GuitarChord(chord string) Sound {
	// Standard guitar tuning: EADGBE
	stringMidi := []int{40, 45, 50, 55, 59, 64}

	noteMidi := []int{}

	for i, fret := range chord {
		if '0' <= fret && fret <= '9' {
			offset, _ := strconv.Atoi(fmt.Sprintf("%c", fret))
			noteMidi = append(noteMidi, stringMidi[i]+offset)
		}
	}
	// TODO - share with above?
	asSounds := make([]Sound, len(noteMidi), len(noteMidi))
	for i, offset := range noteMidi {
		asSounds[i] = MidiToSound(offset)
	}
	return SumSounds(asSounds...)
}
