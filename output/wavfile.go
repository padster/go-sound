// Write a sound to a .wav file
package output

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"

	wav "github.com/cryptix/wav"
	"github.com/padster/go-sound/sounds"
)

const (
	normScale = float64(math.MaxInt16)
)

func WriteSoundToWav(s sounds.Sound, path string) error {
	// Create file first, only if it doesn't exist:
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("Can't write to %s, file already exists\n", path)
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
		SampleRate:      uint32(sounds.CyclesPerSecond()),
		Channels:        1,
		SignificantBits: 16,
	}
	writer, err := wf.NewWriter(file)
	if err != nil {
		return err
	}
	defer writer.Close()

	// Starts the sound, and accesses its sample stream.
	fmt.Printf("Writing sound for %d ms\n", s.DurationMs())
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
