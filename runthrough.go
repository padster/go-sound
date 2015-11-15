package main

import (
    // "fmt"
    "runtime"

    "github.com/padster/go-sound/output"
    "github.com/padster/go-sound/file"
    s "github.com/padster/go-sound/sounds"
    "github.com/padster/go-sound/util"
)

const (
    ONE_SECOND_MS = 1000.0
)

func sineWave() {
    // Example #1: Single tone = sine wave at a given frequency.
    sound := s.NewSineWave(440)
    demonstrateSound(sound, false)
}

func timedSineWave() {
    // Example #2: Single tone for a second.
    sound := s.NewTimedSound(s.NewSineWave(440), ONE_SECOND_MS)
    demonstrateSound(sound, true) 
}

func silence() {
    // Example #3: The sound of silence.
    sound := s.NewTimedSilence(ONE_SECOND_MS)
    demonstrateSound(sound, false) 
}

func oneNoteThenAnother() {
    // Example #4: One note (A 440) followed by a second (A 880)
    sound := s.ConcatSounds(
        s.NewTimedSound(s.NewSineWave(440), ONE_SECOND_MS / 2),
        s.NewTimedSound(s.NewSineWave(880), ONE_SECOND_MS / 2),
    )
    demonstrateSound(sound, true)
}

func simpleWaveTypes() {
    // Example #5: The sound of four basic repeating wave types: Sine, Triangle, Sawtooth, Square
    sound := s.ConcatSounds(
        s.NewTimedSound(s.NewSineWave(440), ONE_SECOND_MS),
        s.NewTimedSound(s.NewTriangleWave(440), ONE_SECOND_MS),
        s.NewTimedSound(s.NewSawtoothWave(440), ONE_SECOND_MS),
        s.NewTimedSound(s.NewSquareWave(440), ONE_SECOND_MS),
    )
    demonstrateSound(sound, true)
}

func twoNotesTogether() {
    // Example #6: Two notes at the same time! A 440 and E660
    sound := s.SumSounds(
        s.NewTimedSound(s.NewSineWave(440.00), ONE_SECOND_MS),
        s.NewTimedSound(s.NewSineWave(659.25), ONE_SECOND_MS),
    )
    demonstrateSound(sound, true)
}

func amplify() {
    // Example #7: One note, and four different amplifications 0, 0.5, 1.0, 1.5
    sound := s.ConcatSounds(
        s.MultiplyWithClip(s.NewTimedSound(s.NewSineWave(440.00), ONE_SECOND_MS / 2), 0.0), // Silence
        s.MultiplyWithClip(s.NewTimedSound(s.NewSineWave(440.00), ONE_SECOND_MS / 2), 0.5), // Quiet
        s.MultiplyWithClip(s.NewTimedSound(s.NewSineWave(440.00), ONE_SECOND_MS / 2), 1.0), // Loud
        s.MultiplyWithClip(s.NewTimedSound(s.NewSineWave(440.00), ONE_SECOND_MS / 2), 1.5), // Clipped!
    )
    demonstrateSound(sound, true)
}

func envelopes() {
    // Example #8: Play three notes normally, then silence, then repeat with envelopes.
    sound := s.ConcatSounds(
        s.NewTimedSound(util.MidiToSound(60), ONE_SECOND_MS / 2),
        s.NewTimedSound(util.MidiToSound(62), ONE_SECOND_MS / 2),
        s.NewTimedSound(util.MidiToSound(64), ONE_SECOND_MS / 2),
        s.NewTimedSilence(ONE_SECOND_MS / 2),
        s.NewADSREnvelope(s.NewTimedSound(util.MidiToSound(60), ONE_SECOND_MS / 2), 50, 200, 0.5, 100),
        s.NewADSREnvelope(s.NewTimedSound(util.MidiToSound(62), ONE_SECOND_MS / 2), 50, 200, 0.5, 100),
        s.NewADSREnvelope(s.NewTimedSound(util.MidiToSound(64), ONE_SECOND_MS / 2), 50, 200, 0.5, 100),
    )
    demonstrateSound(sound, true)
}

func pitchBending() {
    // Example #9: Smooth varying pitch.
    bendWave := s.NewTimedSound(s.NewSawtoothWave(2), 3 * ONE_SECOND_MS)
    bendWave.Start()
    bendWaveSamples := bendWave.GetSamples()

    pitchBendChannel := make(chan float64);
    go func() {
        for s := range bendWaveSamples {
            // [-1, 1] -> [440, 880]
            pitchBendChannel <- 440.0 + (s + 1.0) * 220.0
        }
    }()

    sound := s.NewHzFromChannel(pitchBendChannel)
    demonstrateSound(sound, true)
    bendWave.Stop()
}

func delay() {
    bendWave := s.NewTimedSound(s.NewTriangleWave(1.0 / 4.0), 4 * ONE_SECOND_MS)
    bendWave.Start()
    bendWaveSamples := bendWave.GetSamples()

    pitchBendChannel := make(chan float64);
    go func() {
        for s := range bendWaveSamples {
            pitchBendChannel <- 440.0 + (s + 1.0) * 220.0
        }
    }()

    sound := s.AddDelay(s.NewHzFromChannel(pitchBendChannel), ONE_SECOND_MS)
    // There appears to be a bug in writing this file :/
    demonstrateSound(sound, false)
}

func main() {
    runtime.GOMAXPROCS(4)

    exampleToRun := 8

    switch exampleToRun {
        case 1: sineWave()
        case 2: timedSineWave()
        case 3: silence()
        case 4: oneNoteThenAnother()
        case 5: simpleWaveTypes()
        case 6: twoNotesTogether()
        case 7: amplify()
        case 8: envelopes()
        case 9: pitchBending()
        case 10: delay()
    }
}


// Ignore this section - just splits the sound into three:
// One to be played, one to be rendered, and one to be saved if needed.
func demonstrateSound(sound s.Sound, saveFile bool) {
    saveFile = false // HACK
    sound.Start();
    samples := sound.GetSamples()

    toDraw, toPlay := make(chan float64), make(chan float64)
    
    var toSave chan float64
    if saveFile {
        toSave = make(chan float64)
    }

    go func() {
        for s := range samples {
            toDraw <- s
            toPlay <- s
            if saveFile {
                toSave <- s
            }
        }
        close(toDraw)
        close(toPlay)
        if saveFile {
            close(toSave)
        }
        sound.Stop()
    }()

    go output.Play(s.WrapChannelAsSound(toPlay))
    if saveFile {
        go soundfile.Write(s.WrapChannelAsSound(toSave), "out.wav")
    }
    output.Render(s.WrapChannelAsSound(toDraw), 600, 200, 18)
}