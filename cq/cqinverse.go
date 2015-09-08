package cq

import (
  "fmt"
  "math"
  "math/cmplx"

  "github.com/mjibson/go-dsp/fft"
)

const DEBUG_CQI = true

type CQInverse struct {
  kernel *CQKernel

  octaves int
  OutputLatency int
  buffers [][]float64
  olaBufs [][]float64

  latencies []int
  upsamplers []*Resampler
}

func NewCQInverse(params CQParams) *CQInverse {
  octaves := int(math.Ceil(math.Log(params.maxFrequency / params.minFrequency) / math.Log(2))); // TODO: math.Log2
  if octaves < 1 {
    panic("Need at least one octave")
  }

  kernel := NewCQKernel(params)
  p := kernel.Properties

  // Use exact powers of two for resampling rates. They don't have
  // to be related to our actual samplerate: the resampler only
  // cares about the ratio, but it only accepts integer source and
  // target rates, and if we start from the actual samplerate we
  // risk getting non-integer rates for lower octaves
  sourceRate := unsafeShift(octaves)
  latencies := make([]int, octaves, octaves)
  upsamplers := make([]*Resampler, octaves, octaves)

  // top octave, no resampling
  latencies[0] = 0
  upsamplers[0] = nil

  for i := 1; i < octaves; i++ {
    factor := unsafeShift(i)

    r := NewResampler(sourceRate / factor, sourceRate, 50, 0.05);

    if DEBUG_CQI {
      fmt.Printf("Inverse: octave %d: resample from %d to %d\n", i, sourceRate / factor, sourceRate)
    }

    // See ConstantQ.go for discussion on latency -- output
    // latency here is at target rate which, this way around, is
    // what we want
    latencies[i] = r.GetLatency()
    upsamplers[i] = r
  }

  // additionally we will have fftHop latency at individual octave
  // rate (before upsampling) for the overlap-add in each octave
  for i := 0; i < octaves; i++ {
    latencies[i] += p.fftHop * unsafeShift(i)
  }

  // Now reverse the drop adjustment made in ConstantQ to align the
  // atom centres across different octaves (but this time at output
  // sample rate)
  emptyHops := p.firstCentre / p.atomSpacing

  pushes := make([]int, octaves, octaves)
  for i := 0; i < octaves; i++ {
    factor := unsafeShift(i)
    pushHops := emptyHops * unsafeShift(octaves - i - 1) - emptyHops
    push := ((pushHops * p.fftHop) * factor) / p.atomsPerFrame
    pushes[i] = push
  }

  maxLatLessPush := 0
  for i := 0; i < octaves; i++ {
    latLessPush := latencies[i] - pushes[i]
    if latLessPush > maxLatLessPush {
      maxLatLessPush = latLessPush
    }
  }
  totalLatency := maxLatLessPush + 10
  if totalLatency < 0 {
    totalLatency = 0
  }

  outputLatency := totalLatency + p.firstCentre * unsafeShift(octaves - 1)
  if DEBUG_CQI {
    fmt.Printf("totalLatency = %d, outputLatency = %d\n", totalLatency, outputLatency)
  }

  buffers := make([][]float64, octaves)
  for i := 0; i < octaves; i++ {
    // Calculate the difference between the total latency applied
    // across all octaves, and the existing latency due to the
    // upsampler for this octave.
    latencyPadding := totalLatency - latencies[i] + pushes[i]
    if DEBUG_CQI {
      fmt.Printf("octave %d: push %d, resampler latency inc overlap space %d, latencyPadding %d (/factor = %d)\n",
        i, pushes[i], latencies[i], latencyPadding, latencyPadding / unsafeShift(i));
    }

    buffers[i] = make([]float64, latencyPadding, latencyPadding)
  }

  olaBufs := make([][]float64, octaves, octaves)
  for i := 0; i < octaves; i++ {
    // Fixed-size buffer for IFFT overlap-add
    olaBufs[i] = make([]float64, p.fftSize, p.fftSize)
  }

  return &CQInverse {
    kernel, 
    octaves,

    outputLatency,
    buffers,
    olaBufs,

    latencies,
    upsamplers,
  }
}


