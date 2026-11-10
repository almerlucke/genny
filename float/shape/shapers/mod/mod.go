// Package mod provides Mod, a shaper that returns the floating point remainder of its input divided by
// M.
package mod

import "math"

type Mod struct {
	M float64
}

func New(m float64) *Mod {
	return &Mod{M: m}
}

func (m *Mod) Shape(x float64) float64 {
	return math.Mod(x, m.M)
}
