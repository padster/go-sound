package cq

import (
	"fmt"
	"math"
)

const DEBUG_R = false

type KaiserWindow struct {
	length int
	beta   float64
}

type Phase struct {
	nextPhase int
	filter    []float64
	drop      int
}

type Resampler struct {
	sourceRate int
	targetRate int
	gcd        int
	peakToPole float64

	filterLength int
	phase        int
	latency      int
	phaseData    []Phase

	buffer       []float64
	bufferOrigin int
}

func NewResampler(sourceRate int, targetRate int, snr float64, bandwidth float64) *Resampler {
	higher, lower := sourceRate, targetRate
	if higher < lower {
		higher, lower = lower, higher
	}

	gcd := calcGcd(lower, higher)
	peakToPole := float64(higher / gcd)
	if targetRate < sourceRate {
		// antialiasing filter, should be slightly below nyquist
		peakToPole = peakToPole / (1.0 - bandwidth/2.0)
	}

	params := kaiserForBandwidth(snr, bandwidth, float64(higher/gcd))
	if params.length%2 == 0 {
		params.length++
	}
	params.length = minInt(params.length, 200001)
	filterLength := params.length

	filter := make([]float64, filterLength, filterLength)
	sincWindow := buildSincWindow(filterLength, peakToPole*2)
	kWindow := buildKaiserWindow(params)
	for i := 0; i < filterLength; i++ {
		filter[i] = sincWindow[i] * kWindow[i]
	}

	inputSpacing, outputSpacing := targetRate/gcd, sourceRate/gcd

	if DEBUG_R {
		fmt.Printf("resample %v -> %v: inputSpacing %v, outputSpacing %v: filter length %v\n",
			sourceRate, targetRate, inputSpacing, outputSpacing, filterLength)
	}

	// Now we have a filter of (odd) length flen in which the lower
	// sample rate corresponds to every n'th point and the higher rate
	// to every m'th where n and m are higher and lower rates divided
	// by their gcd respectively. So if x coordinates are on the same
	// scale as our filter resolution, then source sample i is at i *
	// (targetRate / gcd) and target sample j is at j * (sourceRate /
	// gcd).

	// To reconstruct a single target sample, we want a buffer (real
	// or virtual) of flen values formed of source samples spaced at
	// intervals of (targetRate / gcd), in our example case 3.  This
	// is initially formed with the first sample at the filter peak.
	//
	// 0  0  0  0  a  0  0  b  0
	//
	// and of course we have our filter
	//
	// f1 f2 f3 f4 f5 f6 f7 f8 f9
	//
	// We take the sum of products of non-zero values from this buffer
	// with corresponding values in the filter
	//
	// a * f5 + b * f8
	//
	// Then we drop (sourceRate / gcd) values, in our example case 4,
	// from the start of the buffer and fill until it has flen values
	// again
	//
	// a  0  0  b  0  0  c  0  0
	//
	// repeat to reconstruct the next target sample
	//
	// a * f1 + b * f4 + c * f7
	//
	// and so on.
	//
	// Above I said the buffer could be "real or virtual" -- ours is
	// virtual. We don't actually store all the zero spacing values,
	// except for padding at the start; normally we store only the
	// values that actually came from the source stream, along with a
	// phase value that tells us how many virtual zeroes there are at
	// the start of the virtual buffer.  So the two examples above are
	//
	// 0 a b  [ with phase 1 ]
	// a b c  [ with phase 0 ]
	//
	// Having thus broken down the buffer so that only the elements we
	// need to multiply are present, we can also unzip the filter into
	// every-nth-element subsets at each phase, allowing us to do the
	// filter multiplication as a simply vector multiply. That is, rather
	// than store
	//
	// f1 f2 f3 f4 f5 f6 f7 f8 f9
	//
	// we store separately
	//
	// f1 f4 f7
	// f2 f5 f8
	// f3 f6 f9
	//
	// Each time we complete a multiply-and-sum, we need to work out
	// how many (real) samples to drop from the start of our buffer,
	// and how many to add at the end of it for the next multiply.  We
	// know we want to drop enough real samples to move along by one
	// computed output sample, which is our outputSpacing number of
	// virtual buffer samples. Depending on the relationship between
	// input and output spacings, this may mean dropping several real
	// samples, one real sample, or none at all (and simply moving to
	// a different "phase").

	phaseData := make([]Phase, inputSpacing, inputSpacing)

	for phase := 0; phase < inputSpacing; phase++ {
		nextPhase := phase - outputSpacing
		for nextPhase < 0 {
			nextPhase += inputSpacing
		}
		nextPhase = nextPhase % inputSpacing
		filtZipLength := roundUp(float64(filterLength-phase) / float64(inputSpacing))
		drop := roundUp(math.Max(0.0, float64(outputSpacing-phase)/float64(inputSpacing)))

		phaseData[phase] = Phase{
			nextPhase,
			make([]float64, filtZipLength, filtZipLength),
			drop,
		}
		for i := 0; i < filtZipLength; i++ {
			phaseData[phase].filter[i] = filter[i*inputSpacing+phase]
		}

	}

	if DEBUG_R {
		cp, totDrop := 0, 0
		for i := 0; i < inputSpacing; i++ {
			fmt.Printf("Phase = %v, drop = %v, filter length = %v, next phase = %v\n",
				cp, phaseData[cp].drop, len(phaseData[cp].filter), phaseData[cp].nextPhase)
			totDrop += phaseData[cp].drop
			cp = phaseData[cp].nextPhase
		}
		fmt.Printf("total drop = %v\n", totDrop)
	}

	// The May implementation of this uses a pull model -- we ask the
	// resampler for a certain number of output samples, and it asks
	// its source stream for as many as it needs to calculate
	// those. This means (among other things) that the source stream
	// can be asked for enough samples up-front to fill the buffer
	// before the first output sample is generated.
	//
	// In this implementation we're using a push model in which a
	// certain number of source samples is provided and we're asked
	// for as many output samples as that makes available. But we
	// can't return any samples from the beginning until half the
	// filter length has been provided as input. This means we must
	// either return a very variable number of samples (none at all
	// until the filter fills, then half the filter length at once) or
	// else have a lengthy declared latency on the output. We do the
	// latter. (What do other implementations do?)
	//
	// We want to make sure the first "real" sample will eventually be
	// aligned with the centre sample in the filter (it's tidier, and
	// easier to do diagnostic calculations that way). So we need to
	// pick the initial phase and buffer fill accordingly.
	//
	// Example: if the inputSpacing is 2, outputSpacing is 3, and
	// filter length is 7,
	//
	//    x     x     x     x     a     b     c ... input samples
	// 0  1  2  3  4  5  6  7  8  9 10 11 12 13 ...
	//          i        j        k        l    ... output samples
	// [--------|--------] <- filter with centre mark
	//
	// Let h be the index of the centre mark, here 3 (generally
	// int(filterLength/2) for odd-length filters).
	//
	// The smallest n such that h + n * outputSpacing > filterLength
	// is 2 (that is, ceil((filterLength - h) / outputSpacing)), and
	// (h + 2 * outputSpacing) % inputSpacing == 1, so the initial
	// phase is 1.
	//
	// To achieve our n, we need to pre-fill the "virtual" buffer with
	// 4 zero samples: the x's above. This is int((h + n *
	// outputSpacing) / inputSpacing). It's the phase that makes this
	// buffer get dealt with in such a way as to give us an effective
	// index for sample a of 9 rather than 8 or 10 or whatever.
	//
	// This gives us output latency of 2 (== n), i.e. output samples i
	// and j will appear before the one in which input sample a is at
	// the centre of the filter.

	h := filterLength / 2
	n := roundUp(float64(filterLength-h) / float64(outputSpacing))

	phase, fill := (h+n*outputSpacing)%inputSpacing, (h+n*outputSpacing)/inputSpacing

	if DEBUG_R {
		fmt.Printf("initial phase %v (as %v %% %v), latency %v\n", phase, filterLength/2, inputSpacing, n)
	}

	return &Resampler{
		sourceRate,
		targetRate,
		gcd,
		peakToPole,

		filterLength,
		phase,
		n, /* latency */
		phaseData,

		make([]float64, fill, fill), /* buffer */
		0,                           /* bufferOrigin */
	}
}

