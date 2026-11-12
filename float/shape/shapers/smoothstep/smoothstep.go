// Package smoothstep provides the smoothstep family: S shaped transitions from 0 to 1 between two edges
// with a zero slope at both ends.
//
// With t = clamp((x - edge0) / (edge1 - edge0), 0, 1) the shapes are
//
//	Smoothstep:    3t^2 - 2t^3                (continuous first derivative)
//	Smootherstep:  6t^5 - 15t^4 + 10t^3       (Perlin, continuous first and second derivative)
//	Inverse:       1/2 - sin(asin(1-2t) / 3)  (the inverse function of Smoothstep)
//
// Inputs outside the edges are clamped, so the output is always inside [0, 1] and monotone.
package smoothstep

import "math"

// Kind selects the curve
type Kind int

const (
	// Smoothstep is 3t^2 - 2t^3
	Smoothstep Kind = iota
	// Smootherstep is Perlin's 6t^5 - 15t^4 + 10t^3
	Smootherstep
	// Inverse is the inverse of Smoothstep, a curve with a steep middle and flat ends turned on its side
	Inverse
)

// Step is a smoothstep shaper
type Step struct {
	kind         Kind
	edge0, width float64
}

// New creates a Step between edge0 and edge1. It panics if the edges are not finite, edge0 is not
// smaller than edge1, or the kind is unknown.
func New(kind Kind, edge0, edge1 float64) *Step {
	if kind < Smoothstep || kind > Inverse {
		panic("smoothstep: unknown kind")
	}
	if math.IsNaN(edge0) || math.IsInf(edge0, 0) || math.IsNaN(edge1) || math.IsInf(edge1, 0) || !(edge0 < edge1) {
		panic("smoothstep: edges must be finite with edge0 < edge1")
	}
	return &Step{kind: kind, edge0: edge0, width: edge1 - edge0}
}

// Shape maps x
func (s *Step) Shape(x float64) float64 {
	t := (x - s.edge0) / s.width
	if !(t > 0) { // also catches NaN
		return 0
	}
	if t > 1 {
		t = 1
	}
	switch s.kind {
	case Smootherstep:
		return t * t * t * (t*(6*t-15) + 10)
	case Inverse:
		return 0.5 - math.Sin(math.Asin(1-2*t)/3)
	}
	return t * t * (3 - 2*t)
}
