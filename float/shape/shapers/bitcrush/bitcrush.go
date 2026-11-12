// Package bitcrush provides Crusher, a lo-fi shaper that reduces the bit depth and the sample rate of
// a signal in [-1,1].
//
// Bit depth: with b bits there are 2^b levels. The quantizer is mid-tread (0 is a level, as in two's
// complement audio) with L = 2^(b-1) steps per side:
//
//	y = clamp(round(x*L), -L, L-1) / L
//
// so the levels are -1, -1+1/L, ..., 0, ..., 1-1/L. The input is clamped to [-1,1] first, so the
// output lies in [-1, 1-1/L]. One bit has the levels -1 and 0.
//
// Sample rate reduction: the crushed value is taken at the first call and then held for factor calls,
// so the effective sample rate is divided by factor (the signal is not filtered first, which is the
// point: it aliases).
//
// Mix: out = (1-mix)*x + mix*crushed, with x the current (not held) input.
package bitcrush

import (
	"fmt"
	"math"
)

// Crusher reduces bit depth and sample rate. It keeps state (the held value), see Reset
type Crusher struct {
	levels float64 // L
	factor int
	mix    float64
	count  int
	held   float64
}

// New creates a Crusher. bits must be in 1..24, factor (the number of calls a value is held, 1 is
// no sample rate reduction) at least 1 and mix in [0,1] (1 is fully crushed). It panics otherwise
func New(bits, factor int, mix float64) *Crusher {
	if bits < 1 || bits > 24 {
		panic(fmt.Sprintf("bitcrush: bits must be in 1..24, got %d", bits))
	}
	if factor < 1 {
		panic(fmt.Sprintf("bitcrush: factor must be at least 1, got %d", factor))
	}
	if !(mix >= 0 && mix <= 1) {
		panic(fmt.Sprintf("bitcrush: mix must be in [0,1], got %v", mix))
	}
	return &Crusher{levels: float64(int(1) << (bits - 1)), factor: factor, mix: mix}
}

// Quantize quantizes x to the given bit depth (1..24) with the mid-tread quantizer of the package
// doc. It panics if bits is out of range
func Quantize(x float64, bits int) float64 {
	if bits < 1 || bits > 24 {
		panic(fmt.Sprintf("bitcrush: bits must be in 1..24, got %d", bits))
	}
	return quantize(x, float64(int(1)<<(bits-1)))
}

func quantize(x, l float64) float64 {
	if math.IsNaN(x) {
		return 0
	}
	return math.Max(-l, math.Min(l-1, math.Round(x*l))) / l
}

// Shape crushes x. Non finite input is read as 0
func (c *Crusher) Shape(x float64) float64 {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		x = 0
	}
	if c.count == 0 {
		c.held = quantize(x, c.levels)
	}
	c.count++
	if c.count >= c.factor {
		c.count = 0
	}
	return (1-c.mix)*x + c.mix*c.held
}

// Reset forgets the held value: the next call takes a new sample
func (c *Crusher) Reset() {
	c.count = 0
	c.held = 0
}