func (r *Resampler) GetLatency() int {
	return r.latency
}

func (r *Resampler) Process(src []float64) []float64 {
	n := len(src)
	r.buffer = append(r.buffer, src...)

	maxout := roundUp(float64(n) * float64(r.targetRate) / float64(r.sourceRate))
	outidx := 0
	dst := make([]float64, maxout, maxout)

	if DEBUG_R {
		fmt.Printf("process: buf siz %v filt siz for phase %v %v\n", len(r.buffer), r.phase, len(r.phaseData[r.phase].filter))
	}

	scaleFactor := float64(r.targetRate/r.gcd) / r.peakToPole

	for outidx < maxout && len(r.buffer) >= len(r.phaseData[r.phase].filter)+r.bufferOrigin {
		dst[outidx] = scaleFactor * r.reconstructOne()
		outidx++
	}

	r.buffer = r.buffer[r.bufferOrigin:]
	r.bufferOrigin = 0

	if outidx < maxout {
		dst = dst[:outidx]
	}

	return dst
}

func kaiserForBandwidth(attenuation float64, bandwidth float64, sampleRate float64) KaiserWindow {
	transition := bandwidth * 2.0 * math.Pi / sampleRate
	length, beta := 0, 0.0

	if attenuation > 21.0 {
		length = 1 + roundUp((attenuation-7.95)/(2.285*transition))
		if attenuation > 50.0 {
			beta = 0.1102 * (attenuation - 8.7)
		} else {
			beta = 0.5842*math.Pow(attenuation-21.0, 0.4) + 0.07886*(attenuation-21.0)
		}
	} else {
		length = 1 + roundUp(5.79/transition)
		beta = 0
	}
	return KaiserWindow{length, beta}
}

