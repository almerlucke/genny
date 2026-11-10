// Package and provides And, a generator that plays a list of generators one after the other,
// optionally looping.
package and

import (
	"github.com/almerlucke/genny"
)

// And is a sequence of Generators. An And without generators is done at once and generates the zero value
type And[T any] struct {
	gens       []genny.Generator[T]
	current    genny.Generator[T]
	index      int
	continuous bool
	done       bool
}

// New creates an And that plays the generators once, in order
func New[T any](gens ...genny.Generator[T]) *And[T] {
	a := &And[T]{gens: gens, done: len(gens) == 0}
	if len(gens) > 0 {
		a.current = gens[0]
	}
	return a
}

// NewLoop creates an And that loops over the generators forever
func NewLoop[T any](gens ...genny.Generator[T]) *And[T] {
	a := New(gens...)
	a.continuous = true
	return a
}

// Generate from And, return next value from current generator, update based on continuous or not
func (a *And[T]) Generate() (value T) {
	if a.current == nil {
		return
	}

	gotoNext := false

	value = a.current.Generate()

	if a.current.Continuous() {
		gotoNext = true
	} else if a.current.Done() {
		gotoNext = true
		a.current.Reset()
	}

	if gotoNext {
		a.index += 1
		if a.index >= len(a.gens) {
			a.done = !a.continuous
			a.index = 0
		}
		a.current = a.gens[a.index]
	}

	return
}

// Continuous checks if And is continuous
func (a *And[T]) Continuous() bool {
	return a.continuous
}

// Done checks if And is done
func (a *And[T]) Done() bool {
	return a.done
}

// Reset all generators
func (a *And[T]) Reset() {
	a.index = 0
	a.done = len(a.gens) == 0
	if len(a.gens) > 0 {
		a.current = a.gens[0]
	}
	for _, gen := range a.gens {
		gen.Reset()
	}
}

// Randomize shuffles the generators randomly
func (a *And[T]) Randomize() {
	genny.Shuffle(len(a.gens), func(i, j int) {
		a.gens[i], a.gens[j] = a.gens[j], a.gens[i]
	})

	if len(a.gens) > 0 {
		a.current = a.gens[a.index]
	}
}
