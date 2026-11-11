package dist

import (
	"math"
	"math/bits"

	"github.com/almerlucke/genny"
)

// limits keeps values within [lo, hi]. With tries > 0 it first draws again (up to tries times) when a
// value is outside, so the shape of the distribution inside the range is kept, then it clamps
type limits struct {
	lo, hi float64
	tries  int
	on     bool
}

func (l limits) draw(f func() float64) float64 {
	v := f()
	if !l.on {
		return v
	}
	for i := 0; i < l.tries && (v < l.lo || v > l.hi); i++ {
		v = f()
	}
	return min(max(v, l.lo), l.hi)
}

func normal() float64 {
	u1, u2 := 1-genny.Float64(), genny.Float64()
	return math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)
}

// gamma draws from a gamma distribution with the given shape and scale 1 (Marsaglia and Tsang)
func gamma(shape float64) float64 {
	if shape < 1 {
		return gamma(shape+1) * math.Pow(1-genny.Float64(), 1/shape)
	}
	d := shape - 1.0/3
	c := 1 / math.Sqrt(9*d)
	for {
		x := normal()
		v := 1 + c*x
		if v <= 0 {
			continue
		}
		v = v * v * v
		u := 1 - genny.Float64()
		if math.Log(u) < 0.5*x*x+d-d*v+d*math.Log(v) {
			return d * v
		}
	}
}

// SampleBeta draws a value in [0, 1] from a beta distribution with shape parameters a and b, from
// the genny random stream. With a = b = 1 it is uniform, a < b favours low values, a > b high ones, and
// a, b < 1 favours the extremes. a and b must be positive, otherwise it returns 0.5
func SampleBeta(a, b float64) float64 {
	if !(a > 0) || !(b > 0) {
		return 0.5
	}
	x, y := gamma(a), gamma(b)
	if x+y == 0 {
		return 0.5
	}
	return x / (x + y)
}

// Beta generates beta distributed values in [0, 1], or in the range set by Scale
type Beta struct {
	a, b   float64
	lo, hi float64
}

// NewBeta creates a generator of beta distributed values in [0, 1]. It panics if a or b is not positive
func NewBeta(a, b float64) *Beta {
	if !(a > 0) || !(b > 0) {
		panic("dist: beta parameters must be positive")
	}
	return &Beta{a: a, b: b, hi: 1}
}

// Scale moves the output range from [0, 1] to [lo, hi]
func (g *Beta) Scale(lo, hi float64) *Beta {
	g.lo, g.hi = lo, hi
	return g
}

func (g *Beta) Generate() float64 { return g.lo + (g.hi-g.lo)*SampleBeta(g.a, g.b) }

func (g *Beta) Continuous() bool { return true }

func (g *Beta) Done() bool { return false }

func (g *Beta) Reset() {}

// Cauchy generates Cauchy distributed values: mostly near the location, but with now and then a very
// far outlier. Limit the output with Clamp or Retry
type Cauchy struct {
	location, scale float64
	lim             limits
}

// NewCauchy creates a generator around location. Half of the values lie within scale of it. It panics
// if scale is not positive
func NewCauchy(location, scale float64) *Cauchy {
	if !(scale > 0) {
		panic("dist: scale must be positive")
	}
	return &Cauchy{location: location, scale: scale}
}

// Clamp limits the output to [lo, hi]
func (g *Cauchy) Clamp(lo, hi float64) *Cauchy {
	g.lim = limits{lo: lo, hi: hi, on: true}
	return g
}

// Retry limits the output to [lo, hi] by drawing again when a value is outside, at most tries times,
// after that the value is clamped
func (g *Cauchy) Retry(lo, hi float64, tries int) *Cauchy {
	g.lim = limits{lo: lo, hi: hi, tries: tries, on: true}
	return g
}

func (g *Cauchy) Generate() float64 {
	return g.lim.draw(func() float64 {
		return g.location + g.scale*math.Tan(math.Pi*(genny.Float64()-0.5))
	})
}

func (g *Cauchy) Continuous() bool { return true }

func (g *Cauchy) Done() bool { return false }

func (g *Cauchy) Reset() {}

// Weibull generates Weibull distributed values, all at or above the threshold. A shape below 1 gives many
// values near the threshold and a long tail, 1 is exponential, and above 1 a hump that gets narrower
type Weibull struct {
	threshold, spread, shape float64
	lim                      limits
}

