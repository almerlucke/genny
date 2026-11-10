// Package lissajous generates Lissajous curves: two sine waves with a frequency ratio and a phase
// difference, as a two dimensional frame generator. Use the two outputs as two related modulation
// sources (pitch and pan, cutoff and resonance)
package lissajous

import "math"

// Lissajous is a float.FrameGenerator with two dimensions, x = sin(a t + delta) and y = sin(b t)
type Lissajous struct {
	a, b, delta float64
	steps       int
	step        int
	out         [2]float64
}

// New creates a curve with frequencies a and b (ratio 3:2 gives a classic knot) and a phase difference
// delta in radians, one full period of t takes steps steps. It panics if steps is less than 1
func New(a, b, delta float64, steps int) *Lissajous {
	if steps < 1 {
		panic("lissajous: steps must be at least 1")
	}
	return &Lissajous{a: a, b: b, delta: delta, steps: steps}
}

// Generate returns x and y, a slice that is reused between calls
func (l *Lissajous) Generate() []float64 {
	t := 2 * math.Pi * float64(l.step) / float64(l.steps)
	l.out[0] = math.Sin(l.a*t + l.delta)
	l.out[1] = math.Sin(l.b * t)
	l.step = (l.step + 1) % l.steps
	return l.out[:]
}

func (l *Lissajous) Dimensions() int { return 2 }

func (l *Lissajous) Continuous() bool { return true }

func (l *Lissajous) Done() bool { return false }

func (l *Lissajous) Reset() { l.step = 0 }
