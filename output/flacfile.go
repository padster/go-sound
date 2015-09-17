// Write a sound to a .wav file
package output

import (
  "fmt"
  "os"

  goflac "github.com/cocoonlife/goflac"
  s "github.com/padster/go-sound/sounds"
)

// WriteSoundToFlac creates a file at a path, and writes the given sound in the .flac format.
func WriteSoundToFlac(sound s.Sound, path string) error {
  if _, err := os.Stat(path); err == nil {
    return os.ErrExist
  }

  sampleRate := int(s.CyclesPerSecond)
  depth := 24

  fileWriter, err := goflac.NewEncoder(path, 1, depth, sampleRate)
  if err != nil {
    fmt.Printf("Error opening file to write to! %v\n", err)
    panic("Can't write file")
  }
  defer fileWriter.Close()

  // Starts the sound, and accesses its sample stream.
  sound.Start()
  samples := sound.GetSamples()
  defer sound.Stop()

  // TODO: Make a common utility for this, it's used here and in both CQ and CQI.
  frameSize := 44100
  buffer := make([]float64, frameSize, frameSize)
  at := 0 
  for s := range samples {
    if at == frameSize {
      writeFrame(fileWriter, buffer)
      at = 0;
    }
    buffer[at] = s
    at++
  }
  writeFrame(fileWriter, buffer[:at])

  return nil;
}

func writeFrame(fileWriter *goflac.Encoder, samples []float64) {
  n := len(samples)

  frameBuffer := make([]int32, n, n)
  for i, v := range samples {
    frameBuffer[i] = intFromFloatWithDepth(v, fileWriter.Depth)
  }

  frame := goflac.Frame{
    1,                  /* channels */
    fileWriter.Depth,   /* depth */
    fileWriter.Rate,    /* rate */
    frameBuffer,
  }

  if err := fileWriter.WriteFrame(frame); err != nil {
    fmt.Printf("Error writing frame to file :( %v\n", err)
    panic("Can't write to file")
  }
}

// HACK - should share with cq/utils.go
func intFromFloatWithDepth(input float64, depth int) int32 {
  return int32(input * (float64(unsafeShift(depth)) - 1.0))
}
func unsafeShift(s int) int {
  return 1 << uint(s)
}
