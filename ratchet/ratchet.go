// Package ratchet repeats a step several times at higher speed (a roll, a stutter). Generators here
// give for every step the start offsets of its repeats, as fractions of the step: [0] is one hit,
// [0, 0.5] two, [0, 1/3, 2/3] three
package ratchet

import "github.com/almerlucke/genny"

// Offsets returns the start offsets of n equally spaced repeats in one step. It panics if n is less than 1
func Offsets(n int) []float64 {
	if n < 1 {
		panic("ratchet: count must be at least 1")
	}
	o := make([]float64, n)
	for i := range o {
		o[i] = float64(i) / float64(n)
	}
	return o
}

// Random ratchets a step with probability p, into 2..maxCount repeats
type Random struct {
	p        float64
	maxCount int
}

// NewRandom creates a random ratchet generator. It panics if maxCount is less than 2
func NewRandom(p float64, maxCount int) *Random {
	if maxCount < 2 {
		panic("ratchet: maxCount must be at least 2")
	}
	return &Random{p: p, maxCount: maxCount}
}

func (r *Random) Generate() []float64 {
	if genny.Float64() < r.p {
		return Offsets(2 + genny.Intn(r.maxCount-1))
	}
	return Offsets(1)
}

func (r *Random) Continuous() bool { return true }

func (r *Random) Done() bool { return false }

func (r *Random) Reset() {}

// Pattern loops through a list of repeat counts, one for each step
type Pattern struct {
	counts []int
	index  int
}

// NewPattern creates a ratchet pattern, for instance NewPattern(1, 1, 2, 1, 1, 3, 1, 4). It panics if there
// are no counts or a count is less than 1
func NewPattern(counts ...int) *Pattern {
	if len(counts) == 0 {
		panic("ratchet: a pattern needs counts")
	}
	for _, c := range counts {
		if c < 1 {
			panic("ratchet: count must be at least 1")
		}
	}
	return &Pattern{counts: append([]int{}, counts...)}
}

func (p *Pattern) Generate() []float64 {
	o := Offsets(p.counts[p.index])
	p.index = (p.index + 1) % len(p.counts)
	return o
}

func (p *Pattern) Continuous() bool { return true }

func (p *Pattern) Done() bool { return false }

func (p *Pattern) Reset() { p.index = 0 }
