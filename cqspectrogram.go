package main

import (
	// "bytes"
	"flag"
	"fmt"
	// "io/ioutil"
	"runtime"
	"time"

	"github.com/padster/go-sound/cq"
	s "github.com/padster/go-sound/sounds"
	"github.com/padster/go-sound/output"
	"github.com/padster/go-sound/util"
)

// Generates the golden files. See test/sounds_test.go for actual test.
func main() {
	// Needs to be at least 2 when doing openGL + sound output at the same time.
	runtime.GOMAXPROCS(3)

	minFreq := flag.Float64("minFreq", 110.0, "minimum frequency")
	maxFreq := flag.Float64("maxFreq", 3520.0, "maximum frequency")
	bpo := flag.Int("bpo", 48, "Buckets per octave")
	flag.Parse()
	remainingArgs := flag.Args()

	if len(remainingArgs) != 1 {
		panic("Required: <input> filename argument")
	}
	inputFile := remainingArgs[0]
	fmt.Printf("%s\n", inputFile)
	// TODO: renable writing out to file if one is provided.
	// outputFile := "out.raw"

	sampleRate := 44100.0
	// minFreq, maxFreq, bpo := 110.0, 14080.0, 24
	params := cq.NewCQParams(sampleRate, *minFreq, *maxFreq, *bpo)
	spectrogram := cq.NewSpectrogram(params)

	// inputSound := s.LoadFlacAsSound(inputFile)
	inputSound := s.ConcatSounds(
		s.NewTimedSound(s.NewSineWave(440), 5000),
		s.NewTimedSound(s.NewSineWave(880), 5000),
		s.NewTimedSound(s.SumSounds(s.NewSineWave(440), s.NewSineWave(880)), 5000),
	)
	inputSound.Start()
	defer inputSound.Stop()

	soundChannel, specChannel := splitChannel(inputSound.GetSamples())
	startTime := time.Now()

	go func() {
		columns := spectrogram.ProcessChannel(specChannel)
		toShow := util.NewSpectrogramScreen(441, *bpo * 5 * 2)
		toShow.Render(columns, 4)
	}()

	output.Play(s.WrapChannelAsSound(soundChannel))	

/*
	// TODO: renable writing out to file.
	outputBuffer := bytes.NewBuffer(make([]byte, 0, 1024))
	width, height := 0, 0
	for col := range columns {
		for _, c := range col {
			cq.WriteComplex(outputBuffer, c)
		}
		if width % 1000 == 0 {
			fmt.Printf("At frame: %d\n", width)
		}
		width++
		height = len(col)
	}
	fmt.Printf("Done! - %d by %d\n", width, height)
	ioutil.WriteFile(outputFile, outputBuffer.Bytes(), 0644)
*/

	elapsedSeconds := time.Since(startTime).Seconds()
	fmt.Printf("elapsed time (not counting init): %f sec\n", elapsedSeconds)
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
	}()
	return r1, r2
}
