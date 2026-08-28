package ramp

import "math"

// Ramp represents an exponential ramp
type Ramp struct {
	min        float64
	dev        float64
	exp        float64
	steps      int
	index      int
	acc        float64
	inc        float64
	done       bool
	continuous bool
}

// New creates a new ramp
func New(steps int, min float64, max float64, exp float64) *Ramp {
	return newRamp(steps, min, max, exp, false)
}

// NewForever creates a continuous ramp that stays on end value forever
func NewForever(steps int, min float64, max float64, exp float64) *Ramp {
	return newRamp(steps, min, max, exp, true)
}

func newRamp(steps int, min float64, max float64, exp float64, cont bool) *Ramp {
	return &Ramp{
		min:        min,
		dev:        max - min,
		steps:      steps,
		exp:        exp,
		inc:        1.0 / float64(steps-1),
		continuous: cont,
	}
}

func (r *Ramp) MakeContinuous(c bool) *Ramp {
	r.continuous = c
	return r
}

// Reset the ramp
func (r *Ramp) Reset() {
	r.acc = 0
	r.index = 0
	r.done = false
}

// Generate generates a ramp value, return false if end of ramp is reached
func (r *Ramp) Generate() (v float64) {
	v = math.Pow(r.acc, r.exp)*r.dev + r.min

	if r.done {
		return
	}

	r.index++
	r.done = r.index >= r.steps

	if !r.done {
		r.acc += r.inc
	}

	return
}

// Continuous returns continuous flag
func (r *Ramp) Continuous() bool {
	return r.continuous
}

// Done returns true if ramp is done
func (r *Ramp) Done() bool {
	return !r.continuous && r.done
}
