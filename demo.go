// Demo usage of the go-sound Sounds library.
package main

import (
	"github.com/padster/go-sound/sounds"
	"github.com/padster/go-sound/output"
)

func main() {
	sound := sounds.NewSineWave(440)
	renderer := output.NewScreen(1200, 300)
	renderer.Render(sound)
}