func (cqi *CQInverse) Process(block [][]complex128, print bool) []float64 {
  octaves := cqi.octaves

  // The input data is of the form produced by ConstantQ::process --
  // an unknown number N of columns of varying height. We assert
  // that N is a multiple of atomsPerFrame * 2^(octaves-1), as must
  // be the case for data that came directly from our ConstantQ
  // implementation.
  widthProvided := len(block)
  if widthProvided == 0 {
    return cqi.drawFromBuffers()
  }

  blockWidth := cqi.kernel.Properties.atomsPerFrame * unsafeShift(cqi.octaves - 1)
  if widthProvided % blockWidth != 0 {
    fmt.Printf("ERROR: inverse process block size (%d) must be a multiple of atoms * 2^(octaves - 1) = %d * 2^(%d - 1) = %d\n",
      widthProvided, cqi.kernel.Properties.atomsPerFrame, cqi.octaves, blockWidth)
    panic("CQI process block size incorrect")
  }

  // Procedure:
  //
  // 1. Slice the list of columns into a set of lists of columns,
  // one per octave, each of width N / (2^octave-1) and height
  // binsPerOctave, containing the values present in that octave
  //
  // 2. Group each octave list by atomsPerFrame columns at a time,
  // and stack these so as to achieve a list, for each octave, of
  // taller columns of height binsPerOctave * atomsPerFrame
  //
  // 3. For each taller column, take the product with the inverse CQ
  // kernel (which is the conjugate of the forward kernel) and
  // perform an inverse FFT
  //
  // 4. Overlap-add each octave's resynthesised blocks (unwindowed)
  //
  // 5. Resample each octave's overlap-add stream to the original
  // rate
  //
  // 6. Sum the resampled streams and return
  bpo := cqi.kernel.Properties.binsPerOctave
  for i := 0; i < octaves; i++ {
    // Step 1
    oct := make([][]complex128, 0)

    for j := 0; j < widthProvided; j++ {
      h := len(block[j])
      if h < bpo * (i + 1) {
        continue
      }

      col := make([]complex128, bpo, bpo)
      copy(col, block[j][bpo*i:bpo*(i+1)])
      oct = append(oct, col)
    }

    // Steps 2, 3, 4, 5
    cqi.processOctave(i, oct, print)
  }

  // Step 6
  return cqi.drawFromBuffers()
}


func (cqi *CQInverse) drawFromBuffers() []float64 {
  octaves := cqi.octaves

  // 6. Sum the resampled streams and return
  available := 0
  for i := 0; i < octaves; i++ {
    if i == 0 || len(cqi.buffers[i]) < available {
      available = len(cqi.buffers[i])
    }
  }

  result := make([]float64, available, available)
  if available == 0 {
    return result
  }

  for i := 0; i < octaves; i++ {
    for j := 0; j < available; j++ {
      result[j] += cqi.buffers[i][j];
    }
    cqi.buffers[i] = cqi.buffers[i][available:]
  }
  return result
}

func (cqi *CQInverse) GetRemainingOutput() []float64 {
  octaves := cqi.octaves

  for j := 0; j < octaves; j++ {
    factor := unsafeShift(j)
    latency := 0
    if j > 0 {
      // TODO - read from cqi.latencies instead
      latency = cqi.upsamplers[j].GetLatency() / factor
    }

    for i := 0; i < (latency + cqi.kernel.Properties.binsPerOctave) / cqi.kernel.Properties.fftHop; i++ {
      cqi.overlapAddAndResample(j, make([]float64, len(cqi.olaBufs[j]), len(cqi.olaBufs[j])), false)
    }
  }
  return cqi.drawFromBuffers()
}


func (cqi *CQInverse) processOctave(octave int, columns [][]complex128, print bool) {
  // 2. Group each octave list by atomsPerFrame columns at a time,
  // and stack these so as to achieve a list, for each octave, of
  // taller columns of height binsPerOctave * atomsPerFrame

  bpo := cqi.kernel.Properties.binsPerOctave
  apf := cqi.kernel.Properties.atomsPerFrame

  ncols := len(columns)
  if ncols % apf != 0 {
    fmt.Printf("Error: inverse process octave %d, # columns (%d) must be a multiple of atoms/frame (%d)\n",
      octave, ncols, apf)
    panic("Invalid argument to inverse processOctave")
  }

  for i := 0; i < ncols; i+= apf {
    tallCol := make([]complex128, bpo * apf, bpo * apf)

    for b := 0; b < bpo; b++ {
      for a := 0; a < apf; a++ {
        tallCol[b * apf + a] = columns[i + a][bpo - b- 1]
      }
    }

    cqi.processOctaveColumn(octave, tallCol, print && (i == 0))
  }
}


