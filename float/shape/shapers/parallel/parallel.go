package parallel

import "github.com/almerlucke/genny/float/shape"

type Function func(float64, float64) float64

type Parallel struct {
	Shapers  []shape.Shaper
	Function Function
	Start    float64
}

func New(shapers ...shape.Shaper) *Parallel {
	return NewWithFunction(0, func(x float64, v float64) float64 { return x + v }, shapers...)
}

func NewWithFunction(start float64, function Function, shapers ...shape.Shaper) *Parallel {
	return &Parallel{
		Start:    start,
		Shapers:  shapers,
		Function: function,
	}
}

func (p *Parallel) Shape(x float64) float64 {
	v := p.Start

	for _, shaper := range p.Shapers {
		v = p.Function(shaper.Shape(x), v)
	}

	return v
}
