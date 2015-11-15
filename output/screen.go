// Renders a sound wave to screen.
package output

import (
	"github.com/padster/go-sound/sounds"
	"github.com/padster/go-sound/util"
)

// Render starts rendering a sound wave's samples to a screen of given height.
func Render(s sounds.Sound, width int, height int, samplesPerPixel int) {
	screen := util.NewScreen(width, height, samplesPerPixel)
	s.Start()
	// TODO(padster): Currently this generates and renders the samples
	// as quickly as possible. Better instead to render close to realtime?
	screen.Render(s.GetSamples(), 2)
}
