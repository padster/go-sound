package features

import (
    "bytes"
    "fmt"
    "io/ioutil"

    "github.com/padster/go-sound/cq"
)

// PeakDetector takes the constant Q output, and for each sample bin, returns 
// whether that sample is a 'peak' in the music.
type PeakDetector struct {

}

func (pd *PeakDetector) ProcessChannel(samples <-chan []complex128) <-chan []byte {
    result := make(chan []byte)

    go func() {
        i := 0
        for sample := range samples {
            if i % 10 == 0 {
                fmt.Printf("Writing peaks %d\n", i)
            }
            i++
            result <- pd.processColumn(sample)
        }
    }()

    return result
}

func (pd *PeakDetector) processColumn(column []complex128) []byte {
    size := len(column)
    result := make([]byte, size, size)

    // TODO
    return result
}

func WritePeaks(outputFile string, peaks <-chan []byte) {
    ioutil.WriteFile(outputFile, PeaksToBytes(peaks), 0644)
}

func PeaksToBytes(peaks <-chan []byte) []byte {
    outputBuffer := bytes.NewBuffer(make([]byte, 0, 1024))
    width, height := 0, 0
    for col := range peaks {
        for _, c := range col {
            cq.WriteByte(outputBuffer, c)
        }
        if width%100 == 0 {
            fmt.Printf("At frame: %d\n", width)
        }
        width++
        height = len(col)
    }
    fmt.Printf("Done! - %d by %d\n", width, height)
    return outputBuffer.Bytes()
}