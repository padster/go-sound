package mashapp

import (
    // "fmt"

    "github.com/padster/go-sound/file"
    // s "github.com/padster/go-sound/sounds"
    "github.com/padster/go-sound/util"
)

type ServerState struct {
    nextId int
    inputs map[int]InputSound
}

type InputSound struct {
    samples []float64
}

func NewServerState() *ServerState {
    result := ServerState {
        0,
        make(map[int]InputSound),
    }
    return &result
}

func (state *ServerState) loadSound(path string) (int, InputSound) {
    // TODO - lock
    id := state.nextId
    state.nextId++

    loadedSound := soundfile.Read(path)
    inputSound := InputSound {
        util.CacheSamples(loadedSound),
    }
    return id, inputSound
}
