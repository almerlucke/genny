package sequence

// Sequence of values
type Sequence[T any] struct {
	values     []T
	index      int
	continuous bool
	done       bool
}

// NewSequence returns a new non-continuous sequence
func NewSequence[T any](values ...T) *Sequence[T] {
	return &Sequence[T]{
		values:     values,
		index:      0,
		continuous: false,
	}
}

// NewContinuousSequence returns a new continuous (looping) sequence
func NewContinuousSequence[T any](values ...T) *Sequence[T] {
	return &Sequence[T]{
		values:     values,
		index:      0,
		continuous: true,
	}
}

// NextValue returns the next value in a sequence
func (s *Sequence[T]) NextValue() T {
	if s.done {
		return s.values[len(s.values)-1]
	}

	r := s.values[s.index]

	s.index++

	if s.index >= len(s.values) {
		s.index = 0
		s.done = !s.continuous
	}

	return r
}

// Continuous returns true if the sequence is looping
func (s *Sequence[T]) Continuous() bool {
	return s.continuous
}

// Done can be checked for a non-continuous generator to see if the generator is finished generating values
func (s *Sequence[T]) Done() bool {
	return s.done
}

// Reset the sequence
func (s *Sequence[T]) Reset() {
	s.index = 0
	s.done = false
}
