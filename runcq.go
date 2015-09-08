package main

// NOTE: Don't use yet, not ready.

// TODO: 
// - Finish kernel
//   - window functions
//   - hook up FFT
//   - verify with C++?
// - Constant Q transform
// - Inverse CQ transform.

import (
	// "bytes"
	"flag"
	"fmt"
	"io"
	"strings"

	goflac "github.com/cocoonlife/goflac"
	"github.com/padster/go-sound/cq"
  // "github.com/mjibson/go-dsp/fft"
)



// Generates the golden files. See test/sounds_test.go for actual test.
func main() {
	fmt.Printf("Parsing flags\n")
	minFreq := flag.Float64("minFreq", 110.0, "minimum frequency")
	maxFreq := flag.Float64("maxFreq", 14080.0, "maximum frequency")
	bpo := flag.Int("bpo", 24, "Buckets per octave")
	flag.Parse()
	remainingArgs := flag.Args()


	//  BIG HACK - need to get FFT agreeing.
	/*
	SZ := 32
	VALS := 64

	test := make([]float64, VALS, VALS)
	for i := 0; i < VALS; i++ {
		test[i] = float64(i - VALS / 2) / float64(SZ)
	}

	co := fft.FFTReal(test[:SZ])
	for i, v := range co {
		fmt.Printf("v[%d] = %.4f %.4f\n", i, real(v), imag(v));
	}
	*/

	if len(remainingArgs) != 1 {
		panic("Required: <input> filename argument")
	}
	inputFile := remainingArgs[0]

	fmt.Printf("Building parameters\n")

	sampleRate := 44100.0
	// minFreq, maxFreq, bpo := 110.0, 14080.0, 24
	params := cq.NewCQParams(sampleRate, *minFreq, *maxFreq, *bpo)

	fmt.Printf("Params = %v\n", params)
	fmt.Printf("Building CQ... %v\n", params)

	constantQ := cq.NewConstantQ(params)

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

	// HACK - serialize results for now
	// resultBuffer := new(bytes.Buffer)

	frame, err := fileReader.ReadFrame()
	PRINT_FRAME := -1
	frameAt := 0
	for err != io.EOF {
		if frame.Depth != 24 {
			panic("Only depth 24-bit flac supported for now. TODO: support more...")
		}
		if frame.Rate != 44100 {
			panic("Only 44.1khz flac supported for now. TODO: support more...")
		}

		count := len(frame.Buffer) / frame.Channels
		samples := make([]float64, count, count)
		total := 0.0
		for i := 0; i < count; i++ {
			v := 0.0
			for _, c := range frame.Buffer[i * frame.Channels:(i+1) * frame.Channels] {
				v += floatFrom24bit(c)
			}
			v = v / float64(frame.Channels)
			samples[i] = v
			total += v
		}

		result := constantQ.Process(samples, frameAt == PRINT_FRAME)

		// cq.WriteComplexBlock(resultBuffer, result)

		for i, v := range result {
			c := complex(0, 0)
			for _, w := range v {
				c = c + w
			}
			if frameAt == PRINT_FRAME {
				fmt.Printf("result[%d], size = %d, sum = (%.4f, %.4f)\n", i, len(v), real(c), imag(c));
			}
			if (frameAt == PRINT_FRAME && i == 1) {
				for j := 1; j < len(v); j++ {
					fmt.Printf("(%.3f,%.3f), ", real(v[j]), imag(v[j]));
				}
				fmt.Printf("\n");
			}
		}

		if (frameAt == PRINT_FRAME) {
			panic("DONE")
		}

		frame, err = fileReader.ReadFrame()
		frameAt++
	}

	// fmt.Printf("Done! - %d bytes written\n", resultBuffer.Len())
}

func floatFrom24bit(input int32) float64 {
	return float64(input) / (float64(1 << 23) - 1.0) // Hmmm..doesn't seem right?
}


// Frame 685:
// C: 							3.96983290
// Go without - 1: 	3.96982783
// Go with -1: 			3.96982830