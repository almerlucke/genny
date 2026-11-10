// Package flatten provides Flatten, which turns a generator of slices into a generator of the single
// values in those slices.
package flatten

import "github.com/almerlucke/genny"

// Flatten flattens a generator slice values
type Flatten[T any] struct {
	gen     genny.Generator[[]T]
	current []T
	index   int
	done    bool
}

// New creates a Flatten which flattens the output of a slice generator
func New[T any](gen genny.Generator[[]T]) *Flatten[T] {
	f := &Flatten[T]{gen: gen}

	f.load()

	return f
}

// Generate gets the next value from the current slice
func (f *Flatten[T]) Generate() (value T) {
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
			f.load()
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
	f.index = 0
	f.done = false
	f.load()
}

// load fetches the next non-empty slice from the source, empty slices are skipped. When the source
// is done without a non-empty slice, f is done
func (f *Flatten[T]) load() {
	for {
		f.current = f.gen.Generate()
		if len(f.current) > 0 {
			return
		}
		if f.gen.Done() {
			f.done = true
			return
		}
	}
}

// Done checks if Flatten is done
func (f *Flatten[T]) Done() bool {
	return f.done
}
