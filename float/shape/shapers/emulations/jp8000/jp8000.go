package jp8000

import (
	"github.com/almerlucke/genny/float/shape/shapers/function"
	"github.com/almerlucke/genny/float/shape/shapers/linear"
	"github.com/almerlucke/genny/float/shape/shapers/mult"
	"github.com/almerlucke/genny/float/shape/shapers/series"
	"math"
)

type TriMod struct {
	series *series.Series
}

func NewTriMod(m float64) *TriMod {
	return &TriMod{
		series: series.New(
			linear.NewBipolar(),
			function.NewAbs(),
			linear.NewBipolar(),
			function.NewMod1(),
			mult.New(m), // m=0.3 [0.7 - 1.0]
			function.New(func(x float64) float64 { return 2.0 * (x - math.Ceil(x-0.5)) }),
		),
	}
}

func (tri *TriMod) Shape(x float64) float64 {
	return tri.series.Shape(x)
}

func (tri *TriMod) SetMod(m float64) {
	tri.series.Shapers[4].(*mult.Mult).M = m
}
