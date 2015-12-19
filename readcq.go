package main

import (
  "flag"
  "runtime"

  "github.com/padster/go-sound/cq"
  f "github.com/padster/go-sound/file"
  s "github.com/padster/go-sound/sounds"
  "github.com/padster/go-sound/output"
)

// Reads the CQ columns from file, converts back into a sound.
func main() {
  runtime.GOMAXPROCS(6)

  // Parse flags...
  sampleRate := s.CyclesPerSecond
  octaves := flag.Int("octaves", 7, "Range in octaves")
  minFreq := flag.Float64("minFreq", 55.0, "Minimum frequency")
  bpo := flag.Int("bpo", 24, "Buckets per octave")
  zip := flag.Bool("zip", false, "Whether to unzip the input")
  flag.Parse()

  remainingArgs := flag.Args()
  if len(remainingArgs) < 0 || len(remainingArgs) > 1 {
    panic("Only takes at most one arg: input file name.")
  }
  inputFile := "out.cq"
  if len(remainingArgs) == 1 {
    inputFile = remainingArgs[0]
  }
  
  params := cq.NewCQParams(sampleRate, *octaves, *minFreq, *bpo)
  asSound := f.ReadCQ(inputFile, params, *zip)

  asSound.Start()
  defer asSound.Stop()
  output.Play(asSound)
}
