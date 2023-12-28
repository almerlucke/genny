package continuous

import "github.com/almerlucke/genny"

// Continuous makes a non-continuous generator continuous
type Continuous[T any] struct {
	gen genny.Generator[T]
}

// New creates a continuous wrapper around a non-continuous generator
func New[T any](gen genny.Generator[T]) *Continuous[T] {
	return &Continuous[T]{gen: gen}
}

// NextValue gets next value from wrapped generator and resets if generator is done
func (c *Continuous[T]) NextValue() (value T) {
	value = c.gen.NextValue()

	if c.gen.Done() {
		c.gen.Reset()
	}

	return
}

// Continuous will always return true
func (c *Continuous[T]) Continuous() bool {
	return true
}

// Done will always return false
func (c *Continuous[T]) Done() bool {
	return false
}

// Reset the wrapped generator
func (c *Continuous[T]) Reset() {
	c.gen.Reset()
}
