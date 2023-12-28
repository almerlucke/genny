package ramp

import "math"

// Ramp represents an exponential ramp
type Ramp struct {
	min   float64
	dev   float64
	exp   float64
	steps int
	index int
	acc   float64
	inc   float64
	done  bool
}

// New creates a new ramp
func New(steps int, min float64, max float64, exp float64) *Ramp {
	return &Ramp{
		min:   min,
		dev:   max - min,
		steps: steps,
		exp:   exp,
		inc:   1.0 / float64(steps-1),
	}
}

// Reset the ramp
func (r *Ramp) Reset() {
	r.acc = 0
	r.index = 0
	r.done = false
}

// NextValue generates a ramp value, return false if end of ramp is reached
func (r *Ramp) NextValue() (v float64) {
	v = math.Pow(r.acc, r.exp)*r.dev + r.min

	if r.done {
		return
	}

	r.acc += r.inc
	r.index++
	r.done = r.index >= r.steps

	return
}

// Continuous will always return false
func (r *Ramp) Continuous() bool {
	return false
}

// Done returns true if ramp is done
func (r *Ramp) Done() bool {
	return r.done
}
