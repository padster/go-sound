// Converts a .wav file into a Sound.
package sounds

import (
	"math"
	"os"

	wav "github.com/cryptix/wav"
)

const (
	normScale = float64(1) / float64(math.MaxInt16)
)

type WavFileSound struct {
	samples chan float64
	path    string
	channel uint16

	wavReader  *wav.Reader
	meta       wav.File
	durationMs uint64

	samplesLeft uint32
	running     bool
}

// NewSineWave loads a wav file and turns a particular channel into a Sound.
func LoadWavAsSound(path string, channel uint16) *WavFileSound {
	wavReader := loadReaderOrPanic(path)

	meta := wavReader.GetFile()
	if meta.Channels <= channel {
		panic("Unsupported channel number")
	}
	if meta.SampleRate != uint32(CyclesPerSecond()) {
		panic("TODO: Support wav files that aren't 44.1kHz")
	}
	durationMs := uint64(1000.0 * float64(wavReader.GetSampleCount()) / float64(meta.SampleRate))

	ret := WavFileSound{
		make(chan float64),
		path,
		channel,
		wavReader,
		meta,
		durationMs,
		wavReader.GetSampleCount(),
		false, /* running */
	}
	return &ret
}

func (s *WavFileSound) GetSamples() <-chan float64 {
	return s.samples
}

func (s *WavFileSound) DurationMs() uint64 {
	return s.durationMs
}

func (s *WavFileSound) Start() {
	s.running = true

	go func() {
		for s.running && s.samplesLeft > 0 {
			// Read all channels, but pick just the one we want.
			selected := float64(0)
			for i := uint16(0); i < s.meta.Channels; i++ {
				n, err := s.wavReader.ReadSample()
				if err != nil {
					s.running = false
					break
				}
				if i == s.channel {
					selected = float64(int16(n)) * normScale
				}
			}

			if !s.running {
				break
			}
			s.samples <- selected
			s.samplesLeft--
		}
		s.Stop()
		close(s.samples)
	}()
}

func (s *WavFileSound) Stop() {
	s.running = false
}

func (s *WavFileSound) Reset() {
	if s.running {
		panic("Stop before reset!")
	}

	s.samples = make(chan float64)
	s.wavReader = loadReaderOrPanic(s.path)
	s.meta = s.wavReader.GetFile()
	s.samplesLeft = s.wavReader.GetSampleCount()
	s.running = true
}

// Utility to handle failure cases of reading an input file.
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
