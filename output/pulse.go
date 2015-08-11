// Package output it responsible for preparing the sounds for human consumption,
// whether audio, visual or other means.
package output

import (
	"fmt"

	"github.com/padster/go-sound/sounds"
)

// Play plays a sound to audio out via pulseaudio.
func Play(s sounds.Sound) {
	pa := NewPulseMainLoop()
	defer pa.Dispose()
	pa.Start()

	sync_ch := make(chan int)
	go playSamples(s, sync_ch, pa)
	<-sync_ch
}

// playSamples handles the writing of a sound's channel of samples to a pulse stream.
func playSamples(s sounds.Sound, sync_ch chan int, pa *PulseMainLoop) {
	defer func() {
		sync_ch <- 0
	}()

	// Create a pulse audio context to play the sound.
	ctx := pa.NewContext("default", 0)
	if ctx == nil {
		fmt.Println("Failed to create a new context")
		return
	}
	defer ctx.Dispose()

	// Create a single-channel pulse audio stream to write the sound to.
	st := ctx.NewStream("default", &PulseSampleSpec{
		Format:   SAMPLE_FLOAT32LE,
		Rate:     int(sounds.CyclesPerSecond),
		Channels: 1,
	})
	if st == nil {
		fmt.Println("Failed to create a new stream")
		return
	}
	defer st.Dispose()

	// Starts the sound, and accesses its sample stream.
	s.Start()
	samples := s.GetSamples()
	defer s.Stop()

	// Continually buffers data from the stream and writes to audio.
	st.ConnectToSink()
	for {
		toAdd := st.WritableSize()
		if toAdd == 0 {
			continue
		}

		// No buffer - write immediately.
		// TODO(padster): Play with this to see if chunked writes actually reduce delay.
		if toAdd > 1 {
			toAdd = 1
		}

		buffer := make([]float32, toAdd)
		finishedAt := toAdd

		for i := range buffer {
			sample, stream_ok := <-samples
			if !stream_ok {
				finishedAt = i
				break
			}
			buffer[i] = float32(sample)
		}
		if finishedAt == 0 {
			st.Drain()
			return
		}
		st.Write(buffer[0:finishedAt], SEEK_RELATIVE)
	}
}