func (cqi *CQInverse) processOctaveColumn(octave int, column []complex128, print bool) {
  // 3. For each taller column, take the product with the inverse CQ
  // kernel (which is the conjugate of the forward kernel) and
  // perform an inverse FFT
  bpo := cqi.kernel.Properties.binsPerOctave
  apf := cqi.kernel.Properties.atomsPerFrame
  fftSize := cqi.kernel.Properties.fftSize

  if len(column) != bpo * apf {
    fmt.Printf("Error: column in octave %d has size %d, required = %d * %d = %d\n",
      octave, len(column), bpo, apf, bpo * apf)
    panic("Invalid argument to inverse processOctaveColumn")
  }

  if (print) {
    fmt.Printf("Input to POC has size %d\n", len(column));
    c := complex(0, 0);
    for _, v := range column {
      c += v
    }
    fmt.Printf("Input to POC has sum: %.4f, %.4f\n", real(c), imag(c));
  }

  transformed := cqi.kernel.ProcessInverse(column)
  // halfLen := fftSize / 2 + 1

  if (print) {
    c := complex(0, 0)
    for _, v := range transformed {
      c += v
    }
    fmt.Printf("kernel-transformed POC has sum %.4f, %.4f\n", real(c), imag(c))
  }
  if (print) {
    maxr := 0
    for i, v := range transformed {
      if real(v) > real(transformed[maxr]) {
        maxr = i
      }
    }
    for i := 0; i < 10; i++ {
      fmt.Printf("transformed[%d] = %.6f, %.6f\n", i + maxr, real(transformed[i + maxr]), imag(transformed[i + maxr]))
    }
  }

  halfLen := fftSize / 2 + 1
  if (print) {
    fmt.Printf("result[0] / result[%d] = %v / %v\n", halfLen - 1, transformed[0], transformed[halfLen - 1])
  }

  bigDiff := 0.0
  for i := halfLen; i < fftSize; i++ {
    next := complex(real(transformed[fftSize - i]), -imag(transformed[fftSize - i]))
    nextDiff := cmplx.Abs(next - transformed[i])
    if (nextDiff > bigDiff) {
      bigDiff = nextDiff
    }
    transformed[i] = next
  }
  if (print) {
    fmt.Printf("Biggest diff was %.4f\n", bigDiff);
  }


  // fmt.Printf("Transformed length = %d\n", len(transformed))
  result := fft.IFFT(transformed)
  asFloat := make([]float64, fftSize, fftSize)

  if (print) {
    for i := 0; i < 10; i++ {
      fmt.Printf("timeDomain[%d] = %.9f, %.9f\n", i, real(result[i]), imag(result[i]))
    }
  }

  for i := 0; i < fftSize; i++ {
    asFloat[i] = real(result[i])
  }
  cqi.overlapAddAndResample(octave, asFloat, print)

}

func (cqi *CQInverse) overlapAddAndResample(octave int, seq []float64, print bool) {
  // 4. Overlap-add each octave's resynthesised blocks (unwindowed)
  //
  // and
  //
  // 5. Resample each octave's overlap-add stream to the original
  // rate

  fftHop := cqi.kernel.Properties.fftHop

  if len(seq) != len(cqi.olaBufs[octave]) {
    fmt.Printf("Error: inverse overlap add sequence size %d, expected to be OLA buffer size %d\n",
      len(seq), len(cqi.olaBufs[octave]))
    panic ("Illegal argument to Inverse overlapAdd")
  }

  toResample := cqi.olaBufs[octave][:fftHop]
  resampled := toResample
  if octave > 0 {
    resampled = cqi.upsamplers[octave].Process(toResample)
  }

  cqi.buffers[octave] = append(cqi.buffers[octave], resampled...)


  cqi.olaBufs[octave] = cqi.olaBufs[octave][fftHop:]
  pad := make([]float64, fftHop, fftHop)
  cqi.olaBufs[octave] = append(cqi.olaBufs[octave], pad...)

  for i := 0; i < cqi.kernel.Properties.fftSize; i++ {
    cqi.olaBufs[octave][i] += seq[i]
  }
}

