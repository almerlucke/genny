package function

import "math"

type Function struct {
	f func(float64) float64
}

func New(f func(float64) float64) *Function {
	return &Function{f: f}
}

func NewMod(m float64) *Function {
	return New(func(x float64) float64 { return math.Mod(x, m) })
}

func NewMod1() *Function {
	return NewMod(1.0)
}

func NewAbs() *Function {
	return New(math.Abs)
}

func NewSin() *Function {
	return New(func(x float64) float64 { return math.Sin(x * 2.0 * math.Pi) })
}

func NewTri() *Function {
	return New(func(x float64) float64 {
		if x < 0.5 {
			return 2.0 * x
		} else {
			return 2.0 - 2.0*x
		}
	})
}

func (f *Function) Shape(x float64) float64 {
	return f.f(x)
}
