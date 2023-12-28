package flatten

import "github.com/almerlucke/genny"

// Flatten flattens a generator slice values
type Flatten[T any] struct {
	gen     genny.Generator[[]T]
	current []T
	index   int
	done    bool
}

// NewFlatten creates a Flatten which flattens the output of a slice generator
func NewFlatten[T any](gen genny.Generator[[]T]) *Flatten[T] {
	f := &Flatten[T]{gen: gen}

	f.current = gen.NextValue()

	return f
}

// NextValue gets the next value from the current slice
func (f *Flatten[T]) NextValue() (value T) {
	if f.Done() {
		return
	}

	value = f.current[f.index]
	f.index++

	if f.index >= len(f.current) {
		f.index = 0
		if f.gen.Done() {
			f.done = true
		} else {
			f.current = f.gen.NextValue()
		}
	}

	return
}

// Continuous returns if the slice generator is continuous
func (f *Flatten[T]) Continuous() bool {
	return f.gen.Continuous()
}

// Reset the flattener
func (f *Flatten[T]) Reset() {
	f.gen.Reset()
	f.current = f.gen.NextValue()
	f.index = 0
	f.done = false
}

// Done checks if Flatten is done
func (f *Flatten[T]) Done() bool {
	return f.done
}
