package series

import "github.com/almerlucke/genny/float/shape"

type Series struct {
	Shapers []shape.Shaper
}

func New(shapers ...shape.Shaper) *Series {
	return &Series{
		Shapers: shapers,
	}
}

func (s *Series) Shape(x float64) float64 {
	for _, shaper := range s.Shapers {
		x = shaper.Shape(x)
	}

	return x
}
