// Package onepole provides a one pole smoothing filter as a shaper. Every call to Shape is one
// sample. The lowpass is the exponential smoother
//
//	y[n] = y[n-1] + a*(x[n] - y[n-1])
//
// with a in (0,1], a=1 passes the input and a small a smooths heavily. The highpass is the input
// minus the lowpass, x[n] - y[n]. The coefficient follows from a time constant tau (in samples) as
//
//	a = 1 - exp(-1/tau)
//
// The step response of the lowpass reaches 1-1/e (63%) after tau samples. The -3 dB cutoff is at
// about a*fs/(2*pi) for small a. Since 0 < a <= 1 the filter is always stable and for a bounded
// input the output stays within the range of the input and the starting state 0. Non finite input
// is ignored (the last output is returned).
package onepole

import (
	"fmt"
	"math"
)

// Filter is a one pole filter. It keeps state, see Reset
type Filter struct {
	a        float64
	highpass bool
	y        float64 // lowpass state
	last     float64 // last output
}

func checkCoeff(a float64) {
	if !(a > 0 && a <= 1) {
		panic(fmt.Sprintf("onepole: coefficient must be in (0,1], got %v", a))
	}
}

// New creates a lowpass with coefficient a in (0,1]. It panics otherwise
func New(a float64) *Filter {
	checkCoeff(a)
	return &Filter{a: a}
}

// NewHighpass creates a highpass with coefficient a in (0,1]. It panics otherwise
func NewHighpass(a float64) *Filter {
	checkCoeff(a)
	return &Filter{a: a, highpass: true}
}

// Coefficient converts a time constant in samples to the coefficient a = 1-exp(-1/tau). Zero gives
// 1 (no smoothing). It panics if tau is negative or not a number. A huge tau gives a tiny but positive a
func Coefficient(tau float64) float64 {
	if !(tau >= 0) {
		panic(fmt.Sprintf("onepole: time constant must not be negative, got %v", tau))
	}
	if tau == 0 {
		return 1
	}
	a := -math.Expm1(-1 / tau)
	if a <= 0 { // tau so large that 1/tau underflows
		a = math.SmallestNonzeroFloat64
	}
	return a
}

// NewSamples creates a lowpass with time constant tau in samples (calls to Shape)
func NewSamples(tau float64) *Filter { return New(Coefficient(tau)) }

// NewTime creates a lowpass with time constant seconds at a sample rate. It panics if seconds is
// negative or sampleRate is not positive
func NewTime(seconds, sampleRate float64) *Filter {
	return New(Coefficient(samples(seconds, sampleRate)))
}

// NewHighpassTime creates a highpass with time constant seconds at a sample rate (its cutoff is
// where the lowpass with the same time constant has its cutoff). It panics like NewTime
func NewHighpassTime(seconds, sampleRate float64) *Filter {
	return NewHighpass(Coefficient(samples(seconds, sampleRate)))
}

func samples(seconds, sampleRate float64) float64 {
	if !(seconds >= 0) || !(sampleRate > 0) || math.IsInf(seconds, 0) || math.IsInf(sampleRate, 0) {
		panic(fmt.Sprintf("onepole: need finite seconds >= 0 and sampleRate > 0, got %v and %v", seconds, sampleRate))
	}
	return seconds * sampleRate
}

// Shape filters one sample
func (f *Filter) Shape(x float64) float64 {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return f.last
	}
	f.y += f.a * (x - f.y)
	if f.highpass {
		f.last = x - f.y
	} else {
		f.last = f.y
	}
	return f.last
}

// Reset sets the state back to 0
func (f *Filter) Reset() {
	f.y = 0
	f.last = 0
}
