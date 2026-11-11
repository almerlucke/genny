// Package sieve implements Xenakis sieves: sets of integers built from residue classes (every
// mod-th number from a start) within a range, combined with union, intersection, complement and
// difference. A sieve can be a pitch set (Scale), a rhythm (Gate, Filter, Interpret) or a way to
// snap values to a set (Funnel).
package sieve

import (
	"slices"
	"sort"

	"github.com/almerlucke/genny"
)

// Sieve is a sorted set of integers within the range [lo, hi]. The zero Sieve is empty. A Sieve is
// immutable: the operations return new sieves
type Sieve struct {
	values []int
	lo     int
	span   int // number of integers in the range
}

func newSieve(lo, hi int) Sieve { return Sieve{lo: lo, span: max(hi-lo+1, 0)} }

func (s Sieve) hi() int { return s.lo + s.span - 1 }

// New is the residue class of mod and start within [lo, hi]: all x with x = start (mod mod) in the
// range, increasing. It panics if mod is less than 1. If hi is less than lo the sieve is empty
func New(mod, start, lo, hi int) Sieve {
	if mod < 1 {
		panic("sieve: mod must be at least 1")
	}
	s := newSieve(lo, hi)
	if s.span == 0 {
		return s
	}
	first := lo + (((start-lo)%mod)+mod)%mod
	for x := first; x <= hi; x += mod {
		s.values = append(s.values, x)
	}
	return s
}

// Of makes a sieve of the given values (duplicates are dropped) within [lo, hi]: values outside the
// range are dropped
func Of(lo, hi int, values ...int) Sieve {
	s := newSieve(lo, hi)
	for _, v := range values {
		if v >= lo && v <= hi {
			s.values = append(s.values, v)
		}
	}
	slices.Sort(s.values)
	s.values = slices.Compact(s.values)
	return s
}

// Values returns the members, increasing
func (s Sieve) Values() []int { return append([]int{}, s.values...) }

// Len is the number of members
func (s Sieve) Len() int { return len(s.values) }

// Range returns the lo and hi the sieve is limited to (hi is below lo for an empty range)
func (s Sieve) Range() (lo, hi int) { return s.lo, s.hi() }

// Contains reports whether x is a member
func (s Sieve) Contains(x int) bool {
	i := sort.SearchInts(s.values, x)
	return i < len(s.values) && s.values[i] == x
}

// hull is the range that covers both sieves, ignoring an empty range
func hull(a, b Sieve) (lo, hi int) {
	switch {
	case a.span == 0:
		return b.lo, b.hi()
	case b.span == 0:
		return a.lo, a.hi()
	}
	return min(a.lo, b.lo), max(a.hi(), b.hi())
}

// Union has the members of either sieve, the range covers both
func (s Sieve) Union(o Sieve) Sieve {
	lo, hi := hull(s, o)
	r := newSieve(lo, hi)
	r.values = append(append([]int{}, s.values...), o.values...)
	slices.Sort(r.values)
	r.values = slices.Compact(r.values)
	return r
}

// Intersection has the members of both sieves, the range covers both
func (s Sieve) Intersection(o Sieve) Sieve {
	lo, hi := hull(s, o)
	r := newSieve(lo, hi)
	for _, v := range s.values {
		if o.Contains(v) {
			r.values = append(r.values, v)
		}
	}
	return r
}

// Difference has the members of s that are not in o
func (s Sieve) Difference(o Sieve) Sieve {
	r := Sieve{lo: s.lo, span: s.span}
	for _, v := range s.values {
		if !o.Contains(v) {
			r.values = append(r.values, v)
		}
	}
	return r
}

// Complement has the numbers of the range that are not members
func (s Sieve) Complement() Sieve {
	r := Sieve{lo: s.lo, span: s.span}
	for x := s.lo; x <= s.hi(); x++ {
		if !s.Contains(x) {
			r.values = append(r.values, x)
		}
	}
	return r
}

// Scale gives the members as a pitch set: semitones above base
func Scale(s Sieve, base float64) []float64 {
	out := make([]float64, len(s.values))
	for i, v := range s.values {
		out[i] = base + float64(v)
	}
	return out
}

