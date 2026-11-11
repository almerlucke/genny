// Package listtools has small generators that reshape the values of other generators or lists, after
// the list tools of AC Toolbox: jumping through a list, stuttering and grouping values, growing
// series, filling a bar with durations and turning values into rests.
package listtools

import (
	"math"

	"github.com/almerlucke/genny"
)

// Jump plays chunks of a list: chunk values taken step positions apart, then a leap to a new start
type Jump[T any] struct {
	values            []T
	chunk, step, leap int
	start, taken      int
}

// NewJump creates a looping generator over values: from the start position (0) it plays chunk values,
// each step positions after the previous one, then the start moves leap positions and the next chunk
// begins. Positions wrap around the list, and step and leap may be negative. With values 0..7,
// chunk 3, step 1 and leap 4 it plays 0 1 2, 4 5 6, 0 1 2, ... It panics if chunk is less than 1.
// Without values it is done at once and generates the zero value
func NewJump[T any](values []T, chunk, step, leap int) *Jump[T] {
	if chunk < 1 {
		panic("listtools: chunk must be at least 1")
	}
	return &Jump[T]{values: append([]T{}, values...), chunk: chunk, step: step, leap: leap}
}

func (j *Jump[T]) Generate() T {
	var zero T
	n := len(j.values)
	if n == 0 {
		return zero
	}
	pos := j.start + j.taken*j.step
	v := j.values[((pos%n)+n)%n]
	j.taken++
	if j.taken == j.chunk {
		j.taken = 0
		j.start = ((j.start+j.leap)%n + n) % n
	}
	return v
}

func (j *Jump[T]) Continuous() bool { return len(j.values) > 0 }

func (j *Jump[T]) Done() bool { return len(j.values) == 0 }

func (j *Jump[T]) Reset() { j.start, j.taken = 0, 0 }

// Stutter repeats a value after every n values
type Stutter[T any] struct {
	gen          genny.Generator[T]
	n, k         int
	count, extra int
	last         T
}

// NewStutter creates a generator that after every n values of gen repeats the last one k more times.
// It panics if n is less than 1 or k is negative
func NewStutter[T any](gen genny.Generator[T], n, k int) *Stutter[T] {
	if n < 1 || k < 0 {
		panic("listtools: stutter needs n of at least 1 and a k that is not negative")
	}
	return &Stutter[T]{gen: gen, n: n, k: k}
}

func (s *Stutter[T]) Generate() T {
	if s.extra > 0 {
		s.extra--
		return s.last
	}
	s.last = s.gen.Generate()
	s.count++
	if s.count == s.n {
		s.count = 0
		s.extra = s.k
	}
	return s.last
}

func (s *Stutter[T]) Continuous() bool { return s.gen.Continuous() }

func (s *Stutter[T]) Done() bool { return s.gen.Done() && s.extra == 0 }

func (s *Stutter[T]) Reset() {
	s.gen.Reset()
	s.count, s.extra = 0, 0
}

// Group repeats every value k times, where k can come from a generator
type Group[T any] struct {
	gen     genny.Generator[T]
	k       genny.Generator[int]
	current T
	left    int
}

// NewGroup creates a generator that repeats every value of gen: the number of times is taken from k
// whenever a new value starts. A k below 1 counts as 1
func NewGroup[T any](gen genny.Generator[T], k genny.Generator[int]) *Group[T] {
	return &Group[T]{gen: gen, k: k}
}

// constInt is a generator of one number
type constInt int

func (c constInt) Generate() int    { return int(c) }
func (c constInt) Continuous() bool { return true }
func (c constInt) Done() bool       { return false }
func (c constInt) Reset()           {}

// NewGroupN is NewGroup with a fixed number of repeats
func NewGroupN[T any](gen genny.Generator[T], k int) *Group[T] {
	return NewGroup(gen, genny.Generator[int](constInt(k)))
}

func (g *Group[T]) Generate() T {
	if g.left == 0 {
		g.current = g.gen.Generate()
		g.left = max(g.k.Generate(), 1)
	}
	g.left--
	return g.current
}

func (g *Group[T]) Continuous() bool { return g.gen.Continuous() }

func (g *Group[T]) Done() bool { return g.gen.Done() && g.left == 0 }

func (g *Group[T]) Reset() {
	g.gen.Reset()
	g.k.Reset()
	g.left = 0
}

// Mode is how a Grow series gets its next value
type Mode int

const (
	// Geometric multiplies by the amount
	Geometric Mode = iota
	// Arithmetic adds the amount
	Arithmetic
)

