package interpolate

import "github.com/almerlucke/genny/float/shape"

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
	fx := ip.Index * float64(len(ip.Shapers)-1)
	ix1 := int(fx)
	ix2 := ix1 + 1

	if ix2 >= len(ip.Shapers) {
		ix2 = ix1
	}

	sx1 := ip.Shapers[ix1].Shape(x)
	sx2 := ip.Shapers[ix2].Shape(x)

	return sx1 + (fx-float64(ix1))*(sx2-sx1)
}
