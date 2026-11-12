// Package harmonics provides an additive shaper: the transfer curve is a sum of sines, so a sine input
// of the right amplitude comes out as a chosen mixture of harmonics.
//
// The curve is y(x) = sum over k = 1..n of a_k * sin(k*pi*x + phase_k). The input range -1..1 covers
// exactly one full period of the fundamental (sin(pi*x) runs from -pi to pi), so a sine input of
// amplitude 1 sweeps the whole curve and the output contains the input frequency f and 2f .. nf.
// With zero phases every term is odd, so the shaper is odd; non zero phases add even symmetry.
//
// The result is finite for every finite input; an input that overflows when multiplied by k*pi gives 0.
package harmonics

import "math"

// Harmonics is an additive shaper
type Harmonics struct {
	amps, phases []float64
	scale        float64
}

// New creates an additive shaper from the amplitudes of the harmonics 1..len(amps) and optionally their
// phases in radians (nil, or shorter than amps: the missing phases are 0). With normalise the
// result is scaled so that the largest |y| for x in [-1, 1] is 1 (measured on a fine grid, so exact to
// about 1e-6). It panics if amps is empty, if amps or phases hold a NaN or infinite value, if there are
// more phases than amplitudes, or if normalise is set and the curve is zero everywhere.
func New(amps, phases []float64, normalise bool) *Harmonics {
	if len(amps) == 0 {
		panic("harmonics: no amplitudes")
	}
	if len(phases) > len(amps) {
		panic("harmonics: more phases than amplitudes")
	}
	for _, v := range append(append([]float64(nil), amps...), phases...) {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			panic("harmonics: amplitudes and phases must be finite")
		}
	}
	h := &Harmonics{amps: append([]float64(nil), amps...), phases: make([]float64, len(amps)), scale: 1}
	copy(h.phases, phases)
	if normalise {
		const n = 1 << 14
		peak := 0.0
		for i := 0; i <= n; i++ {
			peak = math.Max(peak, math.Abs(h.Shape(-1+2*float64(i)/n)))
		}
		if peak == 0 {
			panic("harmonics: cannot normalise a curve that is zero")
		}
		h.scale = 1 / peak
	}
	return h
}

// Shape evaluates the sum of sines at x
func (h *Harmonics) Shape(x float64) float64 {
	sum := 0.0
	for i, a := range h.amps {
		k := float64(i + 1)
		arg := k * math.Pi * x
		if math.IsNaN(arg) || math.IsInf(arg, 0) {
			return 0
		}
		sum += a * math.Sin(arg+h.phases[i])
	}
	return sum * h.scale
}
