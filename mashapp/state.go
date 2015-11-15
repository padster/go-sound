package mashapp

import (
	// "fmt"

	"github.com/padster/go-sound/cq"
	"github.com/padster/go-sound/file"
	// "github.com/padster/go-sound/output"
	s "github.com/padster/go-sound/sounds"
	"github.com/padster/go-sound/util"
)

const (
	SAMPLE_RATE       = s.CyclesPerSecond
	MIN_FREQ          = 55.0
	OCTAVES           = 7
	MAX_FREQ          = 55.0 * (1 << OCTAVES)
	BINS_PER_SEMITONE = 4
	BPO               = 12 * BINS_PER_SEMITONE
)

type ServerState struct {
	nextId int
	inputs map[int]InputSound
	blocks map[int]Block
}

// HACK - remove, replace with structs from model.go
type InputSound struct {
	samples GoSamples
}

func NewServerState() *ServerState {
	result := ServerState{
		0,
		make(map[int]InputSound),
		make(map[int]Block),
	}
	return &result
}

func (state *ServerState) loadSound(path string) (int, InputSound) {
	// TODO - lock
	id := state.nextId
	state.nextId++

	loadedSound := soundfile.Read(path)
	inputSound := InputSound{
		util.CacheSamples(loadedSound),
	}
	state.inputs[id] = inputSound
	return id, inputSound
}

func (state *ServerState) createBlock(block Block) Block {
	block.ID = state.nextId
	state.nextId++
	return block
}

// Pitch- and/or Time-shift an input, return the resulting samples
func (state *ServerState) shiftInput(input InputMeta) InputSound {
	beforeSamples := state.inputs[input.ID].samples
	beforeSound := s.WrapSliceAsSound(beforeSamples)
	beforeSound.Start()

	// HACK: Only pitch shift for now.
	paramsIn := cq.NewCQParams(SAMPLE_RATE, MIN_FREQ, MAX_FREQ, BPO)
	paramsOut := cq.NewCQParams(SAMPLE_RATE, MIN_FREQ, MAX_FREQ, BPO)
	spectrogram := cq.NewSpectrogram(paramsIn)
	cqInverse := cq.NewCQInverse(paramsOut)

	columns := spectrogram.ProcessChannel(beforeSound.GetSamples())
	outColumns := shiftSpectrogram(input.FinalPitch*(BINS_PER_SEMITONE), 0, columns, OCTAVES, BPO)
	soundChannel := cqInverse.ProcessChannel(outColumns)
	resultSound := s.WrapChannelAsSound(soundChannel)

	afterSamples := util.CacheSamples(resultSound)

	// TODO - update all blocks created off this input.
	return InputSound{
		afterSamples,
	}
}

func shiftSpectrogram(binOffset int, sampleOffset int, samples <-chan []complex128, octaves int, bpo int) <-chan []complex128 {
	result := make(chan []complex128)

	go func() {
		sRead, sWrite := 0, 0

		ignoreSamples := sampleOffset
		at := 0
		for s := range samples {
			sRead++
			if ignoreSamples > 0 {
				ignoreSamples--
				continue
			}

			octaveCount := octaves
			if at > 0 {
				octaveCount = numOctaves(at)
				if octaveCount == octaves {
					at = 0
				}
			}
			at++

			toFill := octaveCount * bpo
			column := make([]complex128, toFill, toFill)

			// NOTE: Zero-padded, not the best...
			if binOffset >= 0 {
				copy(column, s[binOffset:])
			} else {
				copy(column[-binOffset:], s)
			}
			sWrite++
			result <- column
		}
		close(result)
	}()
	return result
}

func numOctaves(at int) int {
	result := 1
	for at%2 == 0 {
		at /= 2
		result++
	}
	return result
}
