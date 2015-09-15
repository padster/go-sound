package main

import (
	"flag"
	"fmt"
	"runtime"
	"strings"
	"time"

	goflac "github.com/cocoonlife/goflac"
	"github.com/padster/go-sound/cq"
	s "github.com/padster/go-sound/sounds"
)

// Generates the golden files. See test/sounds_test.go for actual test.
func main() {
	// Singlethreaded for now...
	runtime.GOMAXPROCS(1)

	fmt.Printf("Parsing flags\n")
	minFreq := flag.Float64("minFreq", 110.0, "minimum frequency")
	maxFreq := flag.Float64("maxFreq", 14080.0, "maximum frequency")
	bpo := flag.Int("bpo", 24, "Buckets per octave")
	flag.Parse()
	remainingArgs := flag.Args()

	if len(remainingArgs) != 2 {
		panic("Required: <input> <output> filename arguments")
	}
	inputFile := remainingArgs[0]
	outputFile := remainingArgs[1]

	fmt.Printf("Building parameters\n")

	sampleRate := 44100.0
	// minFreq, maxFreq, bpo := 110.0, 14080.0, 24
	params := cq.NewCQParams(sampleRate, *minFreq, *maxFreq, *bpo)

	fmt.Printf("Params = %v\n", params)
	fmt.Printf("Building CQ and inverse... %v\n", params)

	constantQ := cq.NewConstantQ(params)
	cqInverse := cq.NewCQInverse(params)

	fmt.Printf("Loading file from %v\n", inputFile)
	if !strings.HasSuffix(inputFile, ".flac") {
		panic("Input file must be .flac")
	}

	flacSound := s.LoadFlacAsSound(inputFile)

	// TODO: Flac output writer for sound library.
	depth := 24
	fileWriter, err := goflac.NewEncoder(outputFile, 1, depth, int(sampleRate))
	if err != nil {
		fmt.Printf("Error opening file to write to! %v\n", err)
		panic("Can't write file")
	}
	defer fileWriter.Close()

	latency := constantQ.OutputLatency + cqInverse.OutputLatency

	fmt.Printf("forward latency = %d, inverse latency = %d, total = %d\n", constantQ.OutputLatency, cqInverse.OutputLatency, latency)

	startTime := time.Now()

	flacSound.Start()
	defer flacSound.Stop()

	// TODO: Make a common utility for this, it's used here and in both CQ and CQI.
	samples := cqInverse.ProcessChannel(constantQ.ProcessChannel(flacSound.GetSamples()))
	frameSize := 44100
	buffer := make([]float64, frameSize, frameSize)
	at := 0 
	for s := range samples {
		if at == frameSize {
			writeFrame(fileWriter, buffer)
			at = 0;
		}
		buffer[at] = s
		at++
	}
	writeFrame(fileWriter, buffer[:at])

	elapsedSeconds := time.Since(startTime).Seconds()
	fmt.Printf("elapsed time (not counting init): %f sec\n", elapsedSeconds)
}

func floatFromBitWithDepth(input int32, depth int) float64 {
	return float64(input) / (float64(unsafeShift(depth)) - 1.0) // Hmmm..doesn't seem right?
}
func intFromFloatWithDepth(input float64, depth int) int32 {
	return int32(input * (float64(unsafeShift(depth)) - 1.0))
}

func writeFrame(file *goflac.Encoder, samples []float64) { // samples in range [-1, 1]
	n := len(samples)
	frame := goflac.Frame{
		1,             /* channels */
		file.Depth,    /* depth */
		44100,         /* rate */
		make([]int32, n, n),
	}
	for i, v := range samples {
		frame.Buffer[i] = intFromFloatWithDepth(v, file.Depth)
	}
	if err := file.WriteFrame(frame); err != nil {
		fmt.Printf("Error writing frame to file :( %v\n", err)
		panic("Can't write to file")
	}
}

func unsafeShift(s int) int {
	return 1 << uint(s)
}
