// Plays a sound via pulseaudio
package output

import (
	"fmt"

	"github.com/moriyoshi/pulsego" // Requires slight modifications.
	"github.com/padster/go-sound/sounds"
)

func Play(s sounds.Sound) {
	pa := pulsego.NewPulseMainLoop()
	defer pa.Dispose()
	pa.Start()

	sync_ch := make(chan int)
	go playSamples(s, sync_ch, pa)
	<-sync_ch
}

func playSamples(s sounds.Sound, sync_ch chan int, pa *pulsego.PulseMainLoop) {
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
	st := ctx.NewStream("default", &pulsego.PulseSampleSpec{
		Format:   pulsego.SAMPLE_FLOAT32LE,
		Rate:     int(sounds.CyclesPerSecond()),
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
		// TODO - read add size from pulseAudio
		toAdd := 65536
		buffer := make([]float32, toAdd)

		// TODO - detect when the sample has finished.
		for i := range buffer {
			buffer[i] = float32(<-samples)
		}
		st.Write(buffer, pulsego.SEEK_RELATIVE)
	}
}
