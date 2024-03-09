package softsync

import (
	"github.com/almerlucke/genny/float/shape/shapers/function"
	"github.com/almerlucke/genny/float/shape/shapers/linear"
	"github.com/almerlucke/genny/float/shape/shapers/series"
)

type Triangle struct {
	series *series.Series
}

func NewTriangle(a1 float64) *Triangle {
	return &Triangle{
		series: series.New(
			linear.NewBipolar(),
			function.NewAbs(),
			linear.New(a1, 0.0), // a=1.25
			function.NewMod1(),
			function.NewTri(),
			linear.NewBipolar(),
		),
	}
}

func (t *Triangle) Shape(x float64) float64 {
	return t.series.Shape(x)
}

func (t *Triangle) SetA1(a1 float64) {
	t.series.Shapers[2].(*linear.Linear).Scale = a1
}
