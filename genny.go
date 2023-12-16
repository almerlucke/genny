package genny

type Generator[T any] interface {
	// NextValue generates a value
	NextValue() T
	// Continuous returns false if done and reset should be used,
	// this allows for generators that can have an end state
	Continuous() bool
	// Done checks if the generator is done generating values
	Done() bool
	// Reset the value generator
	Reset()
}
