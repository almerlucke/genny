// Package pulse provides Pulse, a shaper that gives 1 below a width W and 0 otherwise.
package pulse

type Pulse struct {
	W float64
}

func New(w float64) *Pulse {
	return &Pulse{
		W: w,
	}
}

func (p *Pulse) Shape(x float64) float64 {
	if x < p.W {
		return 1.0
	}

	return 0.0
}
