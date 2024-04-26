package repeat

import (
	"github.com/almerlucke/genny"
	"math/rand"
)

// Repeat can make a continuous generator non-continuous by only repeating NextValue() n times,
// or can extend a non-continuous generator by repeating it n times. N is got by calling f()
type Repeat[T any] struct {
	gen     genny.Generator[T]
	f       func() int
	n       int
	lastVal T
}

// New creates a new repeat generator that repeats n times
func New[T any](gen genny.Generator[T], n int) *Repeat[T] {
	return NewWithFunc[T](gen, func() int { return n })
}

// NewRand creates a new repeat generator that repeats randomly between n1 and n2 times
func NewRand[T any](gen genny.Generator[T], n1 int, n2 int) *Repeat[T] {
	dif := n2 + 1 - n1
	return NewWithFunc[T](gen, func() int {
		return n1 + rand.Intn(dif)
	})
}

// NewWithFunc creates a new repeat generator that repeats f() times
func NewWithFunc[T any](gen genny.Generator[T], f func() int) *Repeat[T] {
	return &Repeat[T]{
		gen: gen,
		f:   f,
		n:   f(),
	}
}

// Generate generates the next value if n > 0, otherwise returns the last value generated
func (r *Repeat[T]) Generate() T {
	v := r.lastVal

	if r.n > 0 {
		v = r.gen.Generate()

		r.lastVal = v

		if r.gen.Continuous() {
			r.n--
		} else if r.gen.Done() {
			r.n--
			if r.n > 0 {
				r.gen.Reset()
			}
		}
	}

	return v
}

// Continuous will always be false for Repeat
func (r *Repeat[T]) Continuous() bool {
	return false
}

// Done is true if n == 0
func (r *Repeat[T]) Done() bool {
	return r.n == 0
}

// Reset the Repeat generator, n is calculated again by calling f()
func (r *Repeat[T]) Reset() {
	if !r.gen.Continuous() {
		// only reset if gen is not continuous
		r.gen.Reset()
	}

	r.n = r.f()
}
