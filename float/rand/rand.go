// Package rand provides Rand, a generator of uniformly distributed random floats between a low and a
// high value.
package rand

import "github.com/almerlucke/genny"

// Rand generates uniform random values between low and high
type Rand struct {
	low  float64
	high float64
}

// New creates a Rand, low and high may be given in either order
func New(low float64, high float64) *Rand {
	return &Rand{
		low:  low,
		high: high,
	}
}

func (r *Rand) SetLow(low float64) {
	r.low = low
}

func (r *Rand) SetHigh(high float64) {
	r.high = high
}

func (r *Rand) Generate() float64 {
	low := r.low
	high := r.high

	if low > high {
		low, high = high, low
	}

	return genny.Float64()*(high-low) + low
}

func (r *Rand) Continuous() bool {
	return true
}

func (r *Rand) Reset() {
}

func (r *Rand) Done() bool {
	return false
}
