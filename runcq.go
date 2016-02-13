package main

import (
	"flag"
	"fmt"
	"math/cmplx"
	"runtime"
	"time"

	"github.com/padster/go-sound/cq"
	f "github.com/padster/go-sound/file"
    "github.com/padster/go-sound/output"
	s "github.com/padster/go-sound/sounds"
)

// Runs CQ, applies some processing, and plays the result.
func main() {
	runtime.GOMAXPROCS(4)

	// Parse flags...
	sampleRate := s.CyclesPerSecond
	octaves := flag.Int("octaves", 7, "Range in octaves")
	minFreq := flag.Float64("minFreq", 55.0, "Minimum frequency")
	bpo := flag.Int("bpo", 24, "Buckets per octave")
	flag.Parse()

	remainingArgs := flag.Args()
	if len(remainingArgs) < 1 || len(remainingArgs) > 2 {
		panic("Required: <input> [<input>] filename arguments")
	}
	inputFile := remainingArgs[0]
	inputFile2 := inputFile
	if len(remainingArgs) == 2 {
		inputFile2 = remainingArgs[1]
	}

	inputSound := f.Read(inputFile)
	// inputSound := s.NewTimedSound(s.NewSineWave(440.0), 1000)
	inputSound.Start()
	defer inputSound.Stop()

	params := cq.NewCQParams(sampleRate, *octaves, *minFreq, *bpo)
	constantQ := cq.NewConstantQ(params)
	cqInverse := cq.NewCQInverse(params)
	latency := constantQ.OutputLatency + cqInverse.OutputLatency

	// Two inputs version - TODO, switch back to input + output.
	inputSound2 := f.Read(inputFile2)
	inputSound2.Start()
	defer inputSound2.Stop()
	constantQ2 := cq.NewConstantQ(params)

	startTime := time.Now()
	// TODO: Skip the first 'latency' samples for the stream.
	fmt.Printf("TODO: Skip latency (= %d) samples)\n", latency)
	columns := constantQ.ProcessChannel(inputSound.GetSamples())
	columns2 := constantQ2.ProcessChannel(inputSound2.GetSamples())
	samples := cqInverse.ProcessChannel(mergeChannels(columns, columns2))
	asSound := s.WrapChannelAsSound(samples)

	// if outputFile != "" {
	//f.Write(asSound, "brickwhite.wav")
	// } else {
	output.Play(asSound)
	// }

	elapsedSeconds := time.Since(startTime).Seconds()
	fmt.Printf("elapsed time (not counting init): %f sec\n", elapsedSeconds)
}

func mergeChannels(in1 <-chan []complex128, in2 <-chan []complex128) chan []complex128 {
	out := make(chan []complex128)
	go func() {
		fmt.Printf("Writing...\n")
		for cIn1 := range in1 {
			cIn2 := <-in2
			if len(cIn1) != len(cIn2) {
				panic("oops, columns don't match... :(")
			}
			cOut := make([]complex128, len(cIn1), len(cIn1))
			for i := range cIn1 {
				power1, angle1 := cmplx.Polar(cIn1[i])
				power1 = 1.0
				power2, angle2 := cmplx.Polar(cIn2[i])
				cOut[i] = cmplx.Rect(power1, angle1)
				// if i > 48 && i <= 72 {
				// cOut[i] = 0
				// }
				// HACK variable to stop go complaining about unused variables :(
				cIn2[i] = cmplx.Rect(power2, angle2)
			}
			out <- cOut
		}
		close(out)
	}()
	return out
}

func shiftChannel(buckets int, in <-chan []complex128) chan []complex128 {
	out := make(chan []complex128)
	go func(b int) {
		for cIn := range in {
			s := len(cIn)
			cOut := make([]complex128, s, s)
			copy(cOut, cIn[buckets:])
			out <- cOut
		}
		close(out)
	}(buckets)
	return out
}
