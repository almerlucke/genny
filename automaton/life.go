package automaton

import (
	"fmt"
	"strings"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/rest"
	"github.com/almerlucke/genny/scale"
)

// Rule is a two dimensional totalistic rule in B/S notation: a dead cell is born when its number of live
// neighbours (of eight) is in the B list, a live cell survives when it is in the S list.
// "B3/S23" is Conway's Game of Life, "B36/S23" is HighLife
type Rule struct {
	birth, survive [9]bool
}

// ParseRule reads a rule in B/S notation such as "B3/S23" (case does not matter, the S part may be
// empty: "B2/S"). It returns an error for anything else
func ParseRule(s string) (Rule, error) {
	var r Rule
	parts := strings.Split(strings.ToUpper(strings.TrimSpace(s)), "/")
	if len(parts) != 2 || !strings.HasPrefix(parts[0], "B") || !strings.HasPrefix(parts[1], "S") {
		return r, fmt.Errorf("automaton: rule %q is not in B/S notation like B3/S23", s)
	}
	for i, part := range parts {
		for _, c := range part[1:] {
			if c < '0' || c > '8' {
				return r, fmt.Errorf("automaton: rule %q has a neighbour count %q outside 0..8", s, c)
			}
			if i == 0 {
				r.birth[c-'0'] = true
			} else {
				r.survive[c-'0'] = true
			}
		}
	}
	return r, nil
}

// String writes the rule in B/S notation
func (r Rule) String() string {
	var b strings.Builder
	b.WriteByte('B')
	for n, v := range r.birth {
		if v {
			b.WriteByte(byte('0' + n))
		}
	}
	b.WriteString("/S")
	for n, v := range r.survive {
		if v {
			b.WriteByte(byte('0' + n))
		}
	}
	return b.String()
}

// Rules in B/S notation
const (
	// Conway is Conway's Game of Life
	Conway = "B3/S23"
	// HighLife has a replicator next to Life's patterns
	HighLife = "B36/S23"
	// DayAndNight is symmetric between live and dead cells
	DayAndNight = "B3678/S34678"
	// Seeds: every live cell dies, a dead cell with two neighbours is born
	Seeds = "B2/S"
)

// Pattern is a picture of cells: '#', 'o' or 'O' is a live cell, anything else dead. The rows may have
// different lengths
type Pattern []string

// Some known patterns
var (
	// Glider moves one cell diagonally (down and to the right) every 4 generations
	Glider = Pattern{".#.", "..#", "###"}
	// Blinker has period 2: a bar of three that turns
	Blinker = Pattern{"###"}
	// Block is a still life
	Block = Pattern{"##", "##"}
	// RPentomino is a small start that stays chaotic for 1103 generations on an infinite grid
	RPentomino = Pattern{".##", "##.", ".#."}
	// Toad has period 2
	Toad = Pattern{".###", "###."}
	// LWSS is the lightweight spaceship, which moves two cells to the right every 4 generations
	LWSS = Pattern{".####", "#...#", "....#", "#..#."}
)

// Life is a two dimensional cellular automaton on a torus (the edges wrap around), by default Conway's
// Game of Life. Set the start cells with Set, Place, Fill or Randomize before the first Step: Reset goes
// back to them
type Life struct {
	w, h       int
	rule       Rule
	cells      []bool
	next       []bool
	init       []bool
	generation int
}

// NewLife creates an empty grid of w by h cells with a rule in B/S notation (see ParseRule). It panics
// if the size is less than 1 or the rule is invalid
func NewLife(w, h int, rule string) *Life {
	if w < 1 || h < 1 {
		panic("automaton: a life grid must be at least 1 by 1")
	}
	r, err := ParseRule(rule)
	if err != nil {
		panic(err.Error())
	}
	return &Life{
		w: w, h: h, rule: r,
		cells: make([]bool, w*h), next: make([]bool, w*h), init: make([]bool, w*h),
	}
}

// NewLifeRandom creates a grid in which every cell is alive with the given probability (clamped to
// [0, 1]), drawn from the random stream. It panics like NewLife
func NewLifeRandom(w, h int, rule string, density float64) *Life {
	l := NewLife(w, h, rule)
	l.Randomize(density)
	return l
}

// Width is the number of columns
func (l *Life) Width() int { return l.w }

// Height is the number of rows
func (l *Life) Height() int { return l.h }

// Generation is the number of steps since the start
func (l *Life) Generation() int { return l.generation }

func (l *Life) index(x, y int) int {
	return ((y%l.h+l.h)%l.h)*l.w + (x%l.w+l.w)%l.w
}

// Alive tells if the cell at column x and row y is alive; coordinates wrap around
func (l *Life) Alive(x, y int) bool { return l.cells[l.index(x, y)] }

// Set makes a cell alive or dead; coordinates wrap around. It changes the start state when no step was taken yet
func (l *Life) Set(x, y int, alive bool) {
	l.cells[l.index(x, y)] = alive
	if l.generation == 0 {
		copy(l.init, l.cells)
	}
}

// Place puts a pattern with its top left corner at column x and row y (wrapping around), dead cells
// of the pattern leave the grid as it is
func (l *Life) Place(p Pattern, x, y int) {
	for dy, row := range p {
		for dx, c := range row {
			if c == '#' || c == 'o' || c == 'O' {
				l.Set(x+dx, y+dy, true)
			}
		}
	}
}

