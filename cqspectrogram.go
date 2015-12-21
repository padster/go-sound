package main

import (
	"flag"
	"fmt"
	"runtime"
	"time"

	"github.com/padster/go-sound/cq"
	f "github.com/padster/go-sound/file"
	"github.com/padster/go-sound/output"
	s "github.com/padster/go-sound/sounds"
	"github.com/padster/go-sound/util"
)

// Generates the golden files. See test/sounds_test.go for actual test.
func main() {
	// Needs to be at least 2 when doing openGL + sound output at the same time.
	runtime.GOMAXPROCS(6)

	sampleRate := s.CyclesPerSecond
	octaves := flag.Int("octaves", 7, "Range in octaves")
	minFreq := flag.Float64("minFreq", 55.0, "Minimum frequency")
	bpo := flag.Int("bpo", 24, "Buckets per octave")
	flag.Parse()

	remainingArgs := flag.Args()
	argCount := len(remainingArgs)
	if argCount < 1 || argCount > 2 {
		panic("Required: <input> [<output>] filename arguments")
	}
	inputFile := remainingArgs[0]
	outputFile := ""
	if argCount == 2 {
		outputFile = remainingArgs[1]
	}

	// minFreq, maxFreq, bpo := 110.0, 14080.0, 24
	params := cq.NewCQParams(sampleRate, *octaves, *minFreq, *bpo)
	spectrogram := cq.NewSpectrogram(params)

	inputSound := f.Read(inputFile)
	inputSound.Start()
	defer inputSound.Stop()

	startTime := time.Now()
	if outputFile != "" {
		columns := spectrogram.ProcessChannel(inputSound.GetSamples())
		// Write to file
		f.WriteColumns(outputFile, columns)
	} else {
		// No file, so play and show instead:
		soundChannel, specChannel := splitChannel(inputSound.GetSamples())
		go func() {
			columns := spectrogram.ProcessChannel(specChannel)
			toShow := util.NewSpectrogramScreen(882, *bpo**octaves, *bpo)
			toShow.Render(columns, 1)
		}()
		output.Play(s.WrapChannelAsSound(soundChannel))
	}

	elapsedSeconds := time.Since(startTime).Seconds()
	fmt.Printf("elapsed time (not counting init): %f sec\n", elapsedSeconds)

	if outputFile == "" {
		// Hang around to the view can be looked at.
		for {
		}
	}
}

// HACK - move to utils, support in both main apps.
func floatFrom16bit(input int32) float64 {
	return float64(input) / (float64(1<<15) - 1.0) // Hmmm..doesn't seem right?
}
func int16FromFloat(input float64) int32 {
	return int32(input * (float64(1<<15) - 1.0))
}

func floatFrom24bit(input int32) float64 {
	return float64(input) / (float64(1<<23) - 1.0) // Hmmm..doesn't seem right?
}
func int24FromFloat(input float64) int32 {
	return int32(input * (float64(1<<23) - 1.0))
}

func splitChannel(samples <-chan float64) (chan float64, chan float64) {
	r1, r2 := make(chan float64), make(chan float64)
	go func() {
		for s := range samples {
			r1 <- s
			r2 <- s
		}
		close(r1)
		close(r2)
	}()
	return r1, r2
}
