// Package sequence provides Sequence, a generator that plays a fixed list of values in order, once or
// looping.
package sequence

import "github.com/almerlucke/genny"

// Sequence of values
type Sequence[T any] struct {
	values     []T
	index      int
	continuous bool
	done       bool
}

// New returns a new non-continuous sequence. A sequence without values is done at once and
// generates the zero value
func New[T any](values ...T) *Sequence[T] {
	return &Sequence[T]{
		values: values,
		index:  0,
		done:   len(values) == 0,
	}
}

// NewLoop returns a new continuous (looping) sequence
func NewLoop[T any](values ...T) *Sequence[T] {
	s := New(values...)
	s.continuous = true
	return s
}

// Generate returns the next value in a sequence
func (s *Sequence[T]) Generate() T {
	if len(s.values) == 0 {
		var zero T
		return zero
	}

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
	s.done = len(s.values) == 0
}

// Randomize shuffles the sequence values randomly
func (s *Sequence[T]) Randomize() {
	genny.Shuffle(len(s.values), func(i, j int) {
		s.values[i], s.values[j] = s.values[j], s.values[i]
	})
}
