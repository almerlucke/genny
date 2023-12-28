package arpeggio

import "math/rand"

type Mode int

type MirrorMode int

const (
	Up Mode = iota
	Converge
	Alternate
	Random
)

const (
	None MirrorMode = iota
	Exclusive
	Inclusive
)

type Arpeggio[T any] struct {
	values     []T
	mode       Mode
	mirrorMode MirrorMode
	reverse    bool
	pattern    []int
	index      int
	continuous bool
	done       bool
}

func New[T any](values []T, mode Mode, mirrorMode MirrorMode, reverse bool) *Arpeggio[T] {
	a := &Arpeggio[T]{
		values:     values,
		mode:       mode,
		mirrorMode: mirrorMode,
		reverse:    reverse,
	}

	a.generatePattern()

	return a
}

func NewContinuous[T any](sequence []T, mode Mode, mirrorMode MirrorMode, reverse bool) *Arpeggio[T] {
	a := New(sequence, mode, mirrorMode, reverse)
	a.continuous = true

	return a
}

func (a *Arpeggio[T]) NextValue() T {
	if a.done {
		return a.values[a.pattern[len(a.pattern)-1]]
	}

	v := a.values[a.pattern[a.index]]

	a.index++

	if a.index >= len(a.pattern) {
		a.index = 0
		a.done = !a.continuous
	}

	return v
}

func (a *Arpeggio[T]) Continuous() bool {
	return a.continuous
}

func (a *Arpeggio[T]) Reset() {
	a.index = 0
	a.done = false
	a.generatePattern()
}

func (a *Arpeggio[T]) Done() bool {
	return a.done
}

func (a *Arpeggio[T]) generatePattern() {
	var pattern []int

	n := len(a.values)

	switch a.mode {
	case Up:
		for i := 0; i < n; i++ {
			pattern = append(pattern, i)
		}
	case Converge:
		for i := 0; i < n/2; i++ {
			pattern = append(pattern, i, n-1-i)
		}
		if n%2 == 1 {
			pattern = append(pattern, n/2)
		}
	case Alternate:
		for i := 1; i < n; i++ {
			pattern = append(pattern, 0, i)
		}
	case Random:
		for i := 0; i < n; i++ {
			pattern = append(pattern, i)
		}
		rand.Shuffle(len(pattern), func(i, j int) { pattern[i], pattern[j] = pattern[j], pattern[i] })
	}

	n = len(pattern)

	if a.mirrorMode == Exclusive {
		for i := n - 2; i >= 1; i-- {
			pattern = append(pattern, pattern[i])
		}
	} else if a.mirrorMode == Inclusive {
		for i := n - 1; i >= 0; i-- {
			pattern = append(pattern, pattern[i])
		}
	}

	n = len(pattern)

	if a.reverse {
		for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
			pattern[i], pattern[j] = pattern[j], pattern[i]
		}
	}

	a.pattern = pattern
}
