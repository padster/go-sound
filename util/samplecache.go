package util

import (
    s "github.com/padster/go-sound/sounds"
)

const (
    LOAD_LIMIT = int(s.CyclesPerSecond * 10 * 60) /* 10 minutes */
)

func CacheSamples(sound s.Sound) []float64 {
    var result []float64

    sound.Start()
    for sample := range sound.GetSamples() {
        result = append(result, sample)
        if len(result) == LOAD_LIMIT {
            break
        }
    }
    sound.Stop()

    return result
}