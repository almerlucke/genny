// Package clock makes ticks from a step counter, to divide a clock and to run patterns of
// different lengths against each other
package clock

// Divider ticks on every nth step
type Divider struct {
	n      int
	offset int
	step   int
}

// NewDivider creates a divider that ticks when (step + offset) is a multiple of n. The first call
// is step 0, so with offset 0 the first call ticks. It panics if n is less than 1
func NewDivider(n, offset int) *Divider {
	if n < 1 {
		panic("clock: divider n must be at least 1")
	}
	return &Divider{n: n, offset: offset}
}

func (d *Divider) Generate() bool {
	t := ((d.step+d.offset)%d.n+d.n)%d.n == 0
	d.step++
	return t
}

func (d *Divider) Continuous() bool { return true }

func (d *Divider) Done() bool { return false }

func (d *Divider) Reset() { d.step = 0 }

// Poly runs cycles of different lengths against each other (polyrhythm). Every call is one step,
// and returns for each cycle if it starts a new round on this step. The cycles realign every
// LCM of the lengths
type Poly struct {
	lengths []int
	out     []bool
	step    int
}

// NewPoly creates a polyrhythm clock, one cycle for each length. It panics if a length is less than 1
func NewPoly(lengths ...int) *Poly {
	for _, v := range lengths {
		if v < 1 {
			panic("clock: cycle length must be at least 1")
		}
	}
	return &Poly{lengths: append([]int{}, lengths...), out: make([]bool, len(lengths))}
}

// Generate returns a slice reused between calls
func (p *Poly) Generate() []bool {
	for i, l := range p.lengths {
		p.out[i] = p.step%l == 0
	}
	p.step++
	return p.out
}

// Cycle returns the number of steps after which all cycles realign
func (p *Poly) Cycle() int {
	c := 1
	for _, l := range p.lengths {
		c = c / gcd(c, l) * l
	}
	return c
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func (p *Poly) Continuous() bool { return true }

func (p *Poly) Done() bool { return false }

func (p *Poly) Reset() { p.step = 0 }
