// Write a sound to a .wav file
package output

import (
	"encoding/binary"
	"math"
	"os"

	wav "github.com/cryptix/wav"
	"github.com/padster/go-sound/sounds"
)

const (
	normScale = float64(math.MaxInt16)
)

// WriteSoundToWav creates a file at a path, and writes the given sound in the .wav format.
func WriteSoundToWav(s sounds.Sound, path string) error {
	// Create file first, only if it doesn't exist:
	if _, err := os.Stat(path); err == nil {
		panic("File already exists! Please delete first")
		return os.ErrExist
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		file.Close()
	}()

	// Create a .wav writer for the file
	var wf = wav.File{
		SampleRate:      uint32(sounds.CyclesPerSecond),
		Channels:        1,
		SignificantBits: 16,
	}
	writer, err := wf.NewWriter(file)
	if err != nil {
		return err
	}
	defer writer.Close()

	// Starts the sound, and accesses its sample stream.
	s.Start()
	samples := s.GetSamples()
	defer s.Stop()

	// Write a single sample at a time, as per the .wav writer API.
	b := make([]byte, 2)
	for sample := range samples {
		toNumber := uint16(sample * normScale) // Inverse the read scaling
		binary.LittleEndian.PutUint16(b, uint16(toNumber))
		writer.WriteSample(b)
	}
	return nil
}
