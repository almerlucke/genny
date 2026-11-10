// Package interp provides Interpolate, a shaper that crossfades between neighbouring shapers with an
// Index between 0 and 1.
package interp

import (
	"math"

	"github.com/almerlucke/genny/float/shape"
)

type Interpolate struct {
	Shapers []shape.Shaper
	Index   float64
}

func New(shapers ...shape.Shaper) *Interpolate {
	return &Interpolate{
		Shapers: shapers,
	}
}

func (ip *Interpolate) Shape(x float64) float64 {
	index := math.Min(math.Max(ip.Index, 0), 1)
	fx := index * float64(len(ip.Shapers)-1)
	ix1 := int(fx)
	ix2 := ix1 + 1

	if ix2 >= len(ip.Shapers) {
		ix2 = ix1
	}

	sx1 := ip.Shapers[ix1].Shape(x)
	sx2 := ip.Shapers[ix2].Shape(x)

	return sx1 + (fx-float64(ix1))*(sx2-sx1)
}
