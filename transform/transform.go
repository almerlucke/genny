// Package transform provides Transform, a wrapper that applies a function or Transformer to every
// value of a generator.
package transform

import "github.com/almerlucke/genny"

// Transformer transforms values, Reset is called when the Transform is reset
type Transformer[T any] interface {
	Transform(T) T
	Reset()
}

// Function is a Transformer without state
type Function[T any] func(T) T

func (f Function[T]) Transform(v T) T {
	return f(v)
}

func (f Function[T]) Reset() {
}

// Transform transforms the output of a value
type Transform[T any] struct {
	gen         genny.Generator[T]
	transformer Transformer[T]
}

// New creates a Transform from a function
func New[T any](gen genny.Generator[T], f func(T) T) *Transform[T] {
	return &Transform[T]{
		gen:         gen,
		transformer: Function[T](f),
	}
}

func NewWithTransformer[T any](gen genny.Generator[T], t Transformer[T]) *Transform[T] {
	return &Transform[T]{
		gen:         gen,
		transformer: t,
	}
}

func (t *Transform[T]) Generate() T {
	return t.transformer.Transform(t.gen.Generate())
}

func (t *Transform[T]) Continuous() bool {
	return t.gen.Continuous()
}

func (t *Transform[T]) Reset() {
	t.gen.Reset()
	t.transformer.Reset()
}

func (t *Transform[T]) Done() bool {
	return t.gen.Done()
}
