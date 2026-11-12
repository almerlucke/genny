// Package deadzone provides shapers that treat small signals specially: a DeadZone that silences them
// and an Expander that attenuates them.
//
// Both are odd (the sign of the input is kept) and never produce NaN for non NaN input.
package deadzone

import "math"

// Mode selects what a DeadZone does with the signal outside of the zone
type Mode int

const (
	// Jump leaves the signal outside the zone untouched, so the output jumps from 0 to +-width at the edge
	Jump Mode = iota
	// Continuous subtracts the width from the magnitude, y = sign(x) * (|x| - width), so there is no jump
	Continuous
)

// DeadZone outputs 0 for inputs with a magnitude up to a width
type DeadZone struct {
	width float64
	mode  Mode
}

// NewDeadZone creates a dead zone: y = 0 for |x| <= width, otherwise x (Jump) or
// sign(x) * (|x| - width) (Continuous). Width 0 is the identity. Both modes are monotone, only Jump has a jump. It panics if width is negative
// or not finite, or the mode is unknown.
func NewDeadZone(width float64, mode Mode) *DeadZone {
	if width < 0 || math.IsNaN(width) || math.IsInf(width, 0) {
		panic("deadzone: width must be finite and not negative")
	}
	if mode != Jump && mode != Continuous {
		panic("deadzone: unknown mode")
	}
	return &DeadZone{width: width, mode: mode}
}

// Shape applies the dead zone to x
func (d *DeadZone) Shape(x float64) float64 {
	a := math.Abs(x)
	if a <= d.width {
		return 0
	}
	if d.mode == Continuous {
		return math.Copysign(a-d.width, x)
	}
	return x
}

// Expander is a downward expander: it makes signals below a threshold even quieter
type Expander struct {
	threshold, ratio float64
}

// NewExpander creates a downward expander. At and above the threshold the input passes unchanged. Below
// it each dB the level drops under the threshold becomes ratio dB, which in linear terms is
// y = sign(x) * threshold * (|x|/threshold)^ratio. The curve is continuous and monotone, a ratio of 1
// is the identity and a large ratio approaches a hard dead zone of width threshold. It panics if
// threshold is not positive and finite, or ratio is below 1 or not finite.
func NewExpander(threshold, ratio float64) *Expander {
	if !(threshold > 0) || math.IsInf(threshold, 0) {
		panic("deadzone: threshold must be positive and finite")
	}
	if !(ratio >= 1) || math.IsInf(ratio, 0) {
		panic("deadzone: ratio must be at least 1 and finite")
	}
	return &Expander{threshold: threshold, ratio: ratio}
}

// Shape expands x
func (e *Expander) Shape(x float64) float64 {
	a := math.Abs(x)
	if a >= e.threshold {
		return x
	}
	return math.Copysign(e.threshold*math.Pow(a/e.threshold, e.ratio), x)
}
