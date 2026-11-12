// Package phasedist provides Casio CZ style phase distortion. The input is a phase in [0,1) (any
// value is wrapped into that range). The phase is first warped by a piecewise linear function g
// and then read by a cosine:
//
//	y = cos(2*pi*g(p))
//
// With g(p) = p this is a plain cosine (the sine at quarter cycle offset), the distortion amount d
// in [0,1] bends g away from the identity. Because the warp is made of line segments the result is a
// sweep from a sine to a saw, square, pulse or resonant waveform without aliasing from a hard edge
// until d approaches 1. All waveforms stay within [-1,1], start a cycle at p=0 with y=1 and never
// return NaN: a non finite phase is read as phase 0.
//
// Edges are never made infinitely steep: the slope of a warp segment is limited so d=1 gives an edge
// of 1% of a cycle (the minimum edge is minEdge).
package phasedist

import (
	"fmt"
	"math"
)

// minEdge is the shortest edge (as a fraction of a cycle) a distortion makes, which keeps the
// slopes of the warp finite
const minEdge = 0.01

// edge maps the amount d in [0,1] to the length factor of an edge: 1 for d=0 down to minEdge for d=1
func edge(d float64) float64 { return 1 - (1-minEdge)*d }

func wrap(p float64) float64 {
	if math.IsNaN(p) || math.IsInf(p, 0) {
		return 0
	}
	p -= math.Floor(p)
	if p >= 1 { // p was a tiny negative number
		p = 0
	}
	return p
}

func checkAmount(name string, d float64) {
	if !(d >= 0 && d <= 1) {
		panic(fmt.Sprintf("phasedist: %s: distortion amount must be in [0,1], got %v", name, d))
	}
}

// Kind is the waveform family of a Shaper
type Kind int

// The waveform families
const (
	// Saw warps with a break point that moves from 0.5 towards 0: g rises fast to 0.5 then slowly
	Saw Kind = iota
	// Square holds the cosine at its extremes and makes fast edges in between
	Square
	// Pulse reads one full compressed cosine cycle in the first part of the period, then rests at 1
	Pulse
	// Resonant multiplies a falling window with a higher frequency cosine
	Resonant
)

// Shaper is a phase distortion waveform: it maps a phase to a sample
type Shaper struct {
	kind     Kind
	d        float64
	width    float64 // square only: fraction of the cycle for the first half
	harmonic float64 // resonant only
}

// NewSaw creates the saw variant. The break point of the warp is at b = 0.5*e, with e = 1-0.99d:
//
//	g(p) = 0.5*p/b                      for p < b
//	g(p) = 0.5 + 0.5*(p-b)/(1-b)        for p >= b
//
// so the cosine falls from 1 to -1 within b and rises slowly in the rest of the cycle. d=0 is the
// cosine, d=1 an almost perfect falling saw. It panics if d is not in [0,1]
func NewSaw(d float64) *Shaper {
	checkAmount("NewSaw", d)
	return &Shaper{kind: Saw, d: d, width: 0.5}
}

// NewSquare creates the square variant with a duty of 50%, see NewSquareWidth
func NewSquare(d float64) *Shaper { return NewSquareWidth(d, 0.5) }

// NewSquareWidth creates a square variant where the first (low) part of the cycle takes the fraction
// width (in (0,1)) of the period. With e = 1-0.99d the warp is
//
//	g(p) = 0.5*min(p/(w*e), 1)                      for p < w
//	g(p) = 0.5 + 0.5*min((p-w)/((1-w)*e), 1)        for p >= w
//
// so g rushes to 0.5 (cosine value -1), rests there, rushes to 1 (cosine value 1) and rests. d=0 with
// width 0.5 is the cosine, d=1 an almost perfect square. It panics if d is not in [0,1] or width not in (0,1)
func NewSquareWidth(d, width float64) *Shaper {
	checkAmount("NewSquareWidth", d)
	if !(width > 0 && width < 1) {
		panic(fmt.Sprintf("phasedist: NewSquareWidth: width must be in (0,1), got %v", width))
	}
	return &Shaper{kind: Square, d: d, width: width}
}

// NewPulse creates the pulse variant: with w = 0.01+0.99(1-d) the warp is g(p) = min(p/w, 1), a full
// cosine cycle in the first fraction w of the period and rest (value 1) afterwards. d=0 is the
// cosine, d=1 a narrow pulse of a 1% period cycle. It panics if d is not in [0,1]
func NewPulse(d float64) *Shaper {
	checkAmount("NewPulse", d)
	return &Shaper{kind: Pulse, d: d, width: 0.5}
}

// NewResonant creates the CZ resonance variant. A falling saw window is multiplied with a cosine of
// a higher frequency, which gives the sweeping resonant peak of the CZ:
//
//	y = (1 - d*p) * cos(2*pi*h*p),  h = 1 + d*(harmonic-1)
//
// d=0 is the cosine, d=1 is (1-p)*cos(2*pi*harmonic*p), which is continuous at the wrap since the
// window reaches 0. In between (and for a non integer h) there is a small step at the wrap, as on
// the hardware. harmonic is the frequency of the resonance in multiples of the fundamental and must
// be at least 1. It panics if d is not in [0,1] or harmonic is below 1 or not finite
func NewResonant(d, harmonic float64) *Shaper {
	checkAmount("NewResonant", d)
	if !(harmonic >= 1) || math.IsInf(harmonic, 0) {
		panic(fmt.Sprintf("phasedist: NewResonant: harmonic must be finite and at least 1, got %v", harmonic))
	}
	return &Shaper{kind: Resonant, d: d, harmonic: harmonic}
}

// Warp returns the warped phase g(p) in [0,1] for the saw, square and pulse variants. For the
// resonant variant, which has no warp, it returns the wrapped phase
func (s *Shaper) Warp(p float64) float64 {
	p = wrap(p)
	e := edge(s.d)
	switch s.kind {
	case Saw:
		b := 0.5 * e
		if p < b {
			return 0.5 * p / b
		}
		return 0.5 + 0.5*(p-b)/(1-b)
	case Square:
		w := s.width
		if p < w {
			return 0.5 * math.Min(p/(w*e), 1)
		}
		return 0.5 + 0.5*math.Min((p-w)/((1-w)*e), 1)
	case Pulse:
		return math.Min(p/e, 1)
	}
	return p
}

// Shape reads the waveform at the phase x
func (s *Shaper) Shape(x float64) float64 {
	if s.kind == Resonant {
		p := wrap(x)
		h := 1 + s.d*(s.harmonic-1)
		return (1 - s.d*p) * math.Cos(2*math.Pi*h*p)
	}
	return math.Cos(2 * math.Pi * s.Warp(x))
}
