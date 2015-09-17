package main

import (
	"flag"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/padster/go-sound/cq"
	"github.com/padster/go-sound/output"
	s "github.com/padster/go-sound/sounds"
)

// Generates the golden files. See test/sounds_test.go for actual test.
func main() {
	// Singlethreaded for now...
	runtime.GOMAXPROCS(1)

	// Parse flags...
	minFreq := flag.Float64("minFreq", 110.0, "minimum frequency")
	maxFreq := flag.Float64("maxFreq", 14080.0, "maximum frequency")
	bpo := flag.Int("bpo", 24, "Buckets per octave")
	flag.Parse()

	remainingArgs := flag.Args()
	if len(remainingArgs) < 1 || len(remainingArgs) > 2 {
		panic("Required: <input> [<output>] filename arguments")
	}
	inputFile := remainingArgs[0]
	outputFile := ""
	if len(remainingArgs) == 2 {
		outputFile = remainingArgs[1]
	}

	// TODO: Better custom load method, to support more filetypes.
	if !strings.HasSuffix(inputFile, ".flac") {
		panic("Input file must be .flac")
	}
	inputSound := s.LoadFlacAsSound(inputFile)
	inputSound.Start()
	defer inputSound.Stop()

	// minFreq, maxFreq, bpo := 110.0, 14080.0, 24
	sampleRate := 44100.0
	params := cq.NewCQParams(sampleRate, *minFreq, *maxFreq, *bpo)
	constantQ := cq.NewConstantQ(params)
	cqInverse := cq.NewCQInverse(params)
	latency := constantQ.OutputLatency + cqInverse.OutputLatency

	startTime := time.Now()
	// TODO: Skip the first 'latency' samples for the stream.
	fmt.Printf("TODO: Skip latency (= %d) samples)\n", latency)
	samples := cqInverse.ProcessChannel(constantQ.ProcessChannel(inputSound.GetSamples()))
	asSound := s.WrapChannelAsSound(samples)

	if outputFile != "" {
		output.WriteSoundToFlac(asSound, outputFile)
	} else {
		output.Play(asSound)
	}

	elapsedSeconds := time.Since(startTime).Seconds()
	fmt.Printf("elapsed time (not counting init): %f sec\n", elapsedSeconds)
}
