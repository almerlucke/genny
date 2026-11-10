// Package unwrap provides Unwrap, which plays the generators produced by a generator of generators one
// after the other.
package unwrap

import "github.com/almerlucke/genny"

// Unwrap a generator that generates generators
type Unwrap[T any] struct {
	gen     genny.Generator[genny.Generator[T]]
	current genny.Generator[T]
	done    bool
}

// New creates an Unwrap
func New[T any](gen genny.Generator[genny.Generator[T]]) *Unwrap[T] {
	u := &Unwrap[T]{gen: gen, current: gen.Generate()}

	return u
}

// Generate value from current or go to next current generator if available
func (u *Unwrap[T]) Generate() T {
	v := u.current.Generate()

	currentDone := !u.current.Continuous() && u.current.Done()
	if currentDone {
		// ready for when this generator is played again
		u.current.Reset()
	}

	if u.current.Continuous() || currentDone {
		if u.gen.Done() {
			u.done = true
		} else {
			u.current = u.gen.Generate()
		}
	}

	return v
}

// Continuous returns if gen is continuous
func (u *Unwrap[T]) Continuous() bool {
	return u.gen.Continuous()
}

// Done returns if Unwrap is done
func (u *Unwrap[T]) Done() bool {
	return u.done
}

// Reset the generator
func (u *Unwrap[T]) Reset() {
	u.gen.Reset()
	u.done = false
	u.current = u.gen.Generate()
	u.current.Reset()
}
