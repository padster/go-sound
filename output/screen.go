// Renders a sound wave to screen.
package output

import (
	"github.com/padster/go-muse/ui"
	"github.com/padster/go-sound/sounds"
)

type Screen struct {
	screen *ui.Screen
}

// NewScreen creates a new output screen of a given size.
func NewScreen(width int, height int) *Screen {
	ret := Screen{
		ui.NewScreen(width, height, 15),
	}
	return &ret
}

// Render starts rendering a sound wave's samples to screen.
func (ui *Screen) Render(s sounds.Sound) {
	s.Start()
	ui.screen.Render(s.GetSamples(), 2)
}
