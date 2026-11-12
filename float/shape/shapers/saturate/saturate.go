// Package saturate provides saturators: smooth, odd (or, for the tube, deliberately asymmetric) soft
// limiters that squeeze a growing input towards a ceiling and add harmonics on the way.
//
// All saturators are normalised so that the input +-1 maps to exactly +-1 whatever the drive, which
// makes the drive a pure amount of saturation: a small drive is almost linear, a large drive is almost
// a hard clip at +-1 (for inputs inside [-1, 1]). Inputs beyond 1 go on to a ceiling slightly above 1.
// The output is finite for any finite or infinite input.
package saturate

import "math"

// Mode selects the saturation curve
type Mode int

const (
	// Tanh is y = tanh(drive*x) / tanh(drive)
	Tanh Mode = iota
	// Atan is y = atan(drive*x) / atan(drive), the same curve as clip.Soft
	Atan
	// Logistic is the logistic sigmoid mapped to -1..1, 2/(1+exp(-drive*x)) - 1, divided by its value at
	// x = 1. It equals tanh(drive*x/2) / tanh(drive/2).
	Logistic
)

// Saturator saturates its input with one of the modes
type Saturator struct {
	mode  Mode
	drive float64
	norm  float64
}

// New creates a Saturator. It panics if drive is not positive and finite or the mode is unknown.
func New(mode Mode, drive float64) *Saturator {
	checkDrive(drive)
	s := &Saturator{mode: mode, drive: drive}
	switch mode {
	case Tanh:
		s.norm = 1 / math.Tanh(drive)
	case Atan:
		s.norm = 1 / math.Atan(drive)
	case Logistic:
		s.norm = 1 / math.Tanh(drive/2)
	default:
		panic("saturate: unknown mode")
	}
	return s
}

// Shape saturates x
func (s *Saturator) Shape(x float64) float64 {
	switch s.mode {
	case Atan:
		return s.norm * math.Atan(s.drive*x)
	case Logistic:
		return s.norm * math.Tanh(s.drive*x/2)
	default:
		return s.norm * math.Tanh(s.drive*x)
	}
}

// Tube is an asymmetric, tube-like saturator
type Tube struct {
	drive, bias float64
	offset      float64
	norm        float64
}

// NewTube creates an asymmetric saturator y = (tanh(drive*(x+bias)) - tanh(drive*bias)) / n. The bias
// moves the signal along the curve, so the positive and negative half waves are squeezed differently,
// which adds even harmonics. Subtracting tanh(drive*bias) removes the static offset (0 maps to 0) but
// not the DC a bias causes in a running signal, for that follow it with a high pass filter. n is the
// larger of |f(1)| and |f(-1)| so that the output stays inside [-1, 1] for inputs in [-1, 1]; with
// bias 0 this is exactly the Tanh saturator. It panics if drive is not positive and finite, or bias
// is not strictly between -1 and 1 (beyond that the whole input range sits on a flat side of the curve).
func NewTube(drive, bias float64) *Tube {
	checkDrive(drive)
	if !(math.Abs(bias) < 1) {
		panic("saturate: bias must be in (-1, 1)")
	}
	t := &Tube{drive: drive, bias: bias, offset: math.Tanh(drive * bias)}
	n := math.Max(math.Abs(math.Tanh(drive*(1+bias))-t.offset), math.Abs(math.Tanh(drive*(bias-1))-t.offset))
	if n == 0 {
		n = 1
	}
	t.norm = 1 / n
	return t
}

// Shape saturates x
func (t *Tube) Shape(x float64) float64 {
	return t.norm * (math.Tanh(t.drive*(x+t.bias)) - t.offset)
}

func checkDrive(drive float64) {
	if !(drive > 0) || math.IsInf(drive, 0) {
		panic("saturate: drive must be positive and finite")
	}
}
