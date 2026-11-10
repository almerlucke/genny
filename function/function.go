// Package function provides Function, a generator that calls a Go function for every value.
package function

import "github.com/almerlucke/genny"

// Function is a generator that calls a func to generate next values, the function is passed a context
type Function[T any] struct {
	ctx any
	cf  func(any) T
	f   func() T
}

// New creates a new function generator
func New[T any](f func() T) *Function[T] {
	return &Function[T]{ctx: nil, f: f}
}

// NewWithContext creates a new function generator with context
func NewWithContext[T any](ctx any, f func(any) T) *Function[T] {
	return &Function[T]{ctx: ctx, cf: f}
}

// NewRandom generates uniform random values in [low, high)
func NewRandom(low float64, high float64) *Function[float64] {
	return New(func() float64 { return genny.Float64()*(high-low) + low })
}

// Generate calls the internal function with ctx to generate a new value
func (f *Function[T]) Generate() T {
	if f.cf != nil {
		return f.cf(f.ctx)
	}

	return f.f()
}

// Continuous will always return true for Function
func (f *Function[T]) Continuous() bool {
	return true
}

// Done will always return false for Function
func (f *Function[T]) Done() bool {
	return false
}

// Reset does nothing for a Function
func (f *Function[T]) Reset() {
}
