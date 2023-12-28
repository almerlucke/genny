package transform

import "github.com/almerlucke/genny"

type Transformer[T any] interface {
	Transform(T) T
	Reset()
}

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

func New[T any](gen genny.Generator[T], t Transformer[T]) *Transform[T] {
	return &Transform[T]{
		gen:         gen,
		transformer: t,
	}
}

func (t *Transform[T]) NextValue() T {
	return t.transformer.Transform(t.gen.NextValue())
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
