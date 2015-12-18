package features

import (
    "fmt" 
)

// PeakDetector takes the constant Q output, and for each sample bin, returns 
// whether that sample is a 'peak' in the music.
type PeakDetector struct {

}

func (pd *PeakDetector) ProcessChannel(samples <-chan []complex128) <-chan []bool {
    result := make(chan []bool)

    for sample := range samples {
        result <- pd.processColumn(sample)
    }

    return result
}

func (pd *PeakDetector) processColumn(column []complex128) []bool {
    size := len(column)
    result := make([]bool, size, size)

    // TODO
    return result
}