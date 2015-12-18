// Write sounds to audio output via Jack (http://jackaudio.org)
package output

import (
	"fmt"

	"github.com/xthexder/go-jack"

	s "github.com/padster/go-sound/sounds"
)

type jackContext struct {
	sound         s.Sound
	sampleChannel <-chan float64
	leftPort      *jack.Port
	rightPort     *jack.Port
	running       bool
}

// Play plays a sound to audio out via jack.
func PlayJack(sound s.Sound) {
	player := &jackContext{
		sound,
		nil,   /* sampleChannel */
		nil,   /* leftPort */
		nil,   /* rightPort */
		false, /* running */
	}

	// Setup copied from https://github.com/xthexder/go-jack readme.
	client, _ := jack.ClientOpen("GoSoundOut", jack.NoStartServer)
	if client == nil {
		fmt.Println("Could not connect to jack server.")
		return
	}
	defer client.Close()

	if code := client.SetProcessCallback(player.process); code != 0 {
		fmt.Println("Failed to set process callback.")
		return
	}
	client.OnShutdown(player.shutdown)

	sound.Start()
	defer player.sound.Stop()

	player.sampleChannel = sound.GetSamples()
	player.running = true

	if code := client.Activate(); code != 0 {
		fmt.Println("Failed to activate client.")
		return
	}

	player.leftPort = client.PortRegister("go-sound-left", jack.DEFAULT_AUDIO_TYPE, jack.PortIsOutput, 0)
	player.rightPort = client.PortRegister("go-sound-right", jack.DEFAULT_AUDIO_TYPE, jack.PortIsOutput, 0)
	for player.running {
	}
}

func (j *jackContext) process(nframes uint32) int {
	leftSamples := j.leftPort.GetBuffer(nframes)
	rightSamples := j.rightPort.GetBuffer(nframes)

	// fmt.Printf("Writing %d samples\n", len(samples))
	for i := range leftSamples {
		sample, stream_ok := <-j.sampleChannel
		if !stream_ok {
			j.running = false
			return 1
		}
		leftSamples[i] = jack.AudioSample(sample)
		rightSamples[i] = jack.AudioSample(sample)
	}
	return 0
}

func (j *jackContext) shutdown() {
}
