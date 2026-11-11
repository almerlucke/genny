package chaos

import (
	"math"

	"github.com/almerlucke/genny/float/iter"
)

// Axes selects which values of a two dimensional state are taken: x, y or both
type Axes int

const (
	// AxisX takes the first value (a frame of one value)
	AxisX Axes = iota
	// AxisY takes the second value (a frame of one value)
	AxisY
	// AxisBoth takes both values (a frame of two values)
	AxisBoth
)

// Projection is a float.FrameGenerator that takes the x, the y or both values of an Iterator
type Projection struct {
	it   *iter.Iterator
	axes Axes
	out  []float64
}

// Project takes the values selected by axes from the frames of an iterator with at least two
// values. An iterator with fewer values gives zeros for the missing ones
func Project(it *iter.Iterator, axes Axes) *Projection {
	n := 1
	if axes == AxisBoth {
		n = 2
	}
	return &Projection{it: it, axes: axes, out: make([]float64, n)}
}

func (p *Projection) Generate() []float64 {
	f := p.it.Generate()
	at := func(i int) float64 {
		if i < len(f) {
			return f[i]
		}
		return 0
	}
	switch p.axes {
	case AxisY:
		p.out[0] = at(1)
	case AxisBoth:
		p.out[0], p.out[1] = at(0), at(1)
	default:
		p.out[0] = at(0)
	}
	return p.out
}

func (p *Projection) Dimensions() int  { return len(p.out) }
func (p *Projection) Continuous() bool { return true }
func (p *Projection) Done() bool       { return false }
func (p *Projection) Reset()           { p.it.Reset() }

// rk4Driven advances a driven two dimensional system (x' = f(x, v, t) in the form dx = v, dv = g)
// one step of dt, the time t is advanced with it
func rk4Driven(v []float64, t *float64, dt float64, g func(x, y, t float64) (float64, float64)) {
	s := [3]float64{v[0], v[1], *t}
	rk4(s[:], dt, func(x, y, t float64) (float64, float64, float64) {
		dx, dy := g(x, y, t)
		return dx, dy, 1
	})
	v[0], v[1], *t = s[0], s[1], s[2]
}

// Duffing is the driven Duffing oscillator x” + Delta x' + Alpha x + Beta x^3 = Gamma cos(Omega t).
// The state is (x, velocity). With the defaults of NewDuffing (a double well) x stays roughly within
// [-1.6, 1.6] and the velocity within [-1.1, 1.1], and the motion is chaotic
type Duffing struct {
	Alpha, Beta, Delta, Gamma, Omega float64
	Dt                               float64
	t                                float64
}

// NewDuffing creates a Duffing updater with the classic chaotic values (alpha -1, beta 1, delta 0.3,
// gamma 0.5, omega 1.2, dt 0.05). Start it for instance from (1, 0)
func NewDuffing() *Duffing {
	return &Duffing{Alpha: -1, Beta: 1, Delta: 0.3, Gamma: 0.5, Omega: 1.2, Dt: 0.05}
}

func (d *Duffing) Update(v []float64) {
	if len(v) < 2 {
		return
	}
	rk4Driven(v, &d.t, d.Dt, func(x, y, t float64) (float64, float64) {
		return y, -d.Delta*y - d.Alpha*x - d.Beta*x*x*x + d.Gamma*math.Cos(d.Omega*t)
	})
	if p := 2 * math.Pi / d.Omega; d.Omega > 0 && d.t > p {
		d.t = math.Mod(d.t, p)
	}
}

// Pendulum is the damped driven pendulum theta” + Damping theta' + sin(theta) = Drive cos(Omega t).
// The state is (angle, angular velocity). With the defaults of NewPendulum it is chaotic, the angle is
// wrapped to [-pi, pi) (set Wrap to false to let it wind up) and the velocity stays within about [-3, 3]
type Pendulum struct {
	Damping, Drive, Omega float64
	Dt                    float64
	Wrap                  bool
	t                     float64
}

// NewPendulum creates a pendulum updater with chaotic values (damping 0.5, drive 1.2, omega 2/3, dt
// 0.05, wrapping the angle). Start it for instance from (0.2, 0)
func NewPendulum() *Pendulum {
	return &Pendulum{Damping: 0.5, Drive: 1.2, Omega: 2.0 / 3, Dt: 0.05, Wrap: true}
}

func (p *Pendulum) Update(v []float64) {
	if len(v) < 2 {
		return
	}
	rk4Driven(v, &p.t, p.Dt, func(x, y, t float64) (float64, float64) {
		return y, -p.Damping*y - math.Sin(x) + p.Drive*math.Cos(p.Omega*t)
	})
	if p.Wrap {
		v[0] = math.Mod(v[0]+math.Pi, 2*math.Pi)
		if v[0] < 0 {
			v[0] += 2 * math.Pi
		}
		v[0] -= math.Pi
	}
	if q := 2 * math.Pi / p.Omega; p.Omega > 0 && p.t > q {
		p.t = math.Mod(p.t, q)
	}
}

// StandardMap is the Chirikov standard map: p' = p + K sin(theta), theta' = theta + p', both mod 2 pi.
// The state is (theta, p), both in [0, 2 pi). Above K of about 0.97 the orbits turn chaotic
type StandardMap struct {
	K float64
}

// NewStandardMap creates a standard map updater, K 1 is mildly chaotic and 2 or more strongly so.
// Start it for instance from (1, 0.5)
func NewStandardMap(k float64) *StandardMap { return &StandardMap{K: k} }

func (s *StandardMap) Update(v []float64) {
	if len(v) < 2 {
		return
	}
	const tau = 2 * math.Pi
	p := math.Mod(v[1]+s.K*math.Sin(v[0]), tau)
	if p < 0 {
		p += tau
	}
	th := math.Mod(v[0]+p, tau)
	if th < 0 {
		th += tau
	}
	v[0], v[1] = th, p
}
