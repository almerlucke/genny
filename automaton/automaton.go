// Package automaton has a one dimensional cellular automaton (Wolfram rules, for instance 30 or
// 110) whose rows make rhythm patterns
package automaton

import "github.com/almerlucke/genny"

// Automaton is a ring of cells that is updated by a rule. Generate returns the current row and
// then calculates the next generation
type Automaton struct {
	rule  uint8
	init  []bool
	cells []bool
	next  []bool
	out   []bool
}

// New creates an automaton of width cells with one live cell in the middle. It panics if width is less than 1
func New(rule uint8, width int) *Automaton {
	checkWidth(width)
	cells := make([]bool, width)
	cells[len(cells)/2] = true
	return NewWithCells(rule, cells)
}

// NewRandom creates an automaton with random start cells. It panics if width is less than 1
func NewRandom(rule uint8, width int) *Automaton {
	checkWidth(width)
	cells := make([]bool, width)
	for i := range cells {
		cells[i] = genny.Float64() < 0.5
	}
	return NewWithCells(rule, cells)
}

// NewWithCells creates an automaton with the given start cells. It panics if there are no cells
func NewWithCells(rule uint8, cells []bool) *Automaton {
	checkWidth(len(cells))
	a := &Automaton{
		rule:  rule,
		init:  append([]bool{}, cells...),
		cells: append([]bool{}, cells...),
		next:  make([]bool, len(cells)),
		out:   make([]bool, len(cells)),
	}
	return a
}

func checkWidth(width int) {
	if width < 1 {
		panic("automaton: width must be at least 1")
	}
}

// Width is the number of cells
func (a *Automaton) Width() int { return len(a.cells) }

// Generate returns the current row (a slice that is reused) and steps to the next generation
func (a *Automaton) Generate() []bool {
	copy(a.out, a.cells)
	a.step()
	return a.out
}

func (a *Automaton) step() {
	n := len(a.cells)
	for i := range n {
		idx := 0
		if a.cells[(i+n-1)%n] {
			idx |= 4
		}
		if a.cells[i] {
			idx |= 2
		}
		if a.cells[(i+1)%n] {
			idx |= 1
		}
		a.next[i] = a.rule>>idx&1 == 1
	}
	a.cells, a.next = a.next, a.cells
}

func (a *Automaton) Continuous() bool { return true }

func (a *Automaton) Done() bool { return false }

func (a *Automaton) Reset() { copy(a.cells, a.init) }

// Gate plays the rows of an automaton cell by cell, so a rule becomes an endless rhythm: the
// pattern drifts a little every width steps
type Gate struct {
	a   *Automaton
	row []bool
	pos int
}

// NewGate creates a gate generator from an automaton
func NewGate(a *Automaton) *Gate { return &Gate{a: a, row: make([]bool, a.Width())} }

func (g *Gate) Generate() bool {
	if g.pos == 0 {
		copy(g.row, g.a.Generate())
	}
	v := g.row[g.pos]
	g.pos = (g.pos + 1) % len(g.row)
	return v
}

func (g *Gate) Continuous() bool { return true }

func (g *Gate) Done() bool { return false }

func (g *Gate) Reset() {
	g.a.Reset()
	g.pos = 0
}
