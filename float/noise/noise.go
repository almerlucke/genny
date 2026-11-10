// Package noise has smooth (gradient, Perlin style) noise: random modulation without the jumps of
// white noise or the drift of a random walk
package noise

import "github.com/almerlucke/genny"

const tableSize = 256

// Noise is one dimensional gradient noise in [-1, 1] (or [0, 1] if unipolar), optionally with
// octaves for more detail. The gradients are drawn from the genny random stream when the
// noise is created, so genny.Seed makes it repeatable
type Noise struct {
	grad        [tableSize]float64
	pos         float64
	start       float64
	step        float64
	octaves     int
	persistence float64
	lacunarity  float64
	unipolar    bool
}

// New creates noise that moves step along the noise curve for every Generate: a small step is slow
// and smooth (1 is about one wiggle per step). Every octave adds noise at lacunarity (2) times the
// frequency and persistence (0.5) times the amplitude
func New(step float64, octaves int) *Noise {
	n := &Noise{step: step, octaves: max(octaves, 1), persistence: 0.5, lacunarity: 2}
	for i := range n.grad {
		n.grad[i] = genny.Float64()*2 - 1
	}
	return n
}

// SetStep changes the speed
func (n *Noise) SetStep(step float64) { n.step = step }

// SetDetail changes the persistence and lacunarity of the octaves
func (n *Noise) SetDetail(persistence, lacunarity float64) {
	n.persistence, n.lacunarity = persistence, lacunarity
}

// SetStart sets the position on the noise curve that Reset returns to
func (n *Noise) SetStart(pos float64) {
	n.start, n.pos = pos, pos
}

// SetUnipolar makes the output range [0, 1] instead of [-1, 1]
func (n *Noise) SetUnipolar(u bool) *Noise {
	n.unipolar = u
	return n
}

func fade(t float64) float64 { return t * t * t * (t*(t*6-15) + 10) }

func (n *Noise) at(x float64) float64 {
	fl := int(x)
	if float64(fl) > x {
		fl--
	}
	f := x - float64(fl)
	g0 := n.grad[fl&(tableSize-1)]
	g1 := n.grad[(fl+1)&(tableSize-1)]
	a, b := g0*f, g1*(f-1)
	return (a + (b-a)*fade(f)) * 2
}

func (n *Noise) Generate() float64 {
	sum, amp, total, freq := 0.0, 1.0, 0.0, 1.0
	for o := range n.octaves {
		// an offset per octave keeps the octaves from lining up
		sum += amp * n.at(n.pos*freq+float64(o)*31.7)
		total += amp
		amp *= n.persistence
		freq *= n.lacunarity
	}
	n.pos += n.step

	v := min(max(sum/total, -1), 1)
	if n.unipolar {
		v = (v + 1) / 2
	}
	return v
}

func (n *Noise) Continuous() bool { return true }

func (n *Noise) Done() bool { return false }

func (n *Noise) Reset() { n.pos = n.start }
