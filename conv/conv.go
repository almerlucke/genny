package conv

import "github.com/almerlucke/genny"

// Converter can convert from one type to the other
type Converter[T1, T2 any] interface {
	Convert(T1) T2
}

// Convert is a generator wrapping another generator and converting the output to another type
type Convert[T1, T2 any] struct {
	converter Converter[T1, T2]
	gen       genny.Generator[T1]
}

// New creates a new Cast object
func New[T1, T2 any](gen genny.Generator[T1], converter Converter[T1, T2]) *Convert[T1, T2] {
	return &Convert[T1, T2]{gen: gen, converter: converter}
}

// Generate calls gen next value and casts it to T2
func (c *Convert[T1, T2]) Generate() T2 {
	return c.converter.Convert(c.gen.Generate())
}

// Continuous returns wrapped gen continuous
func (c *Convert[T1, T2]) Continuous() bool {
	return c.gen.Continuous()
}

// Done returns wrapped gen done
func (c *Convert[T1, T2]) Done() bool {
	return c.gen.Done()
}

// Reset performs wrapped gen reset
func (c *Convert[T1, T2]) Reset() {
	c.gen.Reset()
}

type anyConvert[T any] struct{}

func (ac *anyConvert[T]) Convert(t T) any {
	return t
}

func ToAny[T any](g genny.Generator[T]) *Convert[T, any] {
	return New[T, any](g, &anyConvert[T]{})
}
