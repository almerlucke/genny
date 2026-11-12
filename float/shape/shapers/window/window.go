// Package window provides window functions as shapers. The input is a phase x in [0,1], the position
// in the window, and the output is the window value (usually in [0,1]). Outside [0,1] a window is 0.
//
// The windows are symmetric by default: w(0) and w(1) are the two end points of the window and
// w(x) = w(1-x). This is the form to use for filter design and for windowing a block of N samples
// when the input is n/(N-1). With the Periodic option the input is wrapped (x=1 is the same as x=0) and
// the window is one period of a periodic function, the form for spectral analysis and overlap add
// when the input is n/N (the symmetric window of N+1 points without its last point).
//
// With u = 2x-1 in [-1,1] the formulas are:
//
//	Hann      0.5 - 0.5 cos(2 pi x)
//	Hamming   0.54 - 0.46 cos(2 pi x)
//	Blackman  0.42 - 0.5 cos(2 pi x) + 0.08 cos(4 pi x)
//	Tukey     cosine tapered ends of a fraction alpha/2 each and a flat middle
//	Gaussian  exp(-0.5 (u/sigma)^2)
//	Kaiser    I0(beta sqrt(1-u^2)) / I0(beta)
//	Welch     1 - u^2
//	Triangle  1 - |u|
//
// where I0 is the modified Bessel function of the first kind and order 0 (see I0).
package window

import (
	"fmt"
	"math"
)

// Window is a window function shaper
type Window struct {
	f        func(x float64) float64
	periodic bool
}

// Periodic returns a copy of the window that treats its input as a phase: the input is wrapped
// into [0,1) so the window repeats and is never 0 outside the range
func (w *Window) Periodic() *Window { return &Window{f: w.f, periodic: true} }

// Shape returns the window value at x. Non finite input gives 0
func (w *Window) Shape(x float64) float64 {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return 0
	}
	if w.periodic {
		x -= math.Floor(x)
	} else if x < 0 || x > 1 {
		return 0
	}
	return w.f(x)
}

// Hann is the raised cosine window, 0 at the ends
func Hann() *Window {
	return &Window{f: func(x float64) float64 { return 0.5 - 0.5*math.Cos(2*math.Pi*x) }}
}

// Hamming is the raised cosine window on a pedestal of 0.08 at the ends
func Hamming() *Window {
	return &Window{f: func(x float64) float64 { return 0.54 - 0.46*math.Cos(2*math.Pi*x) }}
}

// Blackman is the three term cosine window, about 0 at the ends
func Blackman() *Window {
	return &Window{f: func(x float64) float64 {
		return 0.42 - 0.5*math.Cos(2*math.Pi*x) + 0.08*math.Cos(4*math.Pi*x)
	}}
}

// Tukey is the tapered cosine window: alpha is the fraction of the window that is tapered, 0 gives a
// rectangular window and 1 the Hann window. It panics if alpha is not in [0,1]
func Tukey(alpha float64) *Window {
	if !(alpha >= 0 && alpha <= 1) {
		panic(fmt.Sprintf("window: Tukey: alpha must be in [0,1], got %v", alpha))
	}
	return &Window{f: func(x float64) float64 {
		if alpha == 0 {
			return 1
		}
		switch {
		case x < alpha/2:
			return 0.5 * (1 - math.Cos(2*math.Pi*x/alpha))
		case x > 1-alpha/2:
			return 0.5 * (1 - math.Cos(2*math.Pi*(1-x)/alpha))
		}
		return 1
	}}
}

// Gaussian is the bell window exp(-0.5 (u/sigma)^2) with u in [-1,1] from edge to edge, so sigma is
// the standard deviation in units of the half width (0.4 is a common value). It panics if sigma is
// not positive and finite
func Gaussian(sigma float64) *Window {
	if !(sigma > 0) || math.IsInf(sigma, 0) {
		panic(fmt.Sprintf("window: Gaussian: sigma must be positive and finite, got %v", sigma))
	}
	return &Window{f: func(x float64) float64 {
		u := (2*x - 1) / sigma
		return math.Exp(-0.5 * u * u)
	}}
}

// maxBeta keeps I0(beta) within the range of a float64
const maxBeta = 500

// Kaiser is the Kaiser-Bessel window, beta sets the trade off between main lobe width and side lobes
// (beta 0 is rectangular, about 5 resembles Hamming, 8.6 resembles Blackman). It panics if beta is
// negative or above 500
func Kaiser(beta float64) *Window {
	if !(beta >= 0 && beta <= maxBeta) {
		panic(fmt.Sprintf("window: Kaiser: beta must be in [0,%d], got %v", maxBeta, beta))
	}
	norm := 1 / I0(beta)
	return &Window{f: func(x float64) float64 {
		u := 2*x - 1
		return I0(beta*math.Sqrt(math.Max(1-u*u, 0))) * norm
	}}
}

// Welch is the parabolic window 1-u^2
func Welch() *Window {
	return &Window{f: func(x float64) float64 {
		u := 2*x - 1
		return 1 - u*u
	}}
}

// Triangle is the window 1-|u|, 0 at the ends and 1 in the middle
func Triangle() *Window {
	return &Window{f: func(x float64) float64 { return 1 - math.Abs(2*x-1) }}
}

// I0 is the modified Bessel function of the first kind of order 0, computed with its power series
// I0(x) = sum_k ((x/2)^k / k!)^2, which converges for every x (the terms are all positive so there
// is no cancellation). It is even, so I0(-x) = I0(x), and overflows to +Inf above about 713
func I0(x float64) float64 {
	q := x * x / 4
	sum, term := 1.0, 1.0
	for k := 1; k < 2000; k++ {
		term *= q / float64(k*k)
		sum += term
		if term < sum*1e-17 {
			break
		}
	}
	return sum
}
