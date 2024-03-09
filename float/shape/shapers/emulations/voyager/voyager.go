package voyager

import (
	"github.com/almerlucke/genny/float/shape/shapers/function"
	"github.com/almerlucke/genny/float/shape/shapers/linear"
	"github.com/almerlucke/genny/float/shape/shapers/series"
)

// Sawtooth Minimoog Voyager emulation
type Sawtooth struct {
	series *series.Series
}

func NewSawtooth() *Sawtooth {
	return &Sawtooth{
		series: series.New(
			linear.New(0.25, 0.0),
			function.NewSin(),
			linear.NewBipolar(),
		),
	}
}

func (saw *Sawtooth) Shape(x float64) float64 {
	return saw.series.Shape(x)
}
