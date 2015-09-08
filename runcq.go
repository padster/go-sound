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
	SZ := 16
	VALS := 16

	test := make([]float64, VALS, VALS)
	for i := 0; i < VALS; i++ {
		test[i] = float64(i - VALS / 2) / float64(SZ)
	}

	co := fft.FFTReal(test[:SZ])
	for i, v := range co {
		fmt.Printf("v[%d] = %.6f %.6f\n", i, real(v), imag(v));
	}
	
	co2 := fft.IFFT(co)
	fmt.Printf("#2: %v\n", co2)
	for i, v := range co2 {
		fmt.Printf("v[%d] = %.6f %.6f\n", i, real(v), imag(v));
	}
	*/

	

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

	fileReader, err := goflac.NewDecoder(inputFile)
	if err != nil {
		fmt.Printf("Error reading file! %v\n", err)
		panic("Can't read file")
	}
	defer fileReader.Close()

	fileWriter, err := goflac.NewEncoder(outputFile, 1, 24, int(sampleRate))
	if err != nil {
		fmt.Printf("Error opening file to write to! %v\n", err)
		panic("Can't write file")
	}
	defer fileWriter.Close()

	// HACK - serialize results for now
	// resultBuffer := new(bytes.Buffer)

	inframe := 0
	outframe := 0
	latency := constantQ.OutputLatency + cqInverse.OutputLatency

	fmt.Printf("forward latency = %d, inverse latency = %d, total = %d\n", constantQ.OutputLatency, cqInverse.OutputLatency, latency)

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

		result := constantQ.Process(samples, false /* TODO - remove */)
		fmt.Printf("%d values returned from CQ\n", len(result))

		cqout := cqInverse.Process(result, frameAt == PRINT_FRAME)
		fmt.Printf("%d values returned from CQI\n", len(cqout))

		// cq.WriteComplexBlock(resultBuffer, result)

		// for i, v := range result {
		// 	c := complex(0, 0)
		// 	for _, w := range v {
		// 		c = c + w
		// 	}
		// 	if frameAt == PRINT_FRAME {
		// 		fmt.Printf("result[%d], size = %d, sum = (%.4f, %.4f)\n", i, len(v), real(c), imag(c));
		// 	}
		// 	if (frameAt == PRINT_FRAME && i == 1) {
		// 		for j := 1; j < len(v); j++ {
		// 			fmt.Printf("(%.3f,%.3f), ", real(v[j]), imag(v[j]));
		// 		}
		// 		fmt.Printf("\n");
		// 	}
		// }

		if frameAt == PRINT_FRAME {
			s := 0.0
			for _, v := range cqout {
				s += v
			}
			fmt.Printf("Sum of cqi out: %.6f\n", s)
		}

		if (frameAt == PRINT_FRAME) {
			panic("DONE")
		}

		frameAt++

		for i := 0; i < len(cqout); i++ {
			if cqout[i] > 1.0 {
				cqout[i] = 1.0
			} else if cqout[i] < -1.0 {
				cqout[i] = -1.0
			} 
		}

		if (outframe >= latency) {	
			writeFrame(fileWriter, cqout)

		} else if (outframe + len(cqout) >= latency) {
			offset := latency - outframe
			writeFrame(fileWriter, cqout[offset:])
		}

		inframe += count
		outframe += len(cqout)
		frame, err = fileReader.ReadFrame()
	}

	r := cqInverse.Process(constantQ.GetRemainingOutput(), false)
	r2 := cqInverse.GetRemainingOutput()
	r = append(r, r2...)

	for i := 0; i < len(r); i++ {
		if r[i] > 1.0 {
			r[i] = 1.0
		}
		if r[i] < -1.0 {
			r[i] = -1.0
		}
	}

	writeFrame(fileWriter, r)
	outframe += len(r)

	fmt.Printf("in: %d, out: %d\n", inframe, outframe - latency)
	// TODO: latency	
}

func floatFrom24bit(input int32) float64 {
	return float64(input) / (float64(1 << 23) - 1.0) // Hmmm..doesn't seem right?
}
func int24FromFloat(input float64) int32 {
	return int32(input * (float64(1 << 23) - 1.0))
}

func writeFrame(file *goflac.Encoder, samples []float64) { // samples in range [-1, 1]
	n := len(samples)
	frame := goflac.Frame {
		1, /* channels */
		24, /* depth */
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