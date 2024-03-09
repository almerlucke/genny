package route

import "github.com/almerlucke/genny/float/shape"

type Route struct {
	Shapers []shape.Shaper
	Index   int
}

func New(index int, shapers ...shape.Shaper) *Route {
	return &Route{
		Shapers: shapers,
		Index:   index,
	}
}

func (r *Route) Shape(x float64) float64 {
	return r.Shapers[r.Index].Shape(x)
}

func (r *Route) Selected() shape.Shaper {
	return r.Shapers[r.Index]
}
