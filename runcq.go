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
	"fmt"

	"github.com/padster/go-sound/cq"
)

// Generates the golden files. See test/sounds_test.go for actual test.
func main() {
	fmt.Printf("Building parameters\n")

	sampleRate := 44100.0
	minFreq, maxFreq, bpo := 100.0, sampleRate / 3.0, 24
	params := cq.NewCQParams(sampleRate, minFreq, maxFreq, bpo)

	fmt.Printf("Params = %v\n", params)
	fmt.Printf("Building Kernel... %v\n", params)

	kernel := cq.NewCQKernel(params)

	fmt.Printf("Done! Kernel = %v\n", kernel)
}