// Gen plays the members of a sieve, once or in a loop
type Gen struct {
	values []int
	index  int
	loop   bool
	done   bool
}

// NewGen creates a generator of the members of the sieve. An empty sieve is done at once and
// generates 0. A generator that has played all members keeps generating the last one
func NewGen(s Sieve, loop bool) *Gen {
	return &Gen{values: s.Values(), loop: loop, done: len(s.values) == 0}
}

func (g *Gen) Generate() int {
	if len(g.values) == 0 {
		return 0
	}
	if g.done {
		return g.values[len(g.values)-1]
	}
	v := g.values[g.index]
	g.index++
	if g.index == len(g.values) {
		g.index = 0
		g.done = !g.loop
	}
	return v
}

func (g *Gen) Continuous() bool { return g.loop }

func (g *Gen) Done() bool { return g.done }

func (g *Gen) Reset() {
	g.index = 0
	g.done = len(g.values) == 0
}

// Gate is a looping gate that is true on the steps that are members of a sieve
type Gate struct {
	s      Sieve
	length int
	step   int
}

// NewGate creates a gate for steps 0, 1, 2, ... that is true when the step is a member. After length
// steps it starts again at step 0. It panics if length is less than 1
func NewGate(s Sieve, length int) *Gate {
	if length < 1 {
		panic("sieve: gate length must be at least 1")
	}
	return &Gate{s: s, length: length}
}

func (g *Gate) Generate() bool {
	v := g.s.Contains(g.step)
	g.step = (g.step + 1) % g.length
	return v
}

func (g *Gate) Continuous() bool { return true }

func (g *Gate) Done() bool { return false }

func (g *Gate) Reset() { g.step = 0 }

// Select returns the values at the positions that are members of the sieve, in order
func Select[T any](values []T, s Sieve) []T {
	var out []T
	for _, i := range s.values {
		if i >= 0 && i < len(values) {
			out = append(out, values[i])
		}
	}
	return out
}

// Funnel snaps every value to the nearest of the allowed values. When two allowed values are
// equally near one is picked at random (from the genny stream). Without allowed values the values
// are returned unchanged
func Funnel(values, allowed []float64) []float64 {
	out := make([]float64, len(values))
	for i, v := range values {
		out[i] = funnel(v, allowed)
	}
	return out
}

func funnel(v float64, allowed []float64) float64 {
	if len(allowed) == 0 {
		return v
	}
	best := allowed[0]
	dist := abs(v - best)
	tie := false
	var other float64
	for _, a := range allowed[1:] {
		d := abs(v - a)
		switch {
		case d < dist:
			best, dist, tie = a, d, false
		case d == dist && a != best:
			if !tie {
				other, tie = a, true
			}
		}
	}
	if tie && genny.Intn(2) == 1 {
		return other
	}
	return best
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// Interpret applies a sieve to a rhythm pattern of durations (negative is a rest). The members of
// the sieve are positions in the pattern. Kind 1 turns the values that are not members into rests.
// Kind 2 lets every member absorb the non-members that follow it, so they are added to its
// duration (non-members before the first member become one rest). Kinds 3 and 4 are like 1 and 2
// with the roles inverted: the non-members are the ones that count. The total duration is kept.
// It panics on another kind
func Interpret(pattern []float64, s Sieve, kind int) []float64 {
	if kind < 1 || kind > 4 {
		panic("sieve: kind must be 1, 2, 3 or 4")
	}
	indexed := func(i int) bool { return s.Contains(i) != (kind > 2) }

	var out []float64
	if kind == 1 || kind == 3 {
		for i, v := range pattern {
			if indexed(i) {
				out = append(out, v)
			} else {
				out = append(out, -abs(v))
			}
		}
		return out
	}

	lead := 0.0
	for i, v := range pattern {
		switch {
		case indexed(i):
			out = append(out, v)
		case len(out) == 0:
			lead += abs(v)
		default:
			last := &out[len(out)-1]
			if *last < 0 {
				*last -= abs(v)
			} else {
				*last += abs(v)
			}
		}
	}
	if lead > 0 {
		out = append([]float64{-lead}, out...)
	}
	return out
}
