package cq

import (
	"fmt"
	"math"
	"math/cmplx"
)

const (
	TAU = 2.0 * math.Pi
)

type Spectrogram struct {
	cq *ConstantQ

	buffer     [][]complex128
	prevColumn []complex128
}

func NewSpectrogram(params CQParams) *Spectrogram {
	return &Spectrogram{
		NewConstantQ(params),
		make([][]complex128, 0, 128),
		make([]complex128, 0, 0),
	}
}

func (spec *Spectrogram) ProcessChannel(samples <-chan float64) <-chan []complex128 {
	result := make(chan []complex128)

	go func() {
		partial := spec.cq.ProcessChannel(samples)

		height := spec.cq.BinCount()
		buffer := make([][]complex128, 0, 128) // HACK - get the correct size

		first := false
		for column := range partial {
			buffer = append(buffer, column)
			if first {
				if len(column) != height {
					panic("First partial info must be for all values.")
				}
				first = false
			} else {
				if len(column) == height {
					full := spec.fullInterpolate(buffer)
					for _, ic := range full {
						result <- ic
					}
					buffer = buffer[len(buffer)-1:]
				}
			}
		}

		if len(buffer[0]) != height {
			panic("Oops - can't interpolate the ending part, wrong height :/")
		}
		for _, ic := range holdInterpolate(buffer) {
			result <- ic
		}
		close(result)
	}()

	return result
}

func (spec *Spectrogram) Process(values []float64) [][]complex128 {
	return spec.interpolate(spec.cq.Process(values), false)
}

func (spec *Spectrogram) GetRemainingOutput() [][]complex128 {
	return spec.interpolate(spec.cq.GetRemainingOutput(), true)
}

// Post process by writing to linear interpolator
func (spec *Spectrogram) interpolate(cq [][]complex128, insist bool) [][]complex128 {
	// TODO: make copy here? currently we copy elsewhere.
	spec.buffer = append(spec.buffer, cq...)
	return spec.fetchInterpolated(insist)
}

// Interpolate by copying from the previous column
func (spec *Spectrogram) fetchHold() [][]complex128 {
	width := len(spec.buffer)
	height := spec.cq.BinCount()

	out := make([][]complex128, width, width)

	for i := 0; i < width; i++ {
		col := spec.buffer[i]

		thisHeight, prevHeight := len(col), len(spec.prevColumn)
		for j := thisHeight; j < height; j++ { // TODO - collapse into two copies, not a for loop.
			if j < prevHeight {
				col = append(col, spec.prevColumn[j])
			} else {
				col = append(col, complex(0, 0))
			}
		}

		spec.prevColumn = col
		out[i] = col
	}

	spec.buffer = make([][]complex128, 0, 256)
	return out
}

func (spec *Spectrogram) fetchInterpolated(insist bool) [][]complex128 {
	width := len(spec.buffer)
	height := spec.cq.BinCount()

	if width == 0 {
		return make([][]complex128, 0, 0)
	}

	firstFullHeight, secondFullHeight := -1, -1

	for i := 0; i < width; i++ {
		if len(spec.buffer[i]) == height {
			if firstFullHeight == -1 {
				firstFullHeight = i
			} else if secondFullHeight == -1 {
				secondFullHeight = i
				break
			}
		}
	}

	if firstFullHeight > 0 {
		// Stuff at the start we can't interpolate. Copy that verbatim, and recurse
		out := spec.buffer[:firstFullHeight]
		spec.buffer = spec.buffer[firstFullHeight:]
		return append(out, spec.fetchInterpolated(insist)...)
	} else if firstFullHeight < 0 || secondFullHeight < 0 {
		// Wait until we have somethinng we can interpolate...
		if insist {
			return spec.fetchHold()
		} else {
			return make([][]complex128, 0, 0)
		}
	} else {
		// firstFullHeight == 0 and secondFullHeight also valid. Can interpolate
		out := spec.fullInterpolate(spec.buffer[:secondFullHeight+1])
		spec.buffer = spec.buffer[secondFullHeight:]
		return append(out, spec.fetchInterpolated(insist)...)
	}
}

