// go run spectrogramshift.go --shift=<semitones> <file>
package main

import (
	"flag"
	"fmt"
	"math"
	"math/cmplx"
	"runtime"

	"github.com/padster/go-sound/cq"
	f "github.com/padster/go-sound/file"
	"github.com/padster/go-sound/output"
	s "github.com/padster/go-sound/sounds"
)

// Takes a spectrogram, applies a shift, inverts back and plays the result.
func main() {
	// Needs to be at least 2 when doing openGL + sound output at the same time.
	runtime.GOMAXPROCS(3)

	sampleRate := s.CyclesPerSecond
	octaves := flag.Int("octaves", 7, "Range in octaves")
	minFreq := flag.Float64("minFreq", 55.0, "Minimum frequency")
	semitones := flag.Int("shift", 0, "Semitones to shift")
	bpo := flag.Int("bpo", 24, "Buckets per octave")
	flag.Parse()

	remainingArgs := flag.Args()
	argCount := len(remainingArgs)
	if argCount < 1 || argCount > 2 {
		panic("Required: <input> [output] argument")
	}
	inputFile := remainingArgs[0]

	// Note: scale the output frequency by this to change pitch dilation into time dilation
	// shift := math.Pow(2.0, float64(-*semitones) / 12.0)

	paramsIn := cq.NewCQParams(sampleRate, *octaves, *minFreq, *bpo)
	paramsOut := cq.NewCQParams(sampleRate, *octaves, *minFreq, *bpo)

	spectrogram := cq.NewSpectrogram(paramsIn)
	cqInverse := cq.NewCQInverse(paramsOut)

	inputSound := f.Read(inputFile)
	inputSound.Start()
	defer inputSound.Stop()

	fmt.Printf("Running...\n")
	columns := spectrogram.ProcessChannel(inputSound.GetSamples())
	outColumns := shiftSpectrogram(
		*semitones*(*bpo/12), 0, flipSpectrogram(columns, *octaves, *bpo), *octaves, *bpo)
	soundChannel := cqInverse.ProcessChannel(outColumns)
	resultSound := s.WrapChannelAsSound(soundChannel)

	// HACK: Amplify for now.
	resultSound = s.MultiplyWithClip(resultSound, 3.0)

	if argCount == 2 {
		f.Write(resultSound, remainingArgs[1])
	} else {
		output.Play(resultSound)
	}
}

func shiftSpectrogram(binOffset int, sampleOffset int, samples <-chan []complex128, octaves int, bpo int) <-chan []complex128 {
	result := make(chan []complex128)

	go func() {
		ignoreSamples := sampleOffset
		at := 0
		for s := range samples {
			if ignoreSamples > 0 {
				ignoreSamples--
				continue
			}

			octaveCount := octaves
			if at > 0 {
				octaveCount = numOctaves(at)
				if octaveCount == octaves {
					at = 0
				}
			}
			at++

			toFill := octaveCount * bpo
			column := make([]complex128, toFill, toFill)

			// NOTE: Zero-padded, not the best...
			if binOffset >= 0 {
				copy(column, s[binOffset:])
			} else {
				copy(column[-binOffset:], s)
			}
			result <- column
		}
		close(result)
	}()
	return result
}

func numOctaves(at int) int {
	result := 1
	for at%2 == 0 {
		at /= 2
		result++
	}
	return result
}

func clone(values []complex128) []complex128 {
	result := make([]complex128, len(values), len(values))
	for i, v := range values {
		result[i] = v
	}
	return result
}

func flipSpectrogram(samples <-chan []complex128, octaves int, bpo int) <-chan []complex128 {
	result := make(chan []complex128)
	go func() {
		var phaseAt []float64 = nil
		for s := range samples {
			if phaseAt == nil {
				phaseAt = make([]float64, len(s), len(s))
			}
			for i, v := range s {
				vp := cmplx.Phase(v)
				vp = makeCloser(phaseAt[i], vp)
				phaseAt[i] = vp
			}

			newSample := make([]complex128, len(s), len(s))
			for i := 0; i < len(s); i++ {
				newSample[i] = s[i]
			}

			for i := 0; i < len(s); i++ {
				other := len(s) - 1 - i
				pFactor := float64(octaves) - float64(2*i+1)/float64(bpo)
				phase := phaseAt[other] / math.Pow(2.0, pFactor)
				newSample[i] = cmplx.Rect(cmplx.Abs(s[other]), phase)
			}

			result <- newSample
		}
		close(result)
	}()
	return result
}

// Return the closest number X to toShift, such that X mod Tau == modTwoPi
func makeCloser(toShift, modTau float64) float64 {
	if math.IsNaN(modTau) {
		modTau = 0.0
	}
	// Minimize |toShift - (modTau + tau * cyclesToAdd)|
	// toShift - modTau - tau * CTA = 0
	cyclesToAdd := (toShift - modTau) / cq.TAU
	return modTau + float64(cq.Round(cyclesToAdd))*cq.TAU
}
