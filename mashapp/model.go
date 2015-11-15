package mashapp

import (
	b64 "encoding/base64"
	"encoding/binary"
	"math"
)

type InputMeta struct {
	ID    int    `json:"id,string"`
	Path  string `json:"path"`
	Muted bool   `json:"muted"`

	// Init -> Final length in samples
	OriginalLength int `json:"originalLength,string"`
	FinalLength    int `json:"finalLength,string"`

	// Init -> Final pitch in semitones
	OriginalPitch int `json:"originalPitch,string"`
	FinalPitch    int `json:"finalPitch,string"`
}

type Block struct {
	ID      int    `json:"id"`
	InputID int    `json:"inputId"`
	Name    string `json:"name"`

	StartSample int `json:"startSample"`
	EndSample   int `json:"endSample"`

	// TODO - use?
	Selected bool `json:"selected"`
}

type OutputMeta struct {
	ID      int `json:"id,string"`
	BlockID int `json:"blockID,string"`

	Line  int  `json:"line,string"`
	Muted bool `json:"muted"`

	// NOTE: TimeShift if Duration() != Duration(Block)
	StartSample int `json:"startSample,string"`
	EndSample   int `json:"endSample,string"`

	// TODO - use?
	Selected bool `json:"selected"`

	Changes []Modification `json:"changes"`
}

type Modification struct {
	Type   int       `json:"type,string"`
	Params []float64 `json:"params"`
}

type GoSamples []float64
type JsonSamples string

// UTIL
func floatsToBase64(values GoSamples) JsonSamples {
	asFloats := ([]float64)(values)
	return JsonSamples(bytesToBase64(floatsToBytes(asFloats)))
}

func floatsToBytes(values []float64) []byte {
	bytes := make([]byte, 0, 4*len(values))
	for _, v := range values {
		bytes = append(bytes, float32ToBytes(float32(v))...)
	}
	return bytes
}

func float32ToBytes(value float32) []byte {
	bits := math.Float32bits(value)
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, bits)
	return bytes
}

func bytesToBase64(values []byte) string {
	return b64.StdEncoding.EncodeToString(values)
}
