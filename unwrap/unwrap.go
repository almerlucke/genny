package unwrap

import "github.com/almerlucke/genny"

// Unwrap a generator that generates generators
type Unwrap[T any] struct {
	gen     genny.Generator[genny.Generator[T]]
	current genny.Generator[T]
	done    bool
}

func New[T any](gen genny.Generator[genny.Generator[T]]) *Unwrap[T] {
	u := &Unwrap[T]{gen: gen, current: gen.NextValue()}

	return u
}

// NextValue generate value from current or go to next current generator if available
func (u *Unwrap[T]) NextValue() T {
	v := u.current.NextValue()

	if u.current.Continuous() || u.current.Done() {
		if u.gen.Done() {
			u.done = true
		} else {
			u.current = u.gen.NextValue()
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
}
