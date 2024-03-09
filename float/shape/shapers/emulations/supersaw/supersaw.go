package supersaw

import (
	"github.com/almerlucke/genny/float/shape/shapers/function"
	"github.com/almerlucke/genny/float/shape/shapers/linear"
	"github.com/almerlucke/genny/float/shape/shapers/mod"
	"github.com/almerlucke/genny/float/shape/shapers/mult"
	"github.com/almerlucke/genny/float/shape/shapers/parallel"
	"github.com/almerlucke/genny/float/shape/shapers/series"
	"math"
)

type SuperSaw struct {
	series *series.Series
}

func New(a1 float64, m1 float64, m2 float64) *SuperSaw {
	return &SuperSaw{
		series: series.New(
			mult.New(a1),                           // a1=1.5
			parallel.New(mod.New(m1), mod.New(m2)), // m1=025, m2=0.88
			function.New(math.Sin),
			linear.NewBipolar(),
		),
	}
}

func (saw *SuperSaw) Shape(x float64) float64 {
	return saw.series.Shape(x)
}

func (saw *SuperSaw) SetA1(a1 float64) {
	saw.series.Shapers[0].(*mult.Mult).M = a1
}

func (saw *SuperSaw) SetM1(m1 float64) {
	saw.series.Shapers[1].(*parallel.Parallel).Shapers[0].(*mod.Mod).M = m1
}

func (saw *SuperSaw) SetM2(m2 float64) {
	saw.series.Shapers[1].(*parallel.Parallel).Shapers[1].(*mod.Mod).M = m2
}
