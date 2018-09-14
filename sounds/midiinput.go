// +build darwin,linux,windows

package sounds

import (
	"fmt"
	"math"
	"time"

	pm "github.com/rakyll/portmidi"
	"gopkg.in/fatih/set.v0"
)

const (
	nsToSeconds        = 1e-9
	noteStart          = int64(144)
	noteEnd            = int64(128)
	pitchBend          = int64(224)
	nsPerCycle         = SecondsPerCycle * 1e9
	outputSampleBuffer = 1 // how many output samples are written in the same loop
	tickerDuration     = time.Duration(outputSampleBuffer) * DurationPerCycle
	MAX_CHANNELS       = 8
)

// Sample organ synth
var organOffset = [...]int{0, 12, 19, 24, 31, 34, 36, 38, 40}
var organVolume = [...]float64{0.18, 0.15, 0.7, 0.62, 1.0, 0.52, 0.4, 0.4, 0.4}
var organSum = 4.37

// A MidiInput is a sound that is wrapping a portmidi Midi input device.
type MidiInput struct {
	samples  chan float64
	deviceId pm.DeviceID
	running  bool
	// TODO(padster): use a more efficient, less general data type.
	notes *set.Set
}

// NewMidiInput takes a given midi device and converts it into a sound that plays
// what the device is playing (as sine waves), and stops once a pitch-bend is received.
func NewMidiInput(deviceId pm.DeviceID) Sound {
	ret := MidiInput{
		nil, /* samples */
		deviceId,
		false,     /* running */
		set.New(), /* notes */
	}
	return &ret
}

// GetSamples returns the samples for this sound, valid between a Start() and Stop()
func (s *MidiInput) GetSamples() <-chan float64 {
	// TODO(padster): Add some tracking here to make sure that GetSamples() is only called once
	// between each Start() and Stop(), if possible, to avoid re-using sounds.
	return s.samples
}

// Length returns the number of samples - unknown in advance, so it returns MaxLength.
func (s *MidiInput) Length() uint64 {
	return MaxLength
}

// Duration returns the duration of time the sound runs for, unknown as above.
func (s *MidiInput) Duration() time.Duration {
	return MaxDuration
}

// Start begins the Sound by opening two goroutines - one to take a set of active notes and convert
// it into sampled sine waves at the right frequencies, and the second to to listen to the midi input
// stream of events and convert that into the live set of active notes.
func (s *MidiInput) Start() {
	fmt.Println("Starting the MIDI sound's channel...")
	s.samples = make(chan float64)
	s.running = true

	// Goroutine to convert the s.notes set to samples.
	go func(midi *MidiInput) {
		fmt.Printf("  MIDI generation begun!\n")
		atNano := float64(time.Now().UnixNano())

		ticker := time.NewTicker(tickerDuration)
		defer ticker.Stop()

		for now := range ticker.C {
			if !midi.running {
				break
			}

			nowNano := float64(now.UnixNano())
			for ; atNano < nowNano && midi.running; atNano += nsPerCycle {
				if s.notes.IsEmpty() {
					if midi.running {
						midi.samples <- 0.0
					}
				} else {
					cycleAtMult := atNano * nsToSeconds
					value := 0.0
					for _, note := range s.notes.List() {
						// TODO(padster): Remove the * -> int64 -> int cast
						iNote := int(note.(int64))
						noteValue := 0.0
						for i, oOff := range organOffset {
							cps := midiToHz(iNote + oOff)
							offset := math.Remainder(cps*cycleAtMult, 1.0) * math.Pi * 2.0
							noteValue += math.Sin(offset) * organVolume[i]
						}
						noteValue *= 1.0 / organSum
						value += noteValue
					}
					if midi.running {
						midi.samples <- value / float64(s.notes.Size())
					}
				}
			}
		}
		close(s.samples)
	}(s)

	// Goroutine for reading from the input:
	go func() {
		fmt.Println("  Opening MIDI stream..")
		in, err := pm.NewInputStream(s.deviceId, 10)
		if err != nil {
			fmt.Printf("Error in reading midi device %d: Ensure portmidi is Initialized, and device is available.\n", s.deviceId)
			panic(err)
		}

		fmt.Println("Listening to stream")
		for event := range in.Listen() {
			// TODO - figure out what event.Data2 is (volumne?) and use it...
			fmt.Printf("Got: %v\n", event)
			if event.Status >= noteStart && event.Status < noteStart+MAX_CHANNELS {
				channel := event.Status - noteStart
				note := int64(event.Data1)
				// Drop channel 0 an octave
				if channel == 0 {
					note -= 12
				}
				s.notes.Add(note)
			} else if event.Status >= noteEnd && event.Status < noteEnd+MAX_CHANNELS {
				channel := event.Status - noteEnd
				note := int64(event.Data1)
				// Drop channel 0 an octave
				if channel == 0 {
					note -= 12
				}
				s.notes.Remove(note)
			} else if event.Status == pitchBend {
				s.Stop()
			}
			if !s.running {
				break
			}
		}
	}()
	// TODO(padster): Move goroutines into struct methods?
}

// Stop ends the sound, preventing any more samples from being written.
func (s *MidiInput) Stop() {
	// TODO(padster): close midi stream, stop timer.
	s.running = false
}

// Reset unsupported for the MIDI stream.
func (s *MidiInput) Reset() {
	panic("Can't reset live sound")
}

// Running returns whether Sound is still generating samples.
func (s *MidiInput) Running() bool {
	return s.running
}

// String returns the textual representation
func (s *MidiInput) String() string {
	return fmt.Sprintf("Midi[device #%d]", s.deviceId)
}

// TODO(padster): Merge this with the parser.go version once deps are sorted out.
// Also, it'd be good to precalculate all of these, to make midiToHz just a lookup.
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

func midiToHz(midiNote int) float64 {
	// Assuming C0 hz == 12 midi
	octave := midiNote/12 - 1
	semitone := midiNote % 12
	scale := 1 << uint(octave)
	return freq[semitone] * float64(scale)
}
