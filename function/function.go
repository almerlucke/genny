package function

// Function is a generator that calls a func to generate next values, the function is passed a context
type Function[T any] struct {
	ctx any
	f   func(any) T
}

// NewFunction creates a new function generator
func NewFunction[T any](ctx any, f func(any) T) *Function[T] {
	return &Function[T]{ctx: ctx, f: f}
}

// NextValue calls the internal function with ctx to generate a new value
func (f *Function[T]) NextValue() T {
	return f.f(f.ctx)
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
