package hardsync

import (
	"github.com/almerlucke/genny/float/shape/shapers/function"
	"github.com/almerlucke/genny/float/shape/shapers/linear"
	"github.com/almerlucke/genny/float/shape/shapers/series"
)

type HardSync struct {
	series *series.Series
}

func New(a1 float64) *HardSync {
	return &HardSync{
		series: series.New(
			linear.New(a1, 0.0),
			function.NewMod1(),
			linear.NewBipolar(),
		),
	}
}

func (hs *HardSync) Shape(x float64) float64 {
	return hs.series.Shape(x)
}

func (hs *HardSync) SetA1(a1 float64) {
	hs.series.Shapers[0].(*linear.Linear).Scale = a1
}
