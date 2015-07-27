package test 

// go test github.com/padster/go-sound/test

import (
  "bytes"
  "io/ioutil"
  "os"
  "testing"

  "github.com/padster/go-sound/sounds"
  "github.com/padster/go-sound/output"
)

// Generates multiple sample sounds, and compares against golden files generated in generatetest.go

func TestTimed(t *testing.T) {
  compareFile(t, "timed.wav", SampleTimedSound())
}

func TestSilence(t *testing.T) {
  compareFile(t, "silence.wav", SampleSilence())
}

func TestConcat(t *testing.T) {
  compareFile(t, "concat.wav", SampleConcat())
}

func TestNormalSum(t *testing.T) {
  compareFile(t, "normalsum.wav", SampleNormalSum())
}

func TestMultiply(t *testing.T) {
  compareFile(t, "multiply.wav", SampleMultiply())
}

func TestRepeater(t *testing.T) {
  compareFile(t, "repeat.wav", SampleRepeater())
}

func TestAdsrEnvelope(t *testing.T) {
  compareFile(t, "adsr.wav", SampleAdsrEnvelope())
}

func TestSampler(t *testing.T) {
  compareFile(t, "sampler.wav", SampleSampler())
}

func TestDelay(t *testing.T) {
  compareFile(t, "delay.wav", SampleAddDelay())
}

// TODO(padster): Add tests for util/parser.go

// compareFile writes a sound to file, compares it to a golden file,
// and fails the test if anything goes wrong.
func compareFile(t *testing.T, path string, sound sounds.Sound) {
  f, err := ioutil.TempFile("", "tmp_")
  os.Remove(f.Name())
  t.Logf("Writing sound to %s\n", f.Name())

  if err != nil {
    t.Fatalf("Error creating temp file: %s\n", err)
    return
  }
  if err = output.WriteSoundToWav(sound, f.Name()); err != nil {
    t.Fatalf("Error writing to temp file: %s\n", err)
    return
  }
  defer os.Remove(f.Name())

  bE, errE := ioutil.ReadFile(path)
  if errE != nil {
    t.Fatalf("Error reading golden file: %s\n", errE)
    return
  }

  bA, errA := ioutil.ReadFile(f.Name())
  if errA != nil {
    t.Fatalf("Error reading temp file: %s\n", errA)
    return
  }

  if !bytes.Equal(bE, bA) {
    t.Fatalf("File does not match %s! Please listen to the new version.\n", path)
    return
  }
}