// Renders various data from a channel of [-1, 1] onto screen.
package util

import (
	"log"
	"runtime"

	"github.com/padster/go-sound/types"

	// TODO(padster) - migrate to core, not compat.
	gl "github.com/go-gl/gl/v3.3-compatibility/gl"
	glfw "github.com/go-gl/glfw/v3.1/glfw"
)

// Screen is an on-screen opengl window that renders the channel.
type Screen struct {
	width           int
	height          int
	pixelsPerSample float64
	buffer          *types.Buffer
}

// NewScreen creates a new output screen of a given size and sample density.
func NewScreen(width int, height int, samplesPerPixel int) *Screen {
	s := Screen{
		width,
		height,
		1.0 / float64(samplesPerPixel),
		types.NewBuffer(width * samplesPerPixel),
	}
	return &s
}

// Render starts rendering a channel of waves samples to screen.
func (s *Screen) Render(values <-chan float64, sampleRate int) {
	runtime.LockOSThread()

	// NOTE: It appears that glfw 3.1 uses its own internal error callback.
	// glfw.SetErrorCallback(func(err glfw.ErrorCode, desc string) {
	// log.Fatalf("%v: %s\n", err, desc)
	// })
	if err := glfw.Init(); err != nil {
		log.Fatalf("Can't init glfw!")
	}
	defer glfw.Terminate()
	if err := gl.Init(); err != nil {
		log.Fatalf("Can't init gl!")
	}

	window, err := glfw.CreateWindow(s.width, s.height, "Muse", nil, nil)
	if err != nil {
		log.Fatalf("CreateWindow failed: %s", err)
	}
	if aw, ah := window.GetSize(); aw != s.width || ah != s.height {
		log.Fatalf("Window doesn't have the requested size: want %d,%d got %d,%d", s.width, s.height, aw, ah)
	}
	window.MakeContextCurrent()

	// Set window up to be [0, -1.0] -> [width, 1.0], black.
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()
	gl.Translated(-1, 0, 0)
	gl.Scaled(2/float64(s.width), 1.0, 1.0)
	gl.ClearColor(0.0, 0.0, 0.0, 0.0)

	// Actually start writing data to the buffer.
	s.buffer.GoPushChannel(values, sampleRate)

	// Keep drawing while we still can (and should).
	for !window.ShouldClose() && !s.buffer.IsFinished() {
		if window.GetKey(glfw.KeyEscape) == glfw.Press {
			break
		}
		gl.Clear(gl.COLOR_BUFFER_BIT)
		s.drawSignal()
		window.SwapBuffers()
		glfw.PollEvents()
	}

	// Keep window around, only close on esc.
	for !window.ShouldClose() && window.GetKey(glfw.KeyEscape) != glfw.Press {
		glfw.PollEvents()
	}
}

// drawSignal writes the input wave form(s) out to screen.
func (s *Screen) drawSignal() {
	gl.Begin(gl.LINE_STRIP)
	s.buffer.Each(func(index int, value float64) {
		gl.Vertex2d(float64(index)*s.pixelsPerSample, value)
	})
	gl.End()
}
