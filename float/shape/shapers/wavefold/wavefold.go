// Package wavefold provides wave folders: shapers that, instead of clipping a signal that grows beyond
// +-1, fold it back so that it keeps moving inside [-1, 1] and gains harmonics the louder it gets.
//
// Sine folds with y = sin(gain * x * pi/2), triangle folds by reflecting the signal at +1 and -1 as
// often as needed. Both are odd, bounded by [-1, 1] and map 0 to 0. For finite input the output is
// finite; an input that overflows after the gain is applied gives 0.
package wavefold

import "math"

// Sine is a sine wave folder
type Sine struct {
	k float64
}

// NewSine creates a sine folder y = sin(gain * x * pi/2). With gain 1 the input range [-1, 1] covers
// a quarter sine either side of zero and the shaper is monotone; every extra 1 of gain adds a
// fold at the peaks of the sine. It panics if gain is negative or not finite.
func NewSine(gain float64) *Sine {
	checkGain(gain)
	return &Sine{k: gain * math.Pi / 2}
}

// Shape folds x
func (s *Sine) Shape(x float64) float64 {
	y := math.Sin(s.k * x)
	if math.IsNaN(y) {
		return 0
	}
	return y
}

// Triangle is a triangle wave folder
type Triangle struct {
	gain float64
}

// NewTriangle creates a triangle folder: the input is multiplied by gain and then reflected at +-1,
// so u = gain*x gives y = u for |u| <= 1, 2-u for 1 <= u <= 3, u-4 for 3 <= u <= 5 and so on, periodic
// with period 4 in u. It panics if gain is negative or not finite.
func NewTriangle(gain float64) *Triangle {
	checkGain(gain)
	return &Triangle{gain: gain}
}

// Shape folds x
func (t *Triangle) Shape(x float64) float64 {
	u := t.gain * x
	if math.IsNaN(u) || math.IsInf(u, 0) {
		return 0
	}
	m := math.Mod(u+1, 4)
	if m < 0 {
		m += 4
	}
	if m > 2 {
		m = 4 - m
	}
	return m - 1
}

func checkGain(gain float64) {
	if gain < 0 || math.IsNaN(gain) || math.IsInf(gain, 0) {
		panic("wavefold: gain must be finite and not negative")
	}
}
