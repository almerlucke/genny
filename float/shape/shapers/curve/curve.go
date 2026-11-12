// Package curve provides power curves with a single tension parameter, for bending a linear ramp or
// a bipolar signal.
//
// The tension follows the shape parameter of the adsr envelope (AttackShape and friends): 0 is linear,
// and the exponent applied is 1/(1+2*tension) for a positive tension and 1+2*|tension| for a negative
// one, so tension = 0.5 gives x^(1/2) and tension = -0.5 gives x^2. A positive tension bends the curve
// above the straight line (fast start, like a rising adsr stage with a positive shape), a negative
// tension below it (slow start). Unlike the envelope the tension is limited to (-1, 1), which
// keeps the exponent between 1/3 and 3.
package curve

import "math"

// Curve is a power curve
type Curve struct {
	exp     float64
	bipolar bool
}

// Exponent returns the exponent for a tension, 1 for tension 0
func Exponent(tension float64) float64 {
	if tension > 0 {
		return 1 / (1 + 2*tension)
	}
	return 1 - 2*tension
}

// NewUnipolar creates a curve y = x^exponent on [0, 1], x is clamped into [0, 1] first so 0 maps to 0
// and 1 to 1. It is monotone. It panics if tension is not in (-1, 1).
func NewUnipolar(tension float64) *Curve {
	return &Curve{exp: Exponent(checkTension(tension))}
}

// NewBipolar creates a sign preserving curve y = sign(x) * |x|^exponent on [-1, 1], x is clamped into
// [-1, 1] first. It is odd and monotone, 0 maps to 0 and +-1 to +-1. It panics if tension is not in
// (-1, 1).
func NewBipolar(tension float64) *Curve {
	return &Curve{exp: Exponent(checkTension(tension)), bipolar: true}
}

// Shape bends x
func (c *Curve) Shape(x float64) float64 {
	if c.bipolar {
		x = math.Max(-1, math.Min(1, x))
		return math.Copysign(math.Pow(math.Abs(x), c.exp), x)
	}
	x = math.Max(0, math.Min(1, x))
	return math.Pow(x, c.exp)
}

func checkTension(t float64) float64 {
	if !(math.Abs(t) < 1) {
		panic("curve: tension must be in (-1, 1)")
	}
	return t
}
