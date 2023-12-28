package cast

import "github.com/almerlucke/genny"

// Caster can cast from one type to the other
type Caster[T1, T2 any] interface {
	Cast(T1) T2
}

// Cast is a generator wrapping another generator and casting the output to another type
type Cast[T1, T2 any] struct {
	caster Caster[T1, T2]
	gen    genny.Generator[T1]
}

// New creates a new Cast object
func New[T1, T2 any](gen genny.Generator[T1], caster Caster[T1, T2]) *Cast[T1, T2] {
	return &Cast[T1, T2]{gen: gen, caster: caster}
}

// NextValue calls gen next value and casts it to T2
func (c *Cast[T1, T2]) NextValue() T2 {
	return c.caster.Cast(c.gen.NextValue())
}

// Continuous returns wrapped gen continuous
func (c *Cast[T1, T2]) Continuous() bool {
	return c.gen.Continuous()
}

// Done returns wrapped gen done
func (c *Cast[T1, T2]) Done() bool {
	return c.gen.Done()
}

// Reset performs wrapped gen reset
func (c *Cast[T1, T2]) Reset() {
	c.gen.Reset()
}
