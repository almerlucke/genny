package combine

import "github.com/almerlucke/genny"

// Combine combines multiple generators in a single generator
type Combine[T any] struct {
	gens []genny.Generator[T]
}

// New returns a new combine
func New[T any](gens ...genny.Generator[T]) *Combine[T] {
	return &Combine[T]{
		gens: gens,
	}
}

// NextValue generates a slice by combining all values generated by the different generators
func (c *Combine[T]) NextValue() []T {
	var vs []T

	for _, g := range c.gens {
		vs = append(vs, g.NextValue())
	}

	return vs
}

// Continuous for combine returns false if any of the gens is not continuous
func (c *Combine[T]) Continuous() bool {
	for _, g := range c.gens {
		if !g.Continuous() {
			return false
		}
	}

	return true
}

// Done returns true if any of the gens returns true
func (c *Combine[T]) Done() bool {
	for _, g := range c.gens {
		if g.Done() {
			return true
		}
	}

	return false
}

// Reset calls reset on all gens
func (c *Combine[T]) Reset() {
	for _, g := range c.gens {
		g.Reset()
	}
}