func (spec *Spectrogram) fullInterpolate(values [][]complex128) [][]complex128 {
	// Last entry is the interpolation end boundary, hence the -1
	width, height := len(values)-1, len(values[0])
	bpo := spec.cq.bpo()

	if height != len(values[width]) {
		fmt.Printf("interpolateInPlace requires start and end arrays to be the same (full) size, %d != %d\n",
			len(values[0]), len(values[width]))
		panic("IAE to interpolateInPlace")
	}

	result := make([][]complex128, width, width)
	for i := 0; i < width; i++ {
		result[i] = make([]complex128, height, height)
		copy(result[i], values[i])
	}

	// For each height...
	for y := 0; y < height; y++ {
		// spacing = index offset to next column bigger than that height (y)
		spacing := width
		for i := 1; i < width; i++ {
			thisHeight := len(values[i])
			if thisHeight > height {
				panic("interpolateInPlace requires the first column to be the highest")
			}
			if thisHeight > y {
				spacing = i
				break // or: remove, and convert to i < spacing in for loop?
			}
		}

		if spacing < 2 {
			continue
		}
		thisOctave, lastOctave := y, y-bpo
		if lastOctave < 0 {
			panic("Oops, can't interpolate in the first octave?")
		}

		for i := 0; i+spacing <= width; i += spacing {
			// NOTE: can't use result[] instead of values[] here, as result[] doesn't include the full right column yet.
			thisStart, thisEnd := values[i][thisOctave], values[i+spacing][thisOctave]

			lastPhaseStart := cmplx.Phase(values[i][lastOctave])
			lastPhaseAt := lastPhaseStart
			for j := 1; j < spacing; j++ {
				lastPhaseAt = makeCloser(lastPhaseAt, cmplx.Phase(result[i+j][lastOctave]))
			}
			totalLastPhaseDiff := lastPhaseAt - lastPhaseStart

			// Tweak this to allow the slope of the lower octave phase change to differ from the higher octave's
			upperScale := 0.5
			targetLastPhase := makeCloser(upperScale*totalLastPhaseDiff, cmplx.Phase(thisEnd)-cmplx.Phase(thisStart))
			diffScale := 0.0
			if math.Abs(totalLastPhaseDiff) > 1e-5 {
				diffScale = targetLastPhase / totalLastPhaseDiff
			}

			lastPhaseAt = lastPhaseStart
			for j := 1; j < spacing; j++ {
				lastPhaseAt := makeCloser(lastPhaseAt, cmplx.Phase(result[i+j][lastOctave]))
				proportion := float64(j) / float64(spacing)
				interpolated := logInterpolate(thisStart, thisEnd, proportion, lastPhaseStart+(lastPhaseAt-lastPhaseStart)*diffScale)
				result[i+j][y] = interpolated
			}
		}
	}

	return result
}

func logInterpolate(this1, thisN complex128, proportion float64, interpolatedPhase float64) complex128 {
	if cmplx.Abs(this1) < cmplx.Abs(thisN) {
		return logInterpolate(thisN, this1, 1-proportion, interpolatedPhase)
	}

	// Simple linear interpolation for DB power.
	z := thisN / this1
	zLogAbs := math.Log(cmplx.Abs(z))
	cLogAbs := zLogAbs * proportion
	cAbs := math.Exp(cLogAbs)

	return cmplx.Rect(cAbs*cmplx.Abs(this1), interpolatedPhase)
}

// Return the closest number X to toShift, such that X mod Tau == modTwoPi
func makeCloser(toShift, modTau float64) float64 {
	if math.IsNaN(modTau) {
		modTau = 0.0
	}
	// Minimize |toShift - (modTau + tau * cyclesToAdd)|
	// toShift - modTau - tau * CTA = 0
	cyclesToAdd := (toShift - modTau) / TAU
	return modTau + float64(Round(cyclesToAdd))*TAU
}

func holdInterpolate(values [][]complex128) [][]complex128 {
	// Hmm...maybe instead interpolate to zeroes?
	width, height := len(values), len(values[0])
	for i := 1; i < width; i++ {
		from := len(values[i])
		if from >= height {
			panic("hold interpolate input has wrong structure :(")
		}
		values[i] = append(values[i], values[i-1][from:height]...)
	}
	return values
}

// HACK
func Round(value float64) int64 {
	if value < 0.0 {
		value -= 0.5
	} else {
		value += 0.5
	}
	return int64(value)
}
