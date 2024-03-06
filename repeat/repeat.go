package repeat

import (
	"github.com/almerlucke/genny"
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

// NewWithFunc creates a new repeat generator that repeats f() times
func NewWithFunc[T any](gen genny.Generator[T], f func() int) *Repeat[T] {
	return &Repeat[T]{
		gen: gen,
		f:   f,
		n:   f(),
	}
}

// NextValue generates the next value if n > 0, otherwise returns the last value generated
func (r *Repeat[T]) NextValue() T {
	v := r.lastVal

	if r.n > 0 {
		v = r.gen.NextValue()
		r.lastVal = v
		r.n--
		if r.gen.Done() && r.n > 0 {
			// reset for the remaining n times
			r.gen.Reset()
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
