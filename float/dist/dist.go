// Package dist has random generators with other distributions than the uniform float/rand: normal
// (bell curve) and exponential values, for natural sounding velocities, durations and timing
package dist

import (
	"math"

	"github.com/almerlucke/genny"
)

type bounds struct {
	lo, hi float64
	on     bool
}

func (b bounds) apply(v float64) float64 {
	if b.on {
		return min(max(v, b.lo), b.hi)
	}
	return v
}

// Gaussian generates normally distributed values
type Gaussian struct {
	mean, stddev float64
	clamp        bounds
}

// NewGaussian creates a generator of normally distributed values around mean. It panics if stddev is negative
func NewGaussian(mean, stddev float64) *Gaussian {
	if stddev < 0 {
		panic("dist: stddev must not be negative")
	}
	return &Gaussian{mean: mean, stddev: stddev}
}

// Clamp limits the output to [lo, hi], for instance to keep velocities valid
func (g *Gaussian) Clamp(lo, hi float64) *Gaussian {
	g.clamp = bounds{lo, hi, true}
	return g
}

func (g *Gaussian) Generate() float64 {
	// Box-Muller, 1-u keeps the logarithm finite
	u1, u2 := 1-genny.Float64(), genny.Float64()
	z := math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)
	return g.clamp.apply(g.mean + g.stddev*z)
}

func (g *Gaussian) Continuous() bool { return true }

func (g *Gaussian) Done() bool { return false }

func (g *Gaussian) Reset() {}

// Exponential generates exponentially distributed values: many small ones and a few large ones, like
// the time between random events
type Exponential struct {
	mean  float64
	clamp bounds
}

// NewExponential creates a generator of values >= 0 with the given mean. It panics if mean is not positive
func NewExponential(mean float64) *Exponential {
	if !(mean > 0) {
		panic("dist: mean must be positive")
	}
	return &Exponential{mean: mean}
}

// Clamp limits the output to [lo, hi]
func (e *Exponential) Clamp(lo, hi float64) *Exponential {
	e.clamp = bounds{lo, hi, true}
	return e
}

func (e *Exponential) Generate() float64 {
	return e.clamp.apply(-e.mean * math.Log(1-genny.Float64()))
}

func (e *Exponential) Continuous() bool { return true }

func (e *Exponential) Done() bool { return false }

func (e *Exponential) Reset() {}
