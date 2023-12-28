package repeat

import (
	"github.com/almerlucke/genny"
	"math/rand"
)

// Repeat can make a continuous generator non-continuous by only repeating NextValue() n times,
// or can extend a non-continuous generator by repeating it n times. N is calculated by taking a
// random value between given min and max inclusive
type Repeat[T any] struct {
	gen     genny.Generator[T]
	min     int
	max     int
	n       int
	lastVal T
}

// New creates a new repeat generator
func New[T any](gen genny.Generator[T], min int, max int) *Repeat[T] {
	if min > max {
		tmp := min
		min = max
		max = tmp
	}

	return &Repeat[T]{
		gen: gen,
		min: min,
		max: max,
		n:   min + rand.Intn((max-min)+1),
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

// Reset the Repeat generator, n is calculated again in a random way between min and max given
func (r *Repeat[T]) Reset() {
	if !r.gen.Continuous() {
		// only reset if gen is not continuous
		r.gen.Reset()
	}

	r.n = r.min + rand.Intn((r.max-r.min)+1)
}
