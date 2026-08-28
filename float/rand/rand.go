package rand

import "math/rand"

type Rand struct {
	low  float64
	high float64
}

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

	return rand.Float64()*(high-low) + low
}

func (r *Rand) Continuous() bool {
	return true
}

func (r *Rand) Reset() {
}

func (r *Rand) Done() bool {
	return false
}
