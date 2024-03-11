package chaos

import (
	"math"
)

type Verhulst struct {
	A float64
	F func(float64) float64
}

func (verhulst *Verhulst) Update(v []float64) {
	v[0] = verhulst.A * verhulst.F(v[0])
}

func NewVerhulst(a float64) *Verhulst {
	return &Verhulst{A: a, F: Iter1}
}

func NewVerhulstWithFunc(a float64, f func(float64) float64) *Verhulst {
	return &Verhulst{A: a, F: f}
}

func Iter1(x float64) float64 {
	return x * (1.0 - x)
}

func Iter1a(x float64) float64 {
	return x * (1.0 - x*x)
}

func Iter1b(x float64) float64 {
	return x * math.Sin(x)
}

func Iter1c(x float64) float64 {
	return x * math.Exp(x)
}

func Iter1d(x float64) float64 {
	t := 1.0 + x*x
	return x / (t * t)
}

func Iter1e(x float64) float64 {
	return x * math.Sqrt(1.0-x)
}

type Henon struct {
	A float64
	B float64
	F func(float64) float64
}

func NewHenon(a float64, b float64) *Henon {
	return NewHenonWithFunc(a, b, func(x float64) float64 { return x * x })
}

func NewHenonWithFunc(a float64, b float64, f func(float64) float64) *Henon {
	return &Henon{
		A: a,
		B: b,
		F: f,
	}
}

func (h *Henon) Update(vec []float64) {
	x := vec[0]
	y := vec[1]

	vec[0] = y + 1.0 - h.A*h.F(x)
	vec[1] = h.B * x
}

type Aronson struct {
	A float64
	F func(float64) float64
}

func NewAronsonWithFunc(a float64, f func(float64) float64) *Aronson {
	return &Aronson{
		A: a,
		F: f,
	}
}

func NewAronson(a float64) *Aronson {
	return NewAronsonWithFunc(a, func(x float64) float64 { return x })
}

func (a *Aronson) Update(vec []float64) {
	x := vec[0]
	y := vec[1]

	vec[0] = y
	vec[1] = a.A * y * (1.0 - a.F(x))
}
