package sounds

import (
	"fmt"
	"math"
	"os"

	wav "github.com/cryptix/wav"
)

const (
	normScale = float64(1) / float64(math.MaxInt16)
)

// A wavFileSound is parameters to the algorithm that converts a channel from a .wav file into a sound.
type wavFileSound struct {
	path    string
	channel uint16

	// TODO(padster): Clean this up, sounds shouldn't have mutable state beyond child sounds.
	// samplesLeft can be removed, the other two may require edits to the wav library.
	wavReader   *wav.Reader
	meta        wav.File
	samplesLeft uint32
}

// LoadWavAsSound loads a .wav file and converts one of its channels into a Sound.
//
// For example, to read the first channel from a local file at 'piano.wav':
//	sounds.LoadWavAsSound("piano.wav", 0)
func LoadWavAsSound(path string, channel uint16) Sound {
	wavReader := loadReaderOrPanic(path)

	meta := wavReader.GetFile()
	if meta.Channels <= channel {
		panic("Unsupported channel number.")
	}
	if meta.SampleRate != uint32(CyclesPerSecond) {
		// TODO(padster): Support more if there's a need.
		panic("Only wav files that are 44.1kHz are supported.")
	}
	durationMs := uint64(1000.0 * float64(wavReader.GetSampleCount()) / float64(meta.SampleRate))

	data := wavFileSound{
		path,
		channel,
		wavReader,
		meta,
		wavReader.GetSampleCount(), /* samplesLeft */
	}

	return NewBaseSound(&data, durationMs)
}

// Run generates the samples by extracting them out of the .wav file.
func (s *wavFileSound) Run(base *BaseSound) {
	for s.samplesLeft > 0 {
		// Read all channels, but pick just the one we want.
		selected := float64(0)
		for i := uint16(0); i < s.meta.Channels; i++ {
			n, err := s.wavReader.ReadSample()
			if err != nil {
				base.Stop()
				break
			}
			if i == s.channel {
				// Need this to convert the 16-bit integer into a [-1, 1] float sample.
				selected = float64(int16(n)) * normScale
			}
		}

		if !base.WriteSample(selected) {
			break
		}
		s.samplesLeft--
	}
}

// Stop cleans up this sound, in this case doing nothing.
func (s *wavFileSound) Stop() {
	// NOTE: It seems like the reader and file API have no Close cleanup.
}

// Reset reopens the file from the start.
func (s *wavFileSound) Reset() {
	s.wavReader = loadReaderOrPanic(s.path)
	s.meta = s.wavReader.GetFile()
	s.samplesLeft = s.wavReader.GetSampleCount()
}

// String returns the textual representation
func (s *wavFileSound) String() string {
	return fmt.Sprintf("Wav[channel %d from path %s]", s.channel, s.path)
}

// loadReaderOrPanic reads a wav file and handles failure cases.
func loadReaderOrPanic(path string) *wav.Reader {
	testInfo, err := os.Stat(path)
	if err != nil {
		panic(err)
	}
	testWav, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	result, err := wav.NewReader(testWav, testInfo.Size())
	if err != nil {
		panic(err)
	}
	return result
}
