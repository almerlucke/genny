package genny

type Generator[T any] interface {
	// Generate generates a value
	Generate() T
	// Continuous should return false if Done() and Reset() are used by this generator,
	// this allows for generators that can have an end state
	Continuous() bool
	// Done returns true if the generator is done generating values
	Done() bool
	// Reset the generator
	Reset()
}
