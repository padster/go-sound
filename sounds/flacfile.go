package sounds

import (
  "fmt"
  "io"
  "os"

  flac "github.com/cocoonlife/goflac"
)

// A flacFileSound is parameters to the algorithm that converts a channel from a .flac file into a sound.
type flacFileSound struct {
  path    string
  fileReader *flac.Decoder
}

// LoadFlacAsSound loads a .flac file and converts the average of its channels to a Sound.
//
// For example, to read the first channel from a local file at 'piano.flac':
//  sounds.LoadFlacAsSound("piano.flac")
func LoadFlacAsSound(path string) Sound {
  flacReader := loadFlacReaderOrPanic(path)

  if flacReader.Rate != int(CyclesPerSecond) {
    // TODO(padster): Support more if there's a need.
    panic("Only wav files that are 44.1kHz are supported.")
  }

  // TODO: Precalculate the duration properly.
  durationMs := MaxLength

  data := flacFileSound{
    path,
    flacReader,
  }

  return NewBaseSound(&data, durationMs)
}

// Run generates the samples by extracting them out of the .flac file.
func (s *flacFileSound) Run(base *BaseSound) {
  frame, err := s.fileReader.ReadFrame()
  for err != io.EOF {
    count := len(frame.Buffer) / frame.Channels

    for i := 0; i < count; i++ {
      v := 0.0
      for _, c := range frame.Buffer[i*frame.Channels : (i+1)*frame.Channels] {
        v += floatFromBitWithDepth(c, frame.Depth)
      }
      v = v / float64(frame.Channels)
      if !base.WriteSample(v) {
        err = io.EOF
      }
    }

    if err != io.EOF {
      frame, err = s.fileReader.ReadFrame()
    }
  }
}

// Stop cleans up this sound, closing the reader.
func (s *flacFileSound) Stop() {
  s.fileReader.Close()
}

// Reset reopens the file from the start.
func (s *flacFileSound) Reset() {
  s.Stop()
  s.fileReader = loadFlacReaderOrPanic(s.path)
}

// String returns the textual representation
func (s *flacFileSound) String() string {
  return fmt.Sprintf("Flac[path %s]", s.path)
}

// loadFlacReaderOrPanic reads a flac file and handles failure cases.
func loadFlacReaderOrPanic(path string) *flac.Decoder {
  _, err := os.Stat(path)
  if err != nil {
    panic(err)
  }
  fileReader, err := flac.NewDecoder(path)
  if err != nil {
    panic(err)
  }
  return fileReader
}

// HACK - should share with cq/utils.go
func floatFromBitWithDepth(input int32, depth int) float64 {
  return float64(input) / (float64(unsafeShift(depth)) - 1.0) // Hmmm..doesn't seem right?
}
func intFromFloatWithDepth(input float64, depth int) int32 {
  return int32(input * (float64(unsafeShift(depth)) - 1.0))
}
func unsafeShift(s int) int {
  return 1 << uint(s)
}

