package soundfile

import (
	"strings"

	o "github.com/padster/go-sound/output"
	s "github.com/padster/go-sound/sounds"
)

func Read(path string) s.Sound {
	switch {
	case strings.HasSuffix(path, ".flac"):
		return s.LoadFlacAsSound(path)
	case strings.HasSuffix(path, ".wav"):
		return s.LoadWavAsSound(path, 0 /* TODO: average all channels. */)
	default:
		panic("Unsupported file type: " + path)
	}
}

func Write(sound s.Sound, path string) {
	switch {
	case strings.HasSuffix(path, ".flac"):
		panic("FLAC support currently broken, please use something else")
		o.WriteSoundToFlac(sound, path)
	case strings.HasSuffix(path, ".wav"):
		o.WriteSoundToWav(sound, path)
	default:
		panic("Unsupported file type: " + path)
	}
}
