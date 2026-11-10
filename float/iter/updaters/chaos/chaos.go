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

// rk4 advances a three dimensional system dv/dt = f(v) one step of dt with Runge-Kutta
func rk4(v []float64, dt float64, f func(x, y, z float64) (float64, float64, float64)) {
	k1x, k1y, k1z := f(v[0], v[1], v[2])
	k2x, k2y, k2z := f(v[0]+dt/2*k1x, v[1]+dt/2*k1y, v[2]+dt/2*k1z)
	k3x, k3y, k3z := f(v[0]+dt/2*k2x, v[1]+dt/2*k2y, v[2]+dt/2*k2z)
	k4x, k4y, k4z := f(v[0]+dt*k3x, v[1]+dt*k3y, v[2]+dt*k3z)

	v[0] += dt / 6 * (k1x + 2*k2x + 2*k3x + k4x)
	v[1] += dt / 6 * (k1y + 2*k2y + 2*k3y + k4y)
	v[2] += dt / 6 * (k1z + 2*k2z + 2*k3z + k4z)
}

// Lorenz is the Lorenz attractor, a continuous 3D system. With the classic values (sigma 10, rho 28,
// beta 8/3) x and y stay roughly within [-20, 20] and z within [0, 50]. dt sets the speed
type Lorenz struct {
	Sigma, Rho, Beta float64
	Dt               float64
}

// NewLorenz creates a Lorenz updater, start it for instance from (1, 1, 1)
func NewLorenz(sigma, rho, beta, dt float64) *Lorenz {
	return &Lorenz{Sigma: sigma, Rho: rho, Beta: beta, Dt: dt}
}

func (l *Lorenz) Update(v []float64) {
	rk4(v, l.Dt, func(x, y, z float64) (float64, float64, float64) {
		return l.Sigma * (y - x), x*(l.Rho-z) - y, x*y - l.Beta*z
	})
}

// Rossler is the Rössler attractor, a continuous 3D system that is smoother than Lorenz. With the
// classic values (a 0.2, b 0.2, c 5.7) x and y stay roughly within [-12, 12] and z within [0, 25]
type Rossler struct {
	A, B, C float64
	Dt      float64
}

// NewRossler creates a Rössler updater, start it for instance from (1, 1, 1)
func NewRossler(a, b, c, dt float64) *Rossler {
	return &Rossler{A: a, B: b, C: c, Dt: dt}
}

func (r *Rossler) Update(v []float64) {
	rk4(v, r.Dt, func(x, y, z float64) (float64, float64, float64) {
		return -y - z, x + r.A*y, r.B + z*(x-r.C)
	})
}
