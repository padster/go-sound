package soundfile

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/padster/go-sound/cq"
)

// Writes the result of a constant Q transform to file.
func WriteColumns(outputFile string, columns <-chan []complex128) {
	ioutil.WriteFile(outputFile, ColumnsToBytes(columns), 0644)
}

// Converts the result of a constant Q transform to a byte stream.
func ColumnsToBytes(columns <-chan []complex128) []byte {
	outputBuffer := bytes.NewBuffer(make([]byte, 0, 1024))
	width, height := 0, 0
	for col := range columns {
		for _, c := range col {
			cq.WriteComplex(outputBuffer, c)
		}
		if width%10000 == 0 {
			fmt.Printf("At frame: %d\n", width)
		}
		width++
		height = len(col)
	}
	fmt.Printf("Done! - %d by %d\n", width, height)
	return outputBuffer.Bytes()
}

// Reads a file and converts back into a CQ channel.
func ReadCQColumns(inputFile string, params cq.CQParams) <-chan []complex128 {
	loaded, err := ioutil.ReadFile(inputFile)
	if err != nil {
		panic("Can't load file " + inputFile)
	}
	complexEntries := len(loaded) / 8 // complex stored as two float32s.
	fmt.Printf("Reading %d entries\n", complexEntries)

	asReader := bytes.NewReader(loaded)

	result := make(chan []complex128)
	go func() {
		heightGen := cq.GenerateHeights(params.Octaves)
		for at := 0; at < complexEntries; {
			nextSize := heightGen() * params.BinsPerOctave
			result <- cq.ReadComplexArray(asReader, nextSize)
			at += nextSize
		}
		close(result)
	}()
	return result
}
