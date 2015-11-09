package mashapp

import (
)

type InputMeta struct {
    ID int `json:"id"`
    Path string `json:"path"`
    Muted bool `json:"muted"`

    // Init -> Final length in samples
    OriginalLength int `json:"originalLength"`
    FinalLength int `json:"finalLength"`
    
    // Init -> Final pitch in semitones
    OriginalPitch int `json:"originalPitch"`
    FinalPitch int `json:"finalPitch"`
}

type Block struct {
    ID int `json:"id"`
    InputID int `json:"inputId"`
    Name string `json:"name"`

    StartSample int `json:"startSample"`
    EndSample int `json:"endSample"`

    // TODO - use?
    Selected bool `json:"selected"`
}

type OutputMeta struct {
    ID int `json:"id"`
    BlockID int `json:"blockID"`

    Line int `json:"line"`
    Muted bool `json:"muted"`

    // NOTE: TimeShift if Duration() != Duration(Block)
    StartSample int `json:"startSample"`
    EndSample int `json:"endSample"`

    // TODO - use?
    Selected bool `json:"selected"`

    Changes []Modification `json:"changes"`
}

type GoSamples []float64
type JsonSamples string

type Modification struct {
    Type int `json:"type"`
    Params []float64 `json:"params"`
}

