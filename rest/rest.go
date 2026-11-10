// Package rest adds rests (silent steps) to generators
package rest

import "github.com/almerlucke/genny"

// Optional is a value or a rest
type Optional[T any] struct {
	Value T
	Rest  bool
}

// Some is a played value
func Some[T any](v T) Optional[T] { return Optional[T]{Value: v} }

// None is a rest
func None[T any]() Optional[T] { return Optional[T]{Rest: true} }

// Played is true if the step is not a rest
func (o Optional[T]) Played() bool { return !o.Rest }

// Gated turns a generator into rests and values with a gate generator (for instance euclid.Euclid):
// on a true gate the next value is played, on false it is a rest
type Gated[T any] struct {
	gen     genny.Generator[T]
	gate    genny.Generator[bool]
	consume bool
}

// NewGated plays gen on the true steps of gate. By default a rest consumes a value of gen, so the
// melody keeps its place in time, see Consume
func NewGated[T any](gen genny.Generator[T], gate genny.Generator[bool]) *Gated[T] {
	return &Gated[T]{gen: gen, gate: gate, consume: true}
}

// Consume sets whether a rest takes (and drops) a value from the wrapped generator (default) or
// leaves it for the next played step
func (g *Gated[T]) Consume(consume bool) *Gated[T] {
	g.consume = consume
	return g
}

func (g *Gated[T]) Generate() Optional[T] {
	if g.gate.Generate() {
		return Some(g.gen.Generate())
	}
	if g.consume {
		g.gen.Generate()
	}
	return None[T]()
}

func (g *Gated[T]) Continuous() bool { return g.gen.Continuous() && g.gate.Continuous() }

func (g *Gated[T]) Done() bool { return g.gen.Done() || g.gate.Done() }

func (g *Gated[T]) Reset() {
	g.gen.Reset()
	g.gate.Reset()
}

// Probability plays each step with probability p, otherwise it is a rest
type Probability[T any] struct {
	gen     genny.Generator[T]
	p       float64
	consume bool
}

// NewProbability creates a probability gate, p in [0, 1] is the chance a step is played
func NewProbability[T any](gen genny.Generator[T], p float64) *Probability[T] {
	return &Probability[T]{gen: gen, p: p, consume: true}
}

// SetProbability changes the chance a step is played
func (g *Probability[T]) SetProbability(p float64) { g.p = p }

// Consume sets whether a rest takes (and drops) a value from the wrapped generator (default) or
// leaves it for the next played step
func (g *Probability[T]) Consume(consume bool) *Probability[T] {
	g.consume = consume
	return g
}

func (g *Probability[T]) Generate() Optional[T] {
	if genny.Float64() < g.p {
		return Some(g.gen.Generate())
	}
	if g.consume {
		g.gen.Generate()
	}
	return None[T]()
}

func (g *Probability[T]) Continuous() bool { return g.gen.Continuous() }

func (g *Probability[T]) Done() bool { return g.gen.Done() }

func (g *Probability[T]) Reset() { g.gen.Reset() }
