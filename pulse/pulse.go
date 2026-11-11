// Package pulse works with metric indispensability, in the spirit of Clarence Barlow: how important each pulse
// of a meter is. The first pulse of a bar matters most, then the start of the next level of the
// division, and so on down to the offbeats. Thinning a rhythm by importance keeps the pulses that
// carry the meter and drops the weak ones first, so a rhythm can fade in and out and still sound
// like the same meter.
package pulse

import (
	"math"
	"slices"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/constant"
)

// baseOrder is the importance order of the positions 1..p-1 inside a division by the prime p
// (as in Barlow's tables), most important first. Larger primes put the later pulses first
func baseOrder(p int) []int {
	switch p {
	case 2:
		return []int{1}
	case 3:
		return []int{2, 1}
	case 5:
		return []int{2, 4, 3, 1}
	case 7:
		return []int{4, 2, 6, 3, 5, 1}
	}
	order := make([]int, 0, p-1)
	for j := p - 1; j >= 1; j-- {
		order = append(order, j)
	}
	return order
}

// primes splits n into prime factors, smallest first
func primes(n int) []int {
	var out []int
	for p := 2; p*p <= n; p++ {
		for n%p == 0 {
			out = append(out, p)
			n /= p
		}
	}
	if n > 1 {
		out = append(out, n)
	}
	return out
}

// Hierarchy returns the indispensability of every pulse of a meter made of nested divisions: 3, 2, 2
// is three groups of two groups of two pulses, 12 pulses. The importances are the numbers 0 to
// n-1, the first pulse has the highest (n-1). A division that is not a prime is split into its
// prime factors (4 is 2, 2). Of equal pulses in different units the one in the unit with the weaker
// start is the more important (in 2, 2 the last pulse beats the second). No divisions give a single pulse. It panics if a division is less than 1
func Hierarchy(divisions ...int) []int {
	var factors []int
	for _, d := range divisions {
		if d < 1 {
			panic("pulse: a division must be at least 1")
		}
		factors = append(factors, primes(d)...)
	}
	n := 1
	for _, f := range factors {
		n *= f
	}

	ind := make([]int, n)
	next := n - 1
	ind[0] = next // the pulse that starts the bar
	next--

	unit := n // length of the units divided at the current level
	for _, p := range factors {
		sub := unit / p // distance between the new pulses of this level
		// the new pulses are the positions j*sub (j in 1..p-1) of every unit of length unit
		for _, j := range baseOrder(p) {
			// units with a less important start go first
			starts := make([]int, 0, n/unit)
			for s := 0; s < n; s += unit {
				starts = append(starts, s)
			}
			slices.SortStableFunc(starts, func(a, b int) int { return ind[a] - ind[b] })
			for _, s := range starts {
				ind[s+j*sub] = next
				next--
			}
		}
		unit = sub
	}
	return ind
}

// Thin returns which pulses play at a thinning level: the pulses with an importance of at least level.
// Level 0 (or less) keeps every pulse, a higher level is sparser, level len(ind) or more keeps none
func Thin(ind []int, level int) []bool {
	out := make([]bool, len(ind))
	for i, v := range ind {
		out[i] = v >= level
	}
	return out
}

// Velocity maps the importance of every pulse to a velocity from lo (the weakest pulse) to hi (the
// first pulse). A single pulse gets hi
func Velocity(ind []int, lo, hi float64) []float64 {
	out := make([]float64, len(ind))
	top := 0
	for _, v := range ind {
		top = max(top, v)
	}
	for i, v := range ind {
		if top == 0 {
			out[i] = hi
		} else {
			out[i] = lo + (hi-lo)*float64(v)/float64(top)
		}
	}
	return out
}

// Interpolate blends two hierarchies of the same meter: rate 0 is a, 1 is b and in between the
// importances are mixed (rounded), to morph from one feel to another. The rate is clamped to [0, 1].
// It panics if the lengths differ
func Interpolate(a, b []int, rate float64) []int {
	if len(a) != len(b) {
		panic("pulse: hierarchies must have the same length")
	}
	rate = min(max(rate, 0), 1)
	out := make([]int, len(a))
	for i := range a {
		out[i] = int(math.Round(float64(a[i]) + float64(b[i]-a[i])*rate))
	}
	return out
}

// Gate is a looping gate over the pulses of a meter: a pulse plays when its importance is at least the
// thinning level of that moment
type Gate struct {
	ind   []int
	level genny.Generator[float64]
	step  int
}

// NewGate creates a gate with a thinning level that is generated for every pulse and rounded, so it
// can follow an envelope or an LFO: a higher level is sparser, a lower level denser. The level is
// clamped to the range of the meter. With an empty hierarchy the gate is done at once and false
func NewGate(ind []int, level genny.Generator[float64]) *Gate {
	return &Gate{ind: append([]int{}, ind...), level: level}
}

// NewConstGate creates a gate with a fixed thinning level, see NewGate
func NewConstGate(ind []int, level int) *Gate {
	return NewGate(ind, constant.New(float64(level)))
}

func (g *Gate) Generate() bool {
	if len(g.ind) == 0 {
		return false
	}
	level := g.level.Generate()
	v := !(float64(g.ind[g.step]) < math.Round(level))
	if math.IsNaN(level) {
		v = true
	}
	g.step = (g.step + 1) % len(g.ind)
	return v
}

func (g *Gate) Continuous() bool { return len(g.ind) > 0 }

func (g *Gate) Done() bool { return len(g.ind) == 0 }

func (g *Gate) Reset() {
	g.step = 0
	g.level.Reset()
}
