package ripple

import "math"

type Ripple struct {
	M float64
}

func New(m float64) *Ripple {
	return &Ripple{M: m}
}

func (r *Ripple) Shape(x float64) float64 {
	return x + math.Mod(x, r.M)
}