// Limit is what a Grow series does when it would leave its range
type Limit int

const (
	// Clip stays at the limit that was reached
	Clip Limit = iota
	// Restart goes back to the start value
	Restart
)

// Grow is a series that starts at a value and grows (or shrinks) step by step
type Grow struct {
	start, amount, lo, hi float64
	mode                  Mode
	limit                 Limit
	current               float64
	started               bool
}

// NewGrow creates a series from start: every next value is the previous times amount (Geometric) or
// plus amount (Arithmetic). When the next value would be outside [lo, hi] it is clipped to the limit
// (Clip) or the series goes back to start (Restart). The start value is clamped to the range. It
// panics if lo is above hi
func NewGrow(start float64, mode Mode, amount, lo, hi float64, limit Limit) *Grow {
	if lo > hi {
		panic("listtools: lo must not be above hi")
	}
	return &Grow{start: min(max(start, lo), hi), amount: amount, lo: lo, hi: hi, mode: mode, limit: limit}
}

func (g *Grow) Generate() float64 {
	if !g.started {
		g.started = true
		g.current = g.start
		return g.current
	}
	next := g.current + g.amount
	if g.mode == Geometric {
		next = g.current * g.amount
	}
	switch {
	case math.IsNaN(next):
		next = g.start
	case next < g.lo || next > g.hi:
		if g.limit == Restart {
			next = g.start
		} else {
			next = min(max(next, g.lo), g.hi)
		}
	}
	g.current = next
	return next
}

func (g *Grow) Continuous() bool { return true }

func (g *Grow) Done() bool { return false }

func (g *Grow) Reset() { g.started = false }

// maxGather bounds GatherUntil when a generator never reaches the total, for instance one of zeros
const maxGather = 1 << 20

// GatherUntil collects values (durations, negative ones are rests) from gen until their absolute
// values add up to total. The last value is trimmed so the sum is exactly total, for instance to fill a
// bar. If gen ends first (or gives nothing but zeros) the values gathered so far are returned. A total
// that is not positive gives no values
func GatherUntil(gen genny.Generator[float64], total float64) []float64 {
	var out []float64
	sum := 0.0
	for i := 0; sum < total && i < maxGather; i++ {
		if !gen.Continuous() && gen.Done() {
			break
		}
		v := gen.Generate()
		if v == 0 || math.IsNaN(v) {
			continue
		}
		if remaining := total - sum; math.Abs(v) >= remaining {
			v = math.Copysign(remaining, v)
		}
		sum += math.Abs(v)
		out = append(out, v)
	}
	return out
}

// PlusMin turns values into rests (negative values) with a probability
type PlusMin struct {
	gen genny.Generator[float64]
	p   float64
}

// NewPlusMin creates a generator that makes each value of gen a rest, the negative of its absolute value,
// with probability p (clamped to [0, 1]). The other values keep their sign
func NewPlusMin(gen genny.Generator[float64], p float64) *PlusMin {
	return &PlusMin{gen: gen, p: p}
}

func (m *PlusMin) Generate() float64 {
	v := m.gen.Generate()
	if genny.Float64() < m.p {
		return -math.Abs(v)
	}
	return v
}

func (m *PlusMin) Continuous() bool { return m.gen.Continuous() }

func (m *PlusMin) Done() bool { return m.gen.Done() }

func (m *PlusMin) Reset() { m.gen.Reset() }

// PositiveNegative gives values a sign in turns
type PositiveNegative struct {
	gen      genny.Generator[float64]
	pos, neg int
	count    int
}

// NewPositiveNegative creates a generator that makes pos values positive, then neg values negative
// (rests), and again. Values keep only their size. It panics if a count is negative or both are zero
func NewPositiveNegative(gen genny.Generator[float64], pos, neg int) *PositiveNegative {
	if pos < 0 || neg < 0 || pos+neg == 0 {
		panic("listtools: counts must not be negative and not both zero")
	}
	return &PositiveNegative{gen: gen, pos: pos, neg: neg}
}

func (p *PositiveNegative) Generate() float64 {
	v := math.Abs(p.gen.Generate())
	negative := p.count >= p.pos
	p.count = (p.count + 1) % (p.pos + p.neg)
	if negative {
		return -v
	}
	return v
}

func (p *PositiveNegative) Continuous() bool { return p.gen.Continuous() }

func (p *PositiveNegative) Done() bool { return p.gen.Done() }

func (p *PositiveNegative) Reset() {
	p.gen.Reset()
	p.count = 0
}
