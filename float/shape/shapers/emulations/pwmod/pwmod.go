package pwmod

import (
	"github.com/almerlucke/genny/float/shape/shapers/function"
	"github.com/almerlucke/genny/float/shape/shapers/linear"
	"github.com/almerlucke/genny/float/shape/shapers/mult"
	"github.com/almerlucke/genny/float/shape/shapers/pulse"
	"github.com/almerlucke/genny/float/shape/shapers/series"
)

type PWMod struct {
	series *series.Series
}

func New() *PWMod {
	return &PWMod{
		series: series.New(
			mult.New(1.25),
			function.NewMod1(),
			pulse.New(0.4),
			linear.NewBipolar(),
		),
	}
}

func (pwm *PWMod) Shape(x float64) float64 {
	return pwm.series.Shape(x)
}

func (pwm *PWMod) SetA1(a1 float64) {
	pwm.series.Shapers[0].(*mult.Mult).M = a1
}

func (pwm *PWMod) SetWidth(w float64) {
	pwm.series.Shapers[2].(*pulse.Pulse).W = w
}