// NewWeibull creates a generator with the lowest value threshold, the spread (scale) of the values
// and a shape. It panics if spread or shape is not positive
func NewWeibull(threshold, spread, shape float64) *Weibull {
	if !(spread > 0) || !(shape > 0) {
		panic("dist: spread and shape must be positive")
	}
	return &Weibull{threshold: threshold, spread: spread, shape: shape}
}

// Clamp limits the output to [lo, hi]
func (g *Weibull) Clamp(lo, hi float64) *Weibull {
	g.lim = limits{lo: lo, hi: hi, on: true}
	return g
}

// Retry limits the output to [lo, hi] by drawing again when a value is outside, at most tries times,
// after that the value is clamped
func (g *Weibull) Retry(lo, hi float64, tries int) *Weibull {
	g.lim = limits{lo: lo, hi: hi, tries: tries, on: true}
	return g
}

func (g *Weibull) Generate() float64 {
	return g.lim.draw(func() float64 {
		return g.threshold + g.spread*math.Pow(-math.Log(1-genny.Float64()), 1/g.shape)
	})
}

func (g *Weibull) Continuous() bool { return true }

func (g *Weibull) Done() bool { return false }

func (g *Weibull) Reset() {}

// Lehmer is a linear congruential generator, X(n+1) = (a*X(n) + b) mod m, that generates X/m in [0, 1).
// It has its own state and does not use the genny random stream, so it is repeatable by its own seed:
// the same seed always gives the same sequence, whatever else is drawn
type Lehmer struct {
	seed, x, a, b, m uint64
}

// NewLehmer creates a generator with the constants of Numerical Recipes (a = 1664525, b = 1013904223,
// m = 2^32)
func NewLehmer(seed uint64) *Lehmer { return NewLehmerParams(seed, 1664525, 1013904223, 1<<32) }

// NewLehmerParams creates a generator with your own constants. It panics if m is 0 or larger than 2^62
func NewLehmerParams(seed, a, b, m uint64) *Lehmer {
	if m == 0 || m > 1<<62 {
		panic("dist: modulus must be between 1 and 2^62")
	}
	return &Lehmer{seed: seed % m, x: seed % m, a: a, b: b, m: m}
}

func (l *Lehmer) Generate() float64 {
	hi, lo := bits.Mul64(l.a, l.x)
	_, r := bits.Div64(hi%l.m, lo, l.m)
	l.x = (r + l.b%l.m) % l.m
	v := float64(l.x) / float64(l.m)
	if v >= 1 { // rounding of a huge x
		v = math.Nextafter(1, 0)
	}
	return v
}

func (l *Lehmer) Continuous() bool { return true }

func (l *Lehmer) Done() bool { return false }

// Reset goes back to the seed
func (l *Lehmer) Reset() { l.x = l.seed }

// Pink generates 1/f (pink) noise in [0, 1) with the Voss-McCartney algorithm: a number of random
// rows that are redrawn at halving rates are added up. It sounds (and plots) between white noise
// and a random walk, with slow drifts and fast detail. The values come from the genny random stream
type Pink struct {
	rows    []float64
	sum     float64
	counter uint64
}

// NewPink creates pink noise with the given number of rows (octaves of detail). It panics if rows
// is less than 1
func NewPink(rows int) *Pink {
	if rows < 1 {
		panic("dist: pink noise needs at least 1 row")
	}
	p := &Pink{rows: make([]float64, rows)}
	for i := range p.rows {
		p.rows[i] = genny.Float64()
		p.sum += p.rows[i]
	}
	return p
}

func (p *Pink) Generate() float64 {
	p.counter++
	// the row to redraw is the number of trailing zeros of the counter: row 0 every other step,
	// row 1 every fourth, and so on. The last row also takes the counters that run past it
	i := min(bits.TrailingZeros64(p.counter), len(p.rows)-1)
	p.sum -= p.rows[i]
	p.rows[i] = genny.Float64()
	p.sum += p.rows[i]
	// one extra white row keeps the fastest detail
	return (p.sum + genny.Float64()) / float64(len(p.rows)+1)
}

func (p *Pink) Continuous() bool { return true }

func (p *Pink) Done() bool { return false }

// Reset restarts the counter, the rows keep their values
func (p *Pink) Reset() { p.counter = 0 }
