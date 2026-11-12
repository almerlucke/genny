package plot

import (
	"errors"
	"math"
	"math/bits"

	"github.com/almerlucke/genny/float"
)

// fft transforms re and im in place (radix-2 decimation in time). Their length must be a power of two
func fft(re, im []float64) {
	n := len(re)
	if n < 2 {
		return
	}
	shift := 64 - uint(bits.TrailingZeros(uint(n)))
	for i := range re {
		if j := int(bits.Reverse64(uint64(i)) >> shift); j > i {
			re[i], re[j] = re[j], re[i]
			im[i], im[j] = im[j], im[i]
		}
	}
	for size := 2; size <= n; size *= 2 {
		half := size / 2
		for k := 0; k < half; k++ {
			s, c := math.Sincos(-2 * math.Pi * float64(k) / float64(size))
			for i := k; i < n; i += size {
				j := i + half
				tr, ti := re[j]*c-im[j]*s, re[j]*s+im[j]*c
				re[j], im[j] = re[i]-tr, im[i]-ti
				re[i], im[i] = re[i]+tr, im[i]+ti
			}
		}
	}
}

// MagnitudeSpectrum returns the frequencies (in Hz, from 0 to half the sample rate) and the
// amplitudes of the samples. The samples are zero-padded to a power of two. The amplitudes are
// scaled so that a sine of amplitude A exactly on a bin shows as A (with or without the Hann
// window). It returns nothing for an empty slice or a sample rate that is not positive
func MagnitudeSpectrum(samples []float64, sampleRate float64, hann bool) (freqs, mags []float64) {
	n := len(samples)
	if n == 0 || !(sampleRate > 0) {
		return nil, nil
	}
	size := 1
	for size < n {
		size *= 2
	}
	size = max(size, 2)
	re, im := make([]float64, size), make([]float64, size)
	gain := 0.0
	for i, v := range samples {
		w := 1.0
		if hann && n > 1 {
			w = 0.5 - 0.5*math.Cos(2*math.Pi*float64(i)/float64(n))
		}
		re[i] = v * w
		gain += w
	}
	fft(re, im)
	if gain == 0 {
		gain = 1
	}
	freqs, mags = make([]float64, size/2+1), make([]float64, size/2+1)
	for k := range freqs {
		freqs[k] = float64(k) * sampleRate / float64(size)
		mags[k] = math.Hypot(re[k], im[k]) * 2 / gain
		if k == 0 || k == size/2 {
			mags[k] /= 2
		}
	}
	return freqs, mags
}

// SpectrumOptions are the options of NewSpectrum
type SpectrumOptions struct {
	// Hann applies a Hann window, which tames the leakage of frequencies between two bins
	Hann bool
	// DB gives decibels (20 log10 of the amplitude, 0 dB is amplitude 1) instead of amplitudes
	DB bool
	// FloorDB is the lowest value in dB (default -120)
	FloorDB float64
	// LogFreq leaves out the 0 Hz bin, which cannot be on a logarithmic axis. Also set Scale to Log
	// on the x axis of the plot
	LogFreq bool
}

// NewSpectrum returns the magnitude spectrum of the samples as a line of frequency in Hz against
// amplitude (or dB). It fails if there are no samples, the sample rate is not positive or a sample
// is not finite
func NewSpectrum(samples []float64, sampleRate float64, opts SpectrumOptions) (*Line, error) {
	if len(samples) == 0 {
		return nil, errors.New("plot: no samples")
	}
	if !(sampleRate > 0) {
		return nil, errors.New("plot: the sample rate must be positive")
	}
	for _, v := range samples {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, errors.New("plot: sample is not finite")
		}
	}
	freqs, mags := MagnitudeSpectrum(samples, sampleRate, opts.Hann)
	floor := opts.FloorDB
	if floor >= 0 {
		floor = -120
	}
	var pts XYs
	for k, f := range freqs {
		if k == 0 && opts.LogFreq {
			continue
		}
		v := mags[k]
		if opts.DB {
			v = math.Max(20*math.Log10(math.Max(v, 1e-300)), floor)
		}
		pts = append(pts, XY{f, v})
	}
	return NewLine(pts)
}

// SpectrumOf is the spectrum of the first n values of the first dimension of a generator
func SpectrumOf(g float.FrameGenerator, n int, sampleRate float64, opts SpectrumOptions) (*Line, error) {
	samples := make([]float64, 0, max(n, 0))
	for i := 0; i < n && !g.Done(); i++ {
		if frame := g.Generate(); len(frame) > 0 {
			samples = append(samples, frame[0])
		}
	}
	return NewSpectrum(samples, sampleRate, opts)
}

// Shaper is something that shapes a signal, like the shapers of the shape package
type Shaper interface{ Shape(float64) float64 }

// ShaperSpectrum is the spectrum of a sine of amplitude amp and frequency freq (in Hz) that is sent
// through the shaper: n samples at the sample rate. Choose freq as a multiple of sampleRate/N, where
// N is n rounded up to a power of two, to keep the sine on a bin (or use the Hann window)
func ShaperSpectrum(s Shaper, freq, amp float64, n int, sampleRate float64, opts SpectrumOptions) (*Line, error) {
	samples := make([]float64, max(n, 0))
	for i := range samples {
		samples[i] = s.Shape(amp * math.Sin(2*math.Pi*freq*float64(i)/sampleRate))
	}
	return NewSpectrum(samples, sampleRate, opts)
}
