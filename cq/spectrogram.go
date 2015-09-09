package cq

import (
	"fmt"
	"math"
	"math/cmplx"
)

type Spectrogram struct {
	cq *ConstantQ

	buffer [][]complex128
	prevColumn []complex128
}

func NewSpectrogram(params CQParams) *Spectrogram {
	return &Spectrogram{
		NewConstantQ(params),
		make([][]complex128, 0, 128),
		make([]complex128, 0, 0),
	}
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
	height := spec.cq.binCount()

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
	return out;
}

func (spec *Spectrogram) fetchInterpolated(insist bool) [][]complex128 {
	width := len(spec.buffer)
	height := spec.cq.binCount()

	if width == 0 {
		return make([][]complex128, 0, 0)
	}

	//!!! This is surprisingly messy. I must be missing something.

	// We can only return any data when we have at least one column
	// that has the full height in the buffer, that is not the first
	// column.
	//
	// If the first col has full height, and there is another one
	// later that also does, then we can interpolate between those, up
	// to but not including the second full height column.  Then we
	// drop and return the columns we interpolated, leaving the second
	// full-height col as the first col in the buffer. And repeat as
	// long as enough columns are available.
	//
	// If the first col does not have full height, then (so long as
	// we're following the logic above) we must simply have not yet
	// reached the first full-height column in the CQ output, and we
	// can interpolate nothing.

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
		out := spec.fullInterpolate(spec.buffer[:secondFullHeight + 1])
		spec.buffer = spec.buffer[secondFullHeight:]
		return append(out, spec.fetchInterpolated(insist)...)
	}
}

func (spec *Spectrogram) fullInterpolate(values [][]complex128) [][]complex128 {
	// Last entry is the interpolation end boundary, hence the -1
	width, height := len(values) - 1, len(values[0]) 

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

		for i := 0; i + spacing <= width; i += spacing {
			for j := 1; j < spacing; j++ {
				proportion := float64(j) / float64(spacing)
				interpolated := logInterpolate(values[i][y], values[i + spacing][y], proportion)
				result[i + j][y] = interpolated
			}
		}
	}

	return result;
}

func logInterpolate(a complex128, b complex128, proportion float64) complex128 {
	// TODO - precalc arg/norm outside the loop.
	if cmplx.Abs(a) < cmplx.Abs(b) {
		return logInterpolate(b, a, 1-proportion)
	}

	z := b / a
	zArg := cmplx.Phase(z)
	if zArg > 0 {
		// aArg -> bArg, or aArg -> bArg + 2PI, whichever is closer
		if zArg > math.Pi {
			zArg -= 2 * math.Pi
		}
	} else {
		// aArg -> bArg, or aArg -> bArg - 2PI, whichever is closer
		if zArg < -math.Pi {
			zArg += 2 * math.Pi
		}
	}

	zLogAbs := math.Log(cmplx.Abs(z))
	cArg, cLogAbs := zArg * proportion, zLogAbs * proportion
	cAbs := math.Exp(cLogAbs)
	return a * cmplx.Rect(cAbs, cArg)
}
