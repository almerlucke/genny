package constant

// Constant value
type Constant[T any] struct {
	value T
}

// NewConstant returns a new constant value
func NewConstant[T any](value T) *Constant[T] {
	return &Constant[T]{
		value: value,
	}
}

// NextValue will always return the constant value
func (c *Constant[T]) NextValue() T {
	return c.value
}

// Continuous will always return true for a constant
func (c *Constant[T]) Continuous() bool {
	return true
}

// Done will always return false for a constant
func (c *Constant[T]) Done() bool {
	return false
}

// Reset does nothing for a constant
func (c *Constant[T]) Reset() {
}
