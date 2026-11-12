// Package rectify provides rectifiers: shapers that remove or flip the negative half of a signal.
//
// Rectifying a signal adds even harmonics (full wave doubles the frequency of a sine) and, for half
// wave and diode, a DC offset. The output is finite for any non NaN input.
package rectify

import "math"

// Mode selects the rectifier
type Mode int

const (
	// HalfWave is y = max(x, 0)
	HalfWave Mode = iota
	// FullWave is y = |x|
	FullWave
	// Diode is a diode: nothing passes until the threshold, with a soft knee, see NewDiode
	Diode
)

// Rectifier rectifies its input
type Rectifier struct {
	mode      Mode
	threshold float64
	knee      float64
}

// New creates a half wave or full wave rectifier. It panics for any other mode, a Diode needs
// parameters and is made with NewDiode.
func New(mode Mode) *Rectifier {
	if mode != HalfWave && mode != FullWave {
		panic("rectify: New takes HalfWave or FullWave, use NewDiode for a diode")
	}
	return &Rectifier{mode: mode}
}

// NewDiode creates a diode rectifier. With u = x - threshold the output is 0 for u <= -knee, u for
// u >= knee and the quadratic (u+knee)^2 / (4*knee) in between, which joins both parts with a
// continuous slope (a soft knee, like the exponential turn on of a real diode). With knee 0 this is
// the hard max(x - threshold, 0). It panics if threshold or knee is negative or not finite.
func NewDiode(threshold, knee float64) *Rectifier {
	for _, v := range []float64{threshold, knee} {
		if v < 0 || math.IsNaN(v) || math.IsInf(v, 0) {
			panic("rectify: threshold and knee must be finite and not negative")
		}
	}
	return &Rectifier{mode: Diode, threshold: threshold, knee: knee}
}

// Shape rectifies x
func (r *Rectifier) Shape(x float64) float64 {
	switch r.mode {
	case FullWave:
		return math.Abs(x)
	case Diode:
		u := x - r.threshold
		switch {
		case u <= -r.knee:
			return 0
		case u >= r.knee:
			return u
		}
		return (u + r.knee) * (u + r.knee) / (4 * r.knee)
	}
	return math.Max(x, 0)
}
