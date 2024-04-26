package function

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

// Generate calls the internal function with ctx to generate a new value
func (f *Function[T]) Generate() T {
	if f.ctx != nil {
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
