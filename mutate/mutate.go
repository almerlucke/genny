// Package mutate plays a phrase that changes a little every time it has been played
package mutate

import (
	"slices"

	"github.com/almerlucke/genny"
)

// Config sets the chance (0..1) of each mutation, every time the phrase has been played
type Config[T any] struct {
	// ReplaceProb is the chance that one note is replaced, with the result of Replace or, if
	// Replace is nil, with a copy of another note of the phrase
	ReplaceProb float64
	Replace     func() T
	// SwapProb is the chance that two notes swap places
	SwapProb float64
	// RotateProb is the chance that the phrase rotates one step, left or right
	RotateProb float64
	// ReverseProb is the chance that the phrase is reversed
	ReverseProb float64
}

// Sequence is a looping phrase with mutations. Reset restores the original phrase
type Sequence[T any] struct {
	orig  []T
	cur   []T
	index int
	cfg   Config[T]
}

// New creates a mutating looping sequence, the values are copied
func New[T any](values []T, cfg Config[T]) *Sequence[T] {
	return &Sequence[T]{
		orig: slices.Clone(values),
		cur:  slices.Clone(values),
		cfg:  cfg,
	}
}

// Values returns a copy of the current phrase
func (s *Sequence[T]) Values() []T { return slices.Clone(s.cur) }

// Mutate applies the mutations once. It is called by Generate at the end of every pass, but can be
// called by hand too
func (s *Sequence[T]) Mutate() {
	n := len(s.cur)
	if n == 0 {
		return
	}
	if genny.Float64() < s.cfg.ReplaceProb {
		i := genny.Intn(n)
		if s.cfg.Replace != nil {
			s.cur[i] = s.cfg.Replace()
		} else {
			s.cur[i] = s.cur[genny.Intn(n)]
		}
	}
	if n > 1 && genny.Float64() < s.cfg.SwapProb {
		i, j := genny.Intn(n), genny.Intn(n)
		s.cur[i], s.cur[j] = s.cur[j], s.cur[i]
	}
	if n > 1 && genny.Float64() < s.cfg.RotateProb {
		if genny.Float64() < 0.5 {
			s.cur = append(s.cur[1:], s.cur[0])
		} else {
			s.cur = append([]T{s.cur[n-1]}, s.cur[:n-1]...)
		}
	}
	if genny.Float64() < s.cfg.ReverseProb {
		slices.Reverse(s.cur)
	}
}

func (s *Sequence[T]) Generate() (v T) {
	if len(s.cur) == 0 {
		return
	}
	v = s.cur[s.index]
	s.index++
	if s.index >= len(s.cur) {
		s.index = 0
		s.Mutate()
	}
	return
}

func (s *Sequence[T]) Continuous() bool { return true }

func (s *Sequence[T]) Done() bool { return false }

func (s *Sequence[T]) Reset() {
	s.cur = slices.Clone(s.orig)
	s.index = 0
}
