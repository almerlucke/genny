package and

import (
	"github.com/almerlucke/genny"
	"math/rand"
)

// And is a sequence of Generators
type And[T any] struct {
	gens       []genny.Generator[T]
	current    genny.Generator[T]
	index      int
	continuous bool
	done       bool
}

func New[T any](gens ...genny.Generator[T]) *And[T] {
	return &And[T]{
		gens:    gens,
		current: gens[0],
	}
}

func NewContinuous[T any](gens ...genny.Generator[T]) *And[T] {
	return &And[T]{
		gens:       gens,
		continuous: true,
		current:    gens[0],
	}
}

// NextValue from And, return next value from current generator, update based on continuous or not
func (a *And[T]) NextValue() (value T) {
	gotoNext := false

	value = a.current.NextValue()

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
	a.done = false
	a.current = a.gens[0]
	for _, gen := range a.gens {
		gen.Reset()
	}
}

// Randomize shuffles the generators randomly
func (a *And[T]) Randomize() {
	rand.Shuffle(len(a.gens), func(i, j int) {
		a.gens[i], a.gens[j] = a.gens[j], a.gens[i]
	})
}
