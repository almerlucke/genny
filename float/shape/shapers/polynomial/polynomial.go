// Package polynomial provides Polynomial, a shaper y = c0 + c1 x + c2 x^2 + ... evaluated with
// Horner's rule, the Legendre polynomials, and Normalized, which scales a shaper so its peak over
// [-1,1] is 1.
//
// Horner's rule evaluates y = (((cn x + c(n-1)) x + ...) x + c0) with n multiplications. The
// polynomials are meant for inputs in [-1,1]: outside that range a polynomial of high degree grows
// fast. Shape never returns NaN: a result that overflows is clamped to +-MaxFloat64 and a NaN input
// gives 0.
//
// The Legendre polynomials P_n are orthogonal on [-1,1] and follow the recurrence
// (n+1) P(n+1) = (2n+1) x P(n) - n P(n-1), with P0 = 1 and P1 = x. P_n has n zeros in (-1,1), is even
// for even n and odd for odd n, and |P_n| <= 1 on [-1,1] with P_n(1) = 1.
package polynomial

import (
	"fmt"
	"math"

	"github.com/almerlucke/genny/float/shape"
)

// Polynomial is a polynomial shaper
type Polynomial struct {
	c []float64
}

// New creates a polynomial from coefficients in ascending order: New(1, 0, -2) is 1 - 2x^2. It
// panics if there are no coefficients or one is not finite
func New(coefficients ...float64) *Polynomial {
	if len(coefficients) == 0 {
		panic("polynomial: New needs at least one coefficient")
	}
	for _, c := range coefficients {
		if math.IsNaN(c) || math.IsInf(c, 0) {
			panic(fmt.Sprintf("polynomial: coefficients must be finite, got %v", c))
		}
	}
	return &Polynomial{c: append([]float64{}, coefficients...)}
}

// maxLegendre is the highest order of NewLegendre. The coefficients grow with the order and
// alternate in sign, so beyond this order Horner's rule loses too much precision
const maxLegendre = 20

// NewLegendre creates the Legendre polynomial of order n, 0 <= n <= 20. It panics otherwise
func NewLegendre(n int) *Polynomial {
	if n < 0 || n > maxLegendre {
		panic(fmt.Sprintf("polynomial: NewLegendre: order must be in 0..%d, got %d", maxLegendre, n))
	}
	prev, cur := []float64{1}, []float64{1}
	if n == 0 {
		return &Polynomial{c: cur}
	}
	cur = []float64{0, 1}
	for k := 1; k < n; k++ {
		next := make([]float64, k+2)
		for i, c := range cur { // (2k+1) x P_k
			next[i+1] += float64(2*k+1) * c
		}
		for i, c := range prev { // - k P_(k-1)
			next[i] -= float64(k) * c
		}
		for i := range next {
			next[i] /= float64(k + 1)
		}
		prev, cur = cur, next
	}
	return &Polynomial{c: cur}
}

// Degree is the degree of the polynomial (the number of coefficients minus one)
func (p *Polynomial) Degree() int { return len(p.c) - 1 }

// Coefficients returns a copy of the coefficients in ascending order
func (p *Polynomial) Coefficients() []float64 { return append([]float64{}, p.c...) }

// Shape evaluates the polynomial at x
func (p *Polynomial) Shape(x float64) float64 {
	if math.IsNaN(x) {
		return 0
	}
	y := 0.0
	for i := len(p.c) - 1; i >= 0; i-- {
		y = y*x + p.c[i]
	}
	switch {
	case math.IsNaN(y):
		return 0
	case y > math.MaxFloat64:
		return math.MaxFloat64
	case y < -math.MaxFloat64:
		return -math.MaxFloat64
	}
	return y
}

// Normalized scales a shaper so that its largest absolute value over [-1,1] is 1
type Normalized struct {
	inner shape.Shaper
	gain  float64
}

// defaultSamples is the number of points NewNormalized evaluates
const defaultSamples = 2001

// NewNormalized wraps s. The peak is found by sampling s at 2001 points spread evenly over
// [-1,1] (the end points included), so for a smooth shaper the gain is exact to about the sampling
// resolution (the peak of a polynomial of low degree is found to 1e-6 or better), but a narrow spike
// between sample points is missed. A shaper that is 0 on all samples (or not finite) gets gain 1.
// s is evaluated at construction, so a stateful shaper is advanced; prefer pure shapers
func NewNormalized(s shape.Shaper) *Normalized { return NewNormalizedN(s, defaultSamples) }

// NewNormalizedN is NewNormalized with n sample points, n >= 2. It panics if s is nil or n < 2
func NewNormalizedN(s shape.Shaper, n int) *Normalized {
	if s == nil || n < 2 {
		panic("polynomial: NewNormalizedN needs a shaper and at least 2 sample points")
	}
	peak := 0.0
	for i := 0; i < n; i++ {
		x := -1 + 2*float64(i)/float64(n-1)
		if a := math.Abs(s.Shape(x)); a > peak && !math.IsInf(a, 0) {
			peak = a
		}
	}
	gain := 1.0
	if peak > 0 {
		gain = 1 / peak
	}
	return &Normalized{inner: s, gain: gain}
}

// Gain is the factor applied to the wrapped shaper
func (n *Normalized) Gain() float64 { return n.gain }

// Shape returns the scaled value of the wrapped shaper
func (n *Normalized) Shape(x float64) float64 { return n.gain * n.inner.Shape(x) }
