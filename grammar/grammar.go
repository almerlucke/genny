// Package grammar is a context free rewrite grammar that makes sequences of values. A rule turns a
// symbol into one of several alternatives (chosen at random by weight, with the random stream of
// genny), each a sequence of terminals (values) and symbols that are rewritten in turn. A
// phrase can be built from motifs, motifs from notes, and every derivation is different:
//
//	g := grammar.New[int]().
//		Rule("phrase", grammar.Choice(1, grammar.Ref[int]("motif"), grammar.Ref[int]("motif"), grammar.Term(0))).
//		Rule("motif", grammar.Choice(2, grammar.Term(7), grammar.Term(9)), grammar.Choice(1, grammar.Term(4), grammar.Ref[int]("motif")))
//	notes := g.Expand("phrase", 6)
//
// Recursive rules are cut off by a maximum depth, so every expansion ends
package grammar

import (
	"fmt"
	"slices"

	"github.com/almerlucke/genny"
)

// MaxLen limits the length of an expansion, longer results are cut off
const MaxLen = 1 << 16

// Symbol is a part of an alternative: a terminal value or a reference to a rule
type Symbol[T any] struct {
	name     string
	value    T
	terminal bool
}

// Term is a terminal: the value itself
func Term[T any](value T) Symbol[T] { return Symbol[T]{value: value, terminal: true} }

// Ref refers to the rule of a symbol name, which is rewritten when the grammar is expanded
func Ref[T any](name string) Symbol[T] { return Symbol[T]{name: name} }

// Alt is one alternative of a rule: symbols that replace the rule's symbol, chosen with a weight
type Alt[T any] struct {
	Weight  float64
	Symbols []Symbol[T]
}

// Choice makes an alternative with a weight (relative to the other alternatives of the rule)
func Choice[T any](weight float64, symbols ...Symbol[T]) Alt[T] {
	return Alt[T]{Weight: weight, Symbols: symbols}
}

// Seq makes an alternative of terminal values with weight 1
func Seq[T any](values ...T) Alt[T] {
	syms := make([]Symbol[T], len(values))
	for i, v := range values {
		syms[i] = Term(v)
	}
	return Alt[T]{Weight: 1, Symbols: syms}
}

// Grammar is a set of rules from symbol names to alternatives
type Grammar[T any] struct {
	rules map[string][]Alt[T]
}

// New creates an empty grammar
func New[T any]() *Grammar[T] { return &Grammar[T]{rules: map[string][]Alt[T]{}} }

// Rule defines (or replaces) the rule of a symbol and returns the grammar, to chain. It panics
// if there is no alternative, if a weight is negative or NaN, or if all weights are 0
func (g *Grammar[T]) Rule(name string, alts ...Alt[T]) *Grammar[T] {
	if len(alts) == 0 {
		panic(fmt.Sprintf("grammar: rule %q has no alternatives", name))
	}
	total := 0.0
	for _, a := range alts {
		if !(a.Weight >= 0) {
			panic(fmt.Sprintf("grammar: rule %q has an invalid weight %v", name, a.Weight))
		}
		total += a.Weight
	}
	if total == 0 {
		panic(fmt.Sprintf("grammar: all weights of rule %q are 0", name))
	}
	cp := make([]Alt[T], len(alts))
	for i, a := range alts {
		cp[i] = Alt[T]{Weight: a.Weight, Symbols: slices.Clone(a.Symbols)}
	}
	g.rules[name] = cp
	return g
}

// Check returns an error if the start symbol or a symbol used in a rule has no rule
func (g *Grammar[T]) Check(start string) error {
	if _, ok := g.rules[start]; !ok {
		return fmt.Errorf("grammar: no rule for start symbol %q", start)
	}
	for name, alts := range g.rules {
		for _, a := range alts {
			for _, s := range a.Symbols {
				if _, ok := g.rules[s.name]; !s.terminal && !ok {
					return fmt.Errorf("grammar: rule %q uses %q, which has no rule", name, s.name)
				}
			}
		}
	}
	return nil
}

// pick chooses an alternative by weight
func pick[T any](alts []Alt[T]) Alt[T] {
	total := 0.0
	for _, a := range alts {
		total += a.Weight
	}
	r := genny.Float64() * total
	for _, a := range alts {
		if r < a.Weight {
			return a
		}
		r -= a.Weight
	}
	return alts[len(alts)-1]
}

// shortest is the alternative with the fewest rule references, then the fewest symbols
func shortest[T any](alts []Alt[T]) Alt[T] {
	refs := func(a Alt[T]) int {
		n := 0
		for _, s := range a.Symbols {
			if !s.terminal {
				n++
			}
		}
		return n
	}
	best := alts[0]
	for _, a := range alts[1:] {
		if refs(a) < refs(best) || refs(a) == refs(best) && len(a.Symbols) < len(best.Symbols) {
			best = a
		}
	}
	return best
}

// Expand rewrites the start symbol into terminal values, depth levels of rules deep (1 is the start
// rule alone). Beyond the maximum depth a rule is replaced by its shortest alternative (the one with the fewest references) and
// the references in that one are dropped, so the result is always finite. A depth below 1 gives
// only the shortest expansion of the start rule. Symbols without a rule are dropped, as is everything
// beyond MaxLen. An unknown start symbol gives nil (see Check)
func (g *Grammar[T]) Expand(start string, depth int) []T {
	if _, ok := g.rules[start]; !ok {
		return nil
	}
	depth = max(depth, 0)
	var out []T
	var rec func(name string, d int)
	rec = func(name string, d int) {
		alts, ok := g.rules[name]
		if !ok {
			return
		}
		last := d > depth
		alt := shortest(alts)
		if !last {
			alt = pick(alts)
		}
		for _, s := range alt.Symbols {
			if len(out) >= MaxLen {
				return
			}
			switch {
			case s.terminal:
				out = append(out, s.value)
			case !last:
				rec(s.name, d+1)
			}
		}
	}
	rec(start, 1)
	return out
}

// Player plays the expansion of a grammar value by value, and when it reaches the end derives a new
// expansion: the same grammar, a different phrase every cycle
type Player[T any] struct {
	g     *Grammar[T]
	start string
	depth int
	cur   []T
	pos   int
	begun bool
}

// NewPlayer creates a generator that plays Expand(start, depth) and expands again at the end of
// each cycle. The first expansion is made at the first Generate, so seed before. It panics if the start symbol
// or a used symbol has no rule. If an expansion is empty the generator generates the zero value
// and is done (Done), until Reset
func NewPlayer[T any](g *Grammar[T], start string, depth int) *Player[T] {
	if err := g.Check(start); err != nil {
		panic(err.Error())
	}
	return &Player[T]{g: g, start: start, depth: depth}
}

// Phrase returns the notes of the cycle that is playing (a copy)
func (p *Player[T]) Phrase() []T { return slices.Clone(p.cur) }

func (p *Player[T]) Generate() T {
	if !p.begun || p.pos >= len(p.cur) {
		p.cur, p.pos, p.begun = p.g.Expand(p.start, p.depth), 0, true
	}
	var zero T
	if len(p.cur) == 0 {
		return zero
	}
	v := p.cur[p.pos]
	p.pos++
	return v
}

// Continuous is true: the player goes on with new expansions
func (p *Player[T]) Continuous() bool { return true }

// Done is true when the last expansion was empty
func (p *Player[T]) Done() bool { return p.begun && len(p.cur) == 0 }

// Reset forgets the expansion: the next Generate derives a new one
func (p *Player[T]) Reset() { p.cur, p.pos, p.begun = nil, 0, false }

var _ genny.Generator[int] = (*Player[int])(nil)
