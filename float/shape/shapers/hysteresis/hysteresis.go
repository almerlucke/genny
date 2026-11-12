// Package hysteresis provides two shapers with memory: Schmitt, a trigger with two thresholds, and
// Play, the backlash (play) operator. Both are rate independent: the output depends on the order in
// which the input visits values, not on how fast it moves.
//
// Schmitt trigger with thresholds low <= high: the state switches on when x >= high and off when
// x <= low and keeps its value in between. The output is the configured high or low value. It starts
// off.
//
// Play (backlash) operator with width w: the output y is a window of width w around the input. With
// h = w/2:
//
//	y' = x - h   if x > y + h
//	y' = x + h   if x < y - h
//	y' = y       otherwise
//
// so the output only follows the input after the input has moved through the dead band, and a
// triangle input draws a parallelogram loop in (x, y). It starts at 0. Non finite input is ignored by
// both (the previous output is returned).
package hysteresis

import (
	"fmt"
	"math"
)

// Schmitt is a Schmitt trigger
type Schmitt struct {
	low, high       float64
	outLow, outHigh float64
	on              bool
}

// NewSchmitt creates a Schmitt trigger with input thresholds low and high and output values outLow
// (state off) and outHigh (state on). It panics if low > high or a value is not finite
func NewSchmitt(low, high, outLow, outHigh float64) *Schmitt {
	for _, v := range []float64{low, high, outLow, outHigh} {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			panic(fmt.Sprintf("hysteresis: NewSchmitt: values must be finite, got %v", v))
		}
	}
	if low > high {
		panic(fmt.Sprintf("hysteresis: NewSchmitt: low %v is above high %v", low, high))
	}
	return &Schmitt{low: low, high: high, outLow: outLow, outHigh: outHigh}
}

// On is true if the trigger is in the high state
func (s *Schmitt) On() bool { return s.on }

// Shape updates the trigger with x and returns the output value
func (s *Schmitt) Shape(x float64) float64 {
	switch {
	case math.IsNaN(x):
	case x >= s.high:
		s.on = true
	case x <= s.low:
		s.on = false
	}
	if s.on {
		return s.outHigh
	}
	return s.outLow
}

// Reset switches the trigger off
func (s *Schmitt) Reset() { s.on = false }

// Play is the backlash operator
type Play struct {
	half float64
	y    float64
}

// NewPlay creates a backlash with the given width (the size of the dead band). It panics if width
// is negative or not finite. Width 0 passes the input
func NewPlay(width float64) *Play {
	if !(width >= 0) || math.IsInf(width, 0) {
		panic(fmt.Sprintf("hysteresis: NewPlay: width must be finite and not negative, got %v", width))
	}
	return &Play{half: width / 2}
}

// Shape moves the output towards x as the operator describes and returns it
func (p *Play) Shape(x float64) float64 {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return p.y
	}
	if x > p.y+p.half {
		p.y = x - p.half
	} else if x < p.y-p.half {
		p.y = x + p.half
	}
	return p.y
}

// Reset sets the output back to 0
func (p *Play) Reset() { p.y = 0 }
