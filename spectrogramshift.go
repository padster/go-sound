// go run spectrogramshift.go --shift=<semitones> <file>
package main

import (
  "flag"
  "fmt"
  // "math"
  "runtime"

  "github.com/padster/go-sound/cq"
  f "github.com/padster/go-sound/file"
  s "github.com/padster/go-sound/sounds"
  "github.com/padster/go-sound/output"
)

// Takes a spectrogram, applies a shift, inverts back and plays the result.
func main() {
  // Needs to be at least 2 when doing openGL + sound output at the same time.
  runtime.GOMAXPROCS(3)

  sampleRate := s.CyclesPerSecond
  octaves := 7
  minFreq := flag.Float64("minFreq", 55.0, "minimum frequency")
  maxFreq := flag.Float64("maxFreq", 55.0 * float64(cq.UnsafeShift(octaves)), "maximum frequency")
  semitones := flag.Int("shift", 0, "Semitones to shift")
  bpo := flag.Int("bpo", 48, "Buckets per octave")
  flag.Parse()

  remainingArgs := flag.Args()
  argCount := len(remainingArgs)
  if argCount < 1 || argCount > 2 {
    panic("Required: <input> [output] argument")
  }
  inputFile := remainingArgs[0]

  // Note: scale the output frequency by this to change pitch dilation into time dilation
  // shift := math.Pow(2.0, float64(-*semitones) / 12.0)  

  paramsIn := cq.NewCQParams(sampleRate, *minFreq, *maxFreq, *bpo)
  paramsOut := cq.NewCQParams(sampleRate, *minFreq, *maxFreq, *bpo)

  spectrogram := cq.NewSpectrogram(paramsIn)
  cqInverse := cq.NewCQInverse(paramsOut)

  inputSound := f.Read(inputFile)
  inputSound.Start()
  defer inputSound.Stop()

  fmt.Printf("Running...\n")
  columns := spectrogram.ProcessChannel(inputSound.GetSamples())
  outColumns := shiftSpectrogram(*semitones * (*bpo / 12), 11, columns, octaves, *bpo)
  soundChannel := cqInverse.ProcessChannel(outColumns)
  resultSound := s.WrapChannelAsSound(soundChannel)

  // HACK: Amplify for now.
  resultSound = s.MultiplyWithClip(resultSound, 2.0)

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
  for at % 2 == 0 {
    at /= 2
    result++
  }
  return result
}
