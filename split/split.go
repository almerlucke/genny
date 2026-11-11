// Package split cuts a duration into smaller ones by recursive splitting: a duration is divided into
// equal parts, each part may be divided again, until the parts would get too short. The result is a
// rhythm of related durations (a half, then two quarters, one of which becomes two eighths, ...) that
// always fills the original duration.
package split

import "github.com/almerlucke/genny"

// Split recursively divides value into the given number of equal branches. A part is divided with
// probability prob (clamped to [0, 1]), but only if its branches would not be shorter than lower.
// The parts are returned in time order and sum to value (up to rounding). A value that is not
// positive gives no parts. It panics if branches is less than 2 or lower is not positive
func Split(value, lower, prob float64, branches int) []float64 {
	if branches < 2 {
		panic("split: branches must be at least 2")
	}
	if !(lower > 0) {
		panic("split: lower must be positive")
	}
	if !(value > 0) {
		return nil
	}
	prob = min(max(prob, 0), 1)
	var out []float64
	var rec func(v float64)
	rec = func(v float64) {
		part := v / float64(branches)
		if part < lower || genny.Float64() >= prob {
			out = append(out, v)
			return
		}
		for range branches {
			rec(part)
		}
	}
	rec(value)
	return out
}

// Metric divides value into n equal parts. It panics if n is less than 1
func Metric(value float64, n int) []float64 {
	if n < 1 {
		panic("split: n must be at least 1")
	}
	out := make([]float64, n)
	for i := range out {
		out[i] = value / float64(n)
	}
	return out
}

// Gen is a looping generator of split durations: every cycle is a new split of the same value
type Gen struct {
	value, lower, prob float64
	branches           int
	current            []float64
	index              int
}

// NewGen creates a generator that plays Split(value, lower, prob, branches) and when it has played
// all parts splits again. It panics for the same arguments as Split, and if value is not positive
func NewGen(value, lower, prob float64, branches int) *Gen {
	if !(value > 0) {
		panic("split: value must be positive")
	}
	g := &Gen{value: value, lower: lower, prob: prob, branches: branches}
	g.current = Split(value, lower, prob, branches) // checks the arguments
	return g
}

func (g *Gen) Generate() float64 {
	if g.index >= len(g.current) {
		g.current = Split(g.value, g.lower, g.prob, g.branches)
		g.index = 0
	}
	v := g.current[g.index]
	g.index++
	return v
}

func (g *Gen) Continuous() bool { return true }

func (g *Gen) Done() bool { return false }

// Reset starts a new split
func (g *Gen) Reset() {
	g.current = Split(g.value, g.lower, g.prob, g.branches)
	g.index = 0
}
