package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"runtime"
	"strings"
	"time"

	goflac "github.com/cocoonlife/goflac"
	"github.com/padster/go-sound/cq"
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

	if len(remainingArgs) != 1 {
		panic("Required: <input> filename argument")
	}
	inputFile := remainingArgs[0]
	outputFile := "out.raw"

	fmt.Printf("Building parameters\n")

	sampleRate := 44100.0
	// minFreq, maxFreq, bpo := 110.0, 14080.0, 24
	params := cq.NewCQParams(sampleRate, *minFreq, *maxFreq, *bpo)

	fmt.Printf("Params = %v\n", params)
	fmt.Printf("Building Spectrogram... %v\n", params)

	spectrogram := cq.NewSpectrogram(params)

	fmt.Printf("Loading file from %v\n", inputFile)
	if !strings.HasSuffix(inputFile, ".flac") {
		panic("Input file must be .flac")
	}

	fileReader, err := goflac.NewDecoder(inputFile)
	if err != nil {
		fmt.Printf("Error reading file! %v\n", err)
		panic("Can't read file")
	}
	defer fileReader.Close()

	outputBuffer := bytes.NewBuffer(make([]byte, 0, 1024))

	inframe := 0

	startTime := time.Now()

	frame, err := fileReader.ReadFrame()
	for err != io.EOF {
		if frame.Depth != 24 {
			fmt.Printf("Only depth 24-bit flac supported for now, file is %d. TODO: support more...", frame.Depth)
			panic("Unsupported input file format")
		}
		if frame.Rate != 44100 {
			fmt.Printf("Only 44.1khz flac supported for now, file is %d. TODO: support more...", frame.Rate)
			panic("Unsupported input file format")
		}

		count := len(frame.Buffer) / frame.Channels
		samples := make([]float64, count, count)
		total := 0.0
		for i := 0; i < count; i++ {
			v := 0.0
			for _, c := range frame.Buffer[i*frame.Channels : (i+1)*frame.Channels] {
				v += floatFrom24bit(c)
			}
			v = v / float64(frame.Channels)
			samples[i] = v
			total += v
		}

		spectrogramBlock := spectrogram.Process(samples)
		for _, col := range spectrogramBlock {
			for _, c := range col {
				cq.WriteComplex(outputBuffer, c)
			}
		}

		inframe += count
		frame, err = fileReader.ReadFrame()
	}

	spectrogramBlock := spectrogram.GetRemainingOutput()
	for _, col := range spectrogramBlock {
		for _, c := range col {
			cq.WriteComplex(outputBuffer, c)
		}
	}

	ioutil.WriteFile(outputFile, outputBuffer.Bytes(), 0644)

	elapsedSeconds := time.Since(startTime).Seconds()
	fmt.Printf("elapsed time (not counting init): %f sec, frames/sec = %f\n",
		elapsedSeconds, float64(inframe)/elapsedSeconds)
}

func floatFrom24bit(input int32) float64 {
	return float64(input) / (float64(1<<23) - 1.0) // Hmmm..doesn't seem right?
}
func int24FromFloat(input float64) int32 {
	return int32(input * (float64(1<<23) - 1.0))
}

func writeFrame(file *goflac.Encoder, samples []float64) { // samples in range [-1, 1]
	n := len(samples)
	frame := goflac.Frame{
		1,     /* channels */
		24,    /* depth */
		44100, /* rate */
		make([]int32, n, n),
	}
	for i, v := range samples {
		frame.Buffer[i] = int24FromFloat(v)
	}
	if err := file.WriteFrame(frame); err != nil {
		fmt.Printf("Error writing frame to file :( %v\n", err)
		panic("Can't write to file")
	}
}