// Randomize makes every cell alive with the given probability, clamped to [0, 1]
func (l *Life) Randomize(density float64) {
	for i := range l.cells {
		l.cells[i] = genny.Float64() < density
	}
	if l.generation == 0 {
		copy(l.init, l.cells)
	}
}

// Population is the number of live cells
func (l *Life) Population() int {
	n := 0
	for _, c := range l.cells {
		if c {
			n++
		}
	}
	return n
}

// Step calculates the next generation
func (l *Life) Step() {
	for y := range l.h {
		for x := range l.w {
			n := 0
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					if (dx != 0 || dy != 0) && l.cells[l.index(x+dx, y+dy)] {
						n++
					}
				}
			}
			i := y*l.w + x
			if l.cells[i] {
				l.next[i] = l.rule.survive[n]
			} else {
				l.next[i] = l.rule.birth[n]
			}
		}
	}
	l.cells, l.next = l.next, l.cells
	l.generation++
}

// Row returns a copy of row y (wrapping around) of the current generation
func (l *Life) Row(y int) []bool {
	out := make([]bool, l.w)
	for x := range out {
		out[x] = l.cells[l.index(x, y)]
	}
	return out
}

// Column returns a copy of column x (wrapping around) of the current generation
func (l *Life) Column(x int) []bool {
	out := make([]bool, l.h)
	for y := range out {
		out[y] = l.cells[l.index(x, y)]
	}
	return out
}

// Reset goes back to the start cells
func (l *Life) Reset() {
	copy(l.cells, l.init)
	l.generation = 0
}

// String draws the grid with '#' for live cells and '.' for dead ones, a line per row
func (l *Life) String() string {
	var b strings.Builder
	for y := range l.h {
		if y > 0 {
			b.WriteByte('\n')
		}
		for x := range l.w {
			if l.cells[y*l.w+x] {
				b.WriteByte('#')
			} else {
				b.WriteByte('.')
			}
		}
	}
	return b.String()
}

// LifeGate plays a row or column of a Life grid cell by cell as a rhythm: a live cell is a hit. After the
// whole line is played the grid steps to the next generation, so the rhythm changes every cycle
type LifeGate struct {
	life     *Life
	index    int
	vertical bool
	line     []bool
	pos      int
}

// NewRowGate creates a gate that reads row y of the grid (a cycle of Width steps)
func NewRowGate(l *Life, y int) *LifeGate { return &LifeGate{life: l, index: y} }

// NewColumnGate creates a gate that reads column x of the grid (a cycle of Height steps)
func NewColumnGate(l *Life, x int) *LifeGate { return &LifeGate{life: l, index: x, vertical: true} }

func (g *LifeGate) Generate() bool {
	if g.pos == 0 {
		if g.vertical {
			g.line = g.life.Column(g.index)
		} else {
			g.line = g.life.Row(g.index)
		}
	}
	v := g.line[g.pos]
	g.pos++
	if g.pos == len(g.line) {
		g.pos = 0
		g.life.Step()
	}
	return v
}

func (g *LifeGate) Continuous() bool { return true }

func (g *LifeGate) Done() bool { return false }

func (g *LifeGate) Reset() {
	g.life.Reset()
	g.pos = 0
}

// LifePitches plays a row of a Life grid as a melody: the cell in column x is a note when it is alive and
// a rest when dead; its pitch is the scale degree degrees[x] (or x itself without degrees; a short list
// repeats). The grid steps after the whole row is played
type LifePitches struct {
	gate    *LifeGate
	scale   *scale.Scale
	degrees []int
	pos     int
}

// NewLifePitches creates a generator of optional pitches from row y. It panics if the scale is nil
func NewLifePitches(l *Life, y int, s *scale.Scale, degrees ...int) *LifePitches {
	if s == nil {
		panic("automaton: life pitches need a scale")
	}
	return &LifePitches{gate: NewRowGate(l, y), scale: s, degrees: degrees}
}

func (p *LifePitches) Generate() rest.Optional[float64] {
	x := p.pos
	p.pos = (p.pos + 1) % p.gate.life.Width()
	if !p.gate.Generate() {
		return rest.None[float64]()
	}
	deg := x
	if len(p.degrees) > 0 {
		deg = p.degrees[x%len(p.degrees)]
	}
	return rest.Some(p.scale.Degree(deg))
}

func (p *LifePitches) Continuous() bool { return true }

func (p *LifePitches) Done() bool { return false }

func (p *LifePitches) Reset() {
	p.gate.Reset()
	p.pos = 0
}

// LifeDensity generates the fraction of live cells of the grid (0..1) and steps to the next generation
// every time, a slowly moving modulation signal that can drive the density of a texture
type LifeDensity struct{ life *Life }

// NewLifeDensity creates a density generator
func NewLifeDensity(l *Life) *LifeDensity { return &LifeDensity{life: l} }

func (d *LifeDensity) Generate() float64 {
	v := float64(d.life.Population()) / float64(d.life.w*d.life.h)
	d.life.Step()
	return v
}

func (d *LifeDensity) Continuous() bool { return true }

func (d *LifeDensity) Done() bool { return false }

func (d *LifeDensity) Reset() { d.life.Reset() }
