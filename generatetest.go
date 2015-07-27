package main 

import (
  "fmt"

  "github.com/padster/go-sound/sounds"
  "github.com/padster/go-sound/test"
  "github.com/padster/go-sound/output"
)

// Generates the golden files. See test/sounds_test.go for actual test.
func main() {
  generate("test/timed.wav", test.SampleTimedSound())
  generate("test/silence.wav", test.SampleSilence())
  generate("test/concat.wav", test.SampleConcat())
  generate("test/normalsum.wav", test.SampleNormalSum())
  generate("test/multiply.wav", test.SampleMultiply())
  generate("test/repeat.wav", test.SampleRepeater())
  generate("test/adsr.wav", test.SampleAdsrEnvelope())
  generate("test/sampler.wav", test.SampleSampler())
}

func generate(path string, sound sounds.Sound) {
  fmt.Printf("Generating sound at %s...\n", path)
  if err := output.WriteSoundToWav(sound, path); err != nil {
    fmt.Printf("  Skipped %s, path exists. Delete to regenerate.\n", path)
  }
}
