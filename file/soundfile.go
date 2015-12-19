package soundfile

import (
	"fmt"
	"strings"

  	"github.com/padster/go-sound/cq"
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

func ReadCQ(path string, params cq.CQParams, zip bool) s.Sound {
	if zip {
		// TODO: Implement
		panic("Zip read CQ unsupported for now.")
	}

	fmt.Printf("Reading columms from %s\n", path)
	cqChannel := ReadCQColumns(path, params)
	inverse := cq.NewCQInverse(params)
	invChannel := inverse.ProcessChannel(cqChannel)
	return s.WrapChannelAsSound(invChannel)
}