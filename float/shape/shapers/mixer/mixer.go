package mixer

import "github.com/almerlucke/genny/float/shape"

type Mixer struct {
	Shapers []shape.Shaper
	Levels  []float64
}

func New(shapers ...shape.Shaper) *Mixer {
	levels := make([]float64, len(shapers))
	for i := 0; i < len(shapers); i++ {
		levels[i] = 1.0
	}

	return NewWithLevels(shapers, levels)
}

func NewWithLevels(shapers []shape.Shaper, levels []float64) *Mixer {
	return &Mixer{
		Shapers: shapers,
		Levels:  levels,
	}
}

func (mix *Mixer) Shape(x float64) float64 {
	accum := 0.0

	for index, shaper := range mix.Shapers {
		accum += shaper.Shape(x) * mix.Levels[index]
	}

	return accum
}
