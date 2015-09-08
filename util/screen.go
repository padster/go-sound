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

// Event is something that occurred at a single sample of the values.
type Event struct {
	// (r, g, b) colours for the event.
	R float32
	G float32
	B float32
}

// Line is the samples channel for the line, plus their color.
type Line struct {
	Values      <-chan float64
	R           float32
	G           float32
	B           float32
	valueBuffer *types.Buffer
}

// NewLine creates a line from the exported fields.
func NewLine(v <-chan float64, r float32, g float32, b float32) Line {
	return Line{v, r, g, b, nil}
}

// Screen is an on-screen opengl window that renders the channel.
type Screen struct {
	width           int
	height          int
	pixelsPerSample float64
	eventBuffer     *types.TypedBuffer

	// TODO(padster): Move this into constructor, and call much later.
	lines []Line
}

// NewScreen creates a new output screen of a given size and sample density.
func NewScreen(width int, height int, samplesPerPixel int) *Screen {
	s := Screen{
		width,
		height,
		1.0 / float64(samplesPerPixel),
		types.NewTypedBuffer(width * samplesPerPixel),
		nil, // lines
	}
	return &s
}

// Render starts rendering a channel of waves samples to screen.
func (s *Screen) Render(values <-chan float64, sampleRate int) {
	s.RenderLinesWithEvents([]Line{
		Line{values, 1.0, 1.0, 1.0, nil}, // Default to white
	}, nil, sampleRate)
}

// RenderLinesWithEvents renders multiple channels of samples to screen, and draws events.
func (s *Screen) RenderLinesWithEvents(lines []Line, events <-chan interface{}, sampleRate int) {
	s.lines = lines

	runtime.LockOSThread()

	// NOTE: It appears that glfw 3.1 uses its own internal error callback.
	// glfw.SetErrorCallback(func(err glfw.ErrorCode, desc string) {
	// log.Fatalf("%v: %s\n", err, desc)
	// })
	if err := glfw.Init(); err != nil {
		log.Fatalf("Can't init glfw: %v!", err)
	}
	defer glfw.Terminate()

	window, err := glfw.CreateWindow(s.width, s.height, "Muse", nil, nil)
	if err != nil {
		log.Fatalf("CreateWindow failed: %s", err)
	}
	if aw, ah := window.GetSize(); aw != s.width || ah != s.height {
		log.Fatalf("Window doesn't have the requested size: want %d,%d got %d,%d", s.width, s.height, aw, ah)
	}
	window.MakeContextCurrent()

	// Must gl.Init() *after* MakeContextCurrent
	if err := gl.Init(); err != nil {
		log.Fatalf("Can't init gl: %v!", err)
	}

	// Set window up to be [0, -1.0] -> [width, 1.0], black.
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()
	gl.Translated(-1, 0, 0)
	gl.Scaled(2/float64(s.width), 1.0, 1.0)
	gl.ClearColor(0.0, 0.0, 0.0, 0.0)

	// Actually start writing data to the buffer
	for i, _ := range s.lines {
		s.lines[i].valueBuffer = types.NewBuffer(int(float64(s.width) / s.pixelsPerSample))
		s.lines[i].valueBuffer.GoPushChannel(s.lines[i].Values, sampleRate)
	}
	if events != nil {
		s.eventBuffer.GoPushChannel(events, sampleRate)
	}

	// Keep drawing while we still can (and should).
	for !window.ShouldClose() && !s.bufferFinished() {
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

// bufferFinished returns whether any of the input channels have closed.
func (s *Screen) bufferFinished() bool {
	if s.eventBuffer.IsFinished() {
		return true
	}
	for _, l := range s.lines {
		if l.valueBuffer.IsFinished() {
			return true
		}
	}
	return false
}

// drawSignal writes the input wave form(s) out to screen.
func (s *Screen) drawSignal() {
	s.eventBuffer.Each(func(index int, value interface{}) {
		if value != nil {
			e := value.(Event)
			gl.Color3f(e.R, e.G, e.B)
			x := float64(index) * s.pixelsPerSample
			gl.Begin(gl.LINE_STRIP)
			gl.Vertex2d(x, -1.0)
			gl.Vertex2d(x, 1.0)
			gl.End()
		}
	})

	for _, l := range s.lines {
		gl.Color3f(l.R, l.G, l.B)
		gl.Begin(gl.LINE_STRIP)
		l.valueBuffer.Each(func(index int, value float64) {
			gl.Vertex2d(float64(index)*s.pixelsPerSample, value)
		})
		gl.End()
	}
}
