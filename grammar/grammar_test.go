package grammar_test

import (
	"slices"
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/grammar"
)

func TestExpandDeterministic(t *testing.T) {
	g := grammar.New[string]().
		Rule("S", grammar.Choice(1, grammar.Ref[string]("A"), grammar.Term("x"), grammar.Ref[string]("A"))).
		Rule("A", grammar.Seq("a", "b"))
	if got := g.Expand("S", 3); !slices.Equal(got, []string{"a", "b", "x", "a", "b"}) {
		t.Errorf("%v", got)
	}
}

func TestExpandDepthLimit(t *testing.T) {
	// S -> a S | b: with weight on the recursion the depth must cut it off
	g := grammar.New[string]().
		Rule("S", grammar.Choice(1000, grammar.Term("a"), grammar.Ref[string]("S")), grammar.Choice(0.000001, grammar.Term("b")))
	genny.Seed(1)
	tests := []struct {
		depth   int
		maxLen  int
		wantEnd string
	}{
		{1, 2, "b"}, {5, 6, "b"}, {20, 21, "b"}, {0, 1, "b"}, {-4, 1, "b"},
	}
	for _, tt := range tests {
		got := g.Expand("S", tt.depth)
		if len(got) > tt.maxLen {
			t.Errorf("depth %d: %d values", tt.depth, len(got))
		}
		// beyond the depth the shortest alternative is "b", which has no references
		if len(got) == 0 || got[len(got)-1] != tt.wantEnd {
			t.Errorf("depth %d: %v", tt.depth, got)
		}
	}
}

func TestExpandShortestBeyondDepth(t *testing.T) {
	g := grammar.New[int]().
		Rule("S", grammar.Choice(1, grammar.Term(1), grammar.Ref[int]("A"), grammar.Term(2))).
		Rule("A", grammar.Choice(1, grammar.Term(5), grammar.Ref[int]("B"), grammar.Term(6)), grammar.Choice(1, grammar.Term(9))).
		Rule("B", grammar.Seq(7))
	// depth 1: A is beyond the depth, replaced by its alternative with no references: 9
	if got := g.Expand("S", 1); !slices.Equal(got, []int{1, 9, 2}) {
		t.Errorf("%v", got)
	}
}

func TestExpandWeights(t *testing.T) {
	g := grammar.New[int]().Rule("S", grammar.Choice(3, grammar.Term(1)), grammar.Choice(1, grammar.Term(2)), grammar.Choice(0, grammar.Term(3)))
	genny.Seed(7)
	counts := map[int]int{}
	for range 4000 {
		counts[g.Expand("S", 1)[0]]++
	}
	if counts[3] != 0 {
		t.Error("weight 0 chosen")
	}
	if r := float64(counts[1]) / 4000; r < 0.72 || r > 0.78 {
		t.Errorf("1 came %.2f, want 0.75", r)
	}
}

func TestExpandEdges(t *testing.T) {
	g := grammar.New[int]().Rule("S", grammar.Choice(1, grammar.Ref[int]("missing"), grammar.Term(1)))
	if got := g.Expand("S", 4); !slices.Equal(got, []int{1}) {
		t.Errorf("missing symbols are dropped: %v", got)
	}
	if g.Expand("nope", 3) != nil {
		t.Error("unknown start")
	}
	if g.Check("S") == nil || g.Check("nope") == nil {
		t.Error("Check should complain")
	}
	// an empty alternative is fine
	e := grammar.New[int]().Rule("S", grammar.Choice(1, []grammar.Symbol[int]{}...))
	if got := e.Expand("S", 3); len(got) != 0 {
		t.Errorf("%v", got)
	}
}

func TestExpandMaxLen(t *testing.T) {
	g := grammar.New[int]().Rule("S", grammar.Choice(1, grammar.Ref[int]("S"), grammar.Ref[int]("S"), grammar.Term(1)))
	if got := g.Expand("S", 40); len(got) > grammar.MaxLen {
		t.Errorf("%d values", len(got))
	}
}

func TestRulePanics(t *testing.T) {
	tests := map[string]func(){
		"no alternatives": func() { grammar.New[int]().Rule("S") },
		"negative weight": func() { grammar.New[int]().Rule("S", grammar.Choice(-1, grammar.Term(1))) },
		"all zero":        func() { grammar.New[int]().Rule("S", grammar.Choice(0, grammar.Term(1))) },
		"player unknown":  func() { grammar.NewPlayer(grammar.New[int](), "S", 3) },
		"player missing ref": func() {
			grammar.NewPlayer(grammar.New[int]().Rule("S", grammar.Choice(1, grammar.Ref[int]("x"))), "S", 3)
		},
	}
	for name, f := range tests {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: no panic", name)
				}
			}()
			f()
		}()
	}
}

func TestPlayer(t *testing.T) {
	g := grammar.New[int]().
		Rule("S", grammar.Choice(1, grammar.Term(1), grammar.Ref[int]("X")), grammar.Choice(1, grammar.Term(2), grammar.Ref[int]("X"))).
		Rule("X", grammar.Seq(10), grammar.Seq(20))
	genny.Seed(2)
	p := grammar.NewPlayer(g, "S", 3)
	if !p.Continuous() || p.Done() {
		t.Error("state before start")
	}
	seen := map[int]bool{}
	for range 200 {
		v := p.Generate()
		seen[v] = true
		if len(p.Phrase()) != 2 {
			t.Fatalf("phrase %v", p.Phrase())
		}
	}
	if !seen[1] && !seen[2] || !seen[10] || !seen[20] {
		t.Errorf("new derivations expected: %v", seen)
	}
}

func TestPlayerNewExpansionPerCycle(t *testing.T) {
	g := grammar.New[int]().Rule("S", grammar.Choice(1, grammar.Term(1), grammar.Term(1)), grammar.Choice(1, grammar.Term(2), grammar.Term(3)))
	genny.Seed(5)
	p := grammar.NewPlayer(g, "S", 1)
	var cycles [][]int
	for range 30 {
		cycles = append(cycles, []int{p.Generate(), p.Generate(), p.Generate()})
	}
	different := false
	for _, c := range cycles[1:] {
		if !slices.Equal(c, cycles[0]) {
			different = true
		}
	}
	if !different {
		t.Error("every cycle is the same")
	}
}

func TestPlayerEmptyAndReset(t *testing.T) {
	e := grammar.New[int]().Rule("S", grammar.Choice[int](1))
	p := grammar.NewPlayer(e, "S", 2)
	if v := p.Generate(); v != 0 || !p.Done() {
		t.Errorf("empty expansion: %d, done %v", v, p.Done())
	}
	p.Reset()
	if p.Done() {
		t.Error("reset")
	}
}
