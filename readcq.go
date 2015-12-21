package main

import (
	"flag"
	"fmt"
	"runtime"

	"github.com/padster/go-sound/cq"
	f "github.com/padster/go-sound/file"
	"github.com/padster/go-sound/output"
	s "github.com/padster/go-sound/sounds"
	"github.com/padster/go-sound/util"
)

const (
	SHOW_SPECTROGRAM = false
)

// Reads the CQ columns from file, converts back into a sound.
func main() {
	runtime.GOMAXPROCS(6)

	// Parse flags...
	sampleRate := s.CyclesPerSecond
	octaves := flag.Int("octaves", 7, "Range in octaves")
	minFreq := flag.Float64("minFreq", 55.0, "Minimum frequency")
	bpo := flag.Int("bpo", 24, "Buckets per octave")
	flag.Parse()

	remainingArgs := flag.Args()
	if len(remainingArgs) < 0 || len(remainingArgs) > 1 {
		panic("Only takes at most one arg: input file name.")
	}
	inputFile := "out.cq"
	if len(remainingArgs) == 1 {
		inputFile = remainingArgs[0]
	}

	params := cq.NewCQParams(sampleRate, *octaves, *minFreq, *bpo)

	if SHOW_SPECTROGRAM {
		cqChannel := f.ReadCQColumns(inputFile, params)
		spectrogram := cq.NewSpectrogram(params)
		columns := spectrogram.InterpolateCQChannel(cqChannel)
		toShow := util.NewSpectrogramScreen(882, *bpo**octaves, *bpo)
		toShow.Render(columns, 1)
	} else {
		asSound := f.ReadCQ(inputFile, params, false)
		fmt.Printf("Playing...\n")
		output.Play(asSound)
		fmt.Printf("Done...\n")
	}
}
