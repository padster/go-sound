// Renders various data from a channel of [-1, 1] onto screen.
package util

import (
  "log"
  // "math"
  "math/cmplx"
  "runtime"

  "github.com/padster/go-sound/types"

  // TODO(padster) - migrate to core, not compat.
  gl "github.com/go-gl/gl/v3.3-compatibility/gl"
  glfw "github.com/go-gl/glfw/v3.1/glfw"
)

// Line is the samples channel for the line, plus their color.
type ComplexLine struct {
  Values      <-chan []complex128
  R           float32
  G           float32
  B           float32
  valueBuffer *types.TypedBuffer
}

// Screen is an on-screen opengl window that renders the channel.
type SpectrogramScreen struct {
  width           int
  height          int
  bpo             int
  pixelsPerSample float64
  eventBuffer     *types.TypedBuffer

  // TODO(padster): Move this into constructor, and call much later.
  line *ComplexLine
}

// NewScreen creates a new output screen of a given size and sample density.
func NewSpectrogramScreen(width int, height int, bpo int) *SpectrogramScreen {
  samplesPerPixel := 1 // HACK - parameterize?
  s := SpectrogramScreen{
    width,
    height,
    bpo,
    1.0 / float64(samplesPerPixel),
    types.NewTypedBuffer(width * samplesPerPixel),
    nil, // lines
  }
  return &s
}

// Render starts rendering a channel of waves samples to screen.
func (s *SpectrogramScreen) Render(values <-chan []complex128, sampleRate int) {
  s.RenderLineWithEvents(
    &ComplexLine{values, 1.0, 1.0, 1.0, nil}, // Default to white
  nil, sampleRate)
}

// RenderLinesWithEvents renders multiple channels of samples to screen, and draws events.
func (s *SpectrogramScreen) RenderLineWithEvents(line *ComplexLine, events <-chan interface{}, sampleRate int) {
  s.line = line

  runtime.LockOSThread()

  // NOTE: It appears that glfw 3.1 uses its own internal error callback.
  // glfw.SetErrorCallback(func(err glfw.ErrorCode, desc string) {
  // log.Fatalf("%v: %s\n", err, desc)
  // })
  if err := glfw.Init(); err != nil {
    log.Fatalf("Can't init glfw: %v!", err)
  }
  defer glfw.Terminate()

  window, err := glfw.CreateWindow(s.width, s.height, "Spectrogram", nil, nil)
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

  // Set window up to be [0, 0] -> [width, height], black.
  gl.MatrixMode(gl.MODELVIEW)
  gl.LoadIdentity()
  gl.Translated(-1, -1, 0)
  gl.Scaled(2/float64(s.width), 2/float64(s.height), 1.0)
  gl.ClearColor(0.0, 0.0, 0.0, 0.0)

  gl.ShadeModel(gl.FLAT)

  // Actually start writing data to the buffer
  s.line.valueBuffer = types.NewTypedBuffer(int(float64(s.width) / s.pixelsPerSample))
  s.line.valueBuffer.GoPushChannel(hackWrapChannel(s.line.Values), sampleRate)
  if events != nil {
    s.eventBuffer.GoPushChannel(events, sampleRate)
  }

  gl.Hint(gl.POINT_SMOOTH_HINT, gl.FASTEST)

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
func (s *SpectrogramScreen) bufferFinished() bool {
  if s.eventBuffer.IsFinished() {
    return true
  }
  return s.line.valueBuffer.IsFinished()
}

// drawSignal writes the input wave form(s) out to screen.
func (s *SpectrogramScreen) drawSignal() {
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

  spacing := 1
  
  gl.PointSize(1.0)
  gl.Begin(gl.POINTS)
  s.line.valueBuffer.Each(func(index int, value interface{}) {
    col := value.([]complex128)
    for i, v := range col { 
      // p = log(|v|) = [-20, 5], so map to [0, 1]
      // p := math.Log(cmplx.Abs(v) + 1.e-8)
      // grey := (p + 20.0) / 25.0
      p := cmplx.Abs(v)
      grey := p / 15.0
      if grey > 1.0 {
        grey = 1.0
      } else if grey < 0.0 {
        grey = 0.0
      }

      // HACK: Stretch to make the darks darker and the whites whiter.
      // grey = grey * grey * grey * grey // more space at the top, [0, 1]
      // grey = 2.0 * grey - 1.0 // [-1, 1]
      // grey = math.Tanh(2.0 * grey) // streched, still [-1, 1] 
      // grey = (grey + 1.0) / 2.0
      

      gl.Color3d(grey, grey, grey)
      // gl.Vertex2d(float64(index)*s.pixelsPerSample, float64(s.height - 1 - (3 * i + 0)))
      gl.Vertex2d(float64(index)*s.pixelsPerSample, float64(s.height - 1 - (spacing * i)))
      // gl.Vertex2d(float64(index)*s.pixelsPerSample, float64(s.height - 1 - (3 * i + 2))) 
    }

    gl.Color3ub(255, 0, 0)
    for i := 1; i < 7; i++ {
      gl.Vertex2d(float64(index)*s.pixelsPerSample, float64(i * spacing * s.bpo))
    }
  })
  gl.End()
}

// Grr...no generics, but also chan T can't be a chan interface{} ???
func hackWrapChannel(in <-chan []complex128) (<-chan interface{}) {
  result := make(chan interface{})
  go func(out chan interface{}) {
    for i := range in {
      out <- interface{}(i)
    }
    close(out)
  }(result)
  return result
}