func (r *Resampler) reconstructOne() float64 {
	v, n := 0.0, len(r.phaseData[r.phase].filter)

	if n+r.bufferOrigin > len(r.buffer) {
		panic("Ooops, can't reconstruct resampler phase")
	}

	for i := 0; i < n; i++ {
		// TODO - do as a for-each loop over r.phaseData[r.phase]
		v += r.buffer[r.bufferOrigin+i] * r.phaseData[r.phase].filter[i]
	}

	r.bufferOrigin += r.phaseData[r.phase].drop
	r.phase = r.phaseData[r.phase].nextPhase
	return v
}

func buildKaiserWindow(params KaiserWindow) []float64 {
	denominator := bessel0(params.beta)
	// even := (params.length % 2 == 0)
	window := make([]float64, params.length, params.length)

	upperBound1 := (params.length + 1) / 2 // even ? m_length/2 : (m_length+1)/2
	upperBound2 := (params.length) / 2     // even ? m_length/2 : (m_length-1)/2

	for i := 0; i < upperBound1; i++ {
		k := float64(2*i)/float64(params.length-1) - 1.0
		window[i] = bessel0(params.beta*math.Sqrt(1.0-k*k)) / denominator
	}
	for i := 0; i < upperBound2; i++ {
		// TODO - simplify, this just makes it symmetric...
		window[i+upperBound1] = window[params.length/2-i-1]
	}
	return window
}

func buildSincWindow(length int, p float64) []float64 {
	if length < 2 {
		panic("Sinc window too short!")
	}
	n0, n1 := 0, 0
	if length%2 == 0 {
		n0, n1 = length/2, length/2
	} else {
		n0, n1 = (length-1)/2, (length+1)/2
	}

	m := 2.0 * math.Pi / p

	window := make([]float64, length, length)

	for i := 0; i < n0; i++ {
		x := float64((length/2)-i) * m
		window[i] = math.Sin(x) / x
	}

	window[n0] = 1.0

	for i := 1; i < n1; i++ {
		x := float64(i) * m
		window[n0+i] = math.Sin(x) / x
	}
	return window
}
