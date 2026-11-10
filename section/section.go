// Package section plays generators one after the other, each for a number of steps (A for 4 bars,
// B for 8 bars, ...)
package section

import "github.com/almerlucke/genny"

// Part is a generator that plays for a number of steps
type Part[T any] struct {
	Gen genny.Generator[T]
	// Steps to play, if 0 or less the part plays until Gen is done (a continuous Gen never ends)
	Steps int
	// Restart resets Gen every time the part starts, otherwise it carries on where it was
	Restart bool
}

// Section plays parts in order
type Section[T any] struct {
	parts      []Part[T]
	index      int
	played     int
	continuous bool
	done       bool
	last       T
}

// New plays the parts once, after the last part the section is done
func New[T any](parts ...Part[T]) *Section[T] {
	s := &Section[T]{parts: parts}
	s.enter()
	return s
}

// NewLoop plays the parts forever
func NewLoop[T any](parts ...Part[T]) *Section[T] {
	s := New(parts...)
	s.continuous = true
	return s
}

func (s *Section[T]) enter() {
	s.played = 0
	if s.index < len(s.parts) && s.parts[s.index].Restart {
		s.parts[s.index].Gen.Reset()
	}
}

// Current is the index of the part that is playing
func (s *Section[T]) Current() int { return s.index }

func (s *Section[T]) Generate() T {
	if s.done || len(s.parts) == 0 {
		return s.last
	}

	p := s.parts[s.index]
	s.last = p.Gen.Generate()
	s.played++

	if (p.Steps > 0 && s.played >= p.Steps) || (p.Steps <= 0 && p.Gen.Done()) {
		s.index++
		if s.index >= len(s.parts) {
			if s.continuous {
				s.index = 0
			} else {
				s.done = true
				return s.last
			}
		}
		s.enter()
	}

	return s.last
}

func (s *Section[T]) Continuous() bool { return s.continuous }

func (s *Section[T]) Done() bool { return s.done }

func (s *Section[T]) Reset() {
	s.index, s.done = 0, false
	for _, p := range s.parts {
		p.Gen.Reset()
	}
	s.enter()
}
