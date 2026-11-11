package genny_test

import (
	"slices"
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/constrain"
	"github.com/almerlucke/genny/function"
	"github.com/almerlucke/genny/grammar"
	"github.com/almerlucke/genny/row"
	"github.com/almerlucke/genny/sequence"
)

func ints(g genny.Generator[int], n int) (out []float64) {
	for range n {
		out = append(out, float64(g.Generate()))
	}
	return
}

func wild() genny.Generator[int] { return function.New(func() int { return 40 + genny.Intn(40) }) }

var randomConstraints = map[string]func() []float64{
	"constrain.AllowInterval": func() []float64 {
		return ints(constrain.AllowInterval(wild(), []int{1, 2, 3, 5}), 40)
	},
	"constrain.BlockInterval": func() []float64 {
		return ints(constrain.BlockInterval(wild(), []int{0, 6, 11}), 40)
	},
	"constrain.RemoveRepetitions": func() []float64 {
		return ints(constrain.RemoveRepetitions(function.New(func() int { return genny.Intn(3) })), 40)
	},
	"constrain.Duplicates": func() []float64 {
		return ints(constrain.Duplicates(function.New(func() int { return genny.Intn(5) }), 5), 40)
	},
	"constrain.Shotgun": func() []float64 {
		return ints(constrain.Shotgun(sequence.NewLoop(1, 2, 3, 4, 5, 6), 0.5), 40)
	},
	"constrain.RandomIntervals": func() []float64 {
		return ints(constrain.RandomIntervals(48, 72, 1, 2, 5, 7), 40)
	},
	"grammar.Expand": func() (out []float64) {
		g := grammar.New[int]().
			Rule("S", grammar.Choice(1, grammar.Ref[int]("A"), grammar.Ref[int]("A")), grammar.Choice(1, grammar.Term(9))).
			Rule("A", grammar.Seq(1, 2), grammar.Seq(3), grammar.Choice(1, grammar.Term(4), grammar.Ref[int]("A")))
		for range 6 {
			for _, v := range g.Expand("S", 5) {
				out = append(out, float64(v))
			}
		}
		return
	},
	"grammar.Player": func() []float64 {
		g := grammar.New[int]().Rule("S", grammar.Seq(1, 2, 3), grammar.Seq(4, 5), grammar.Seq(6))
		return ints(grammar.NewPlayer(g, "S", 2), 40)
	},
	"row.Reader shuffled": func() (out []float64) {
		r := row.NewReader(row.Permutations([]int{0, 1, 2, 3}, 0, 0), row.ReadShuffled)
		for !r.Done() {
			for _, v := range r.Generate() {
				out = append(out, float64(v))
			}
		}
		return
	},
	"row.Reader random": func() (out []float64) {
		r := row.NewReader(row.AllIntervals(8, 0), row.ReadRandom)
		for range 10 {
			for _, v := range r.Generate() {
				out = append(out, float64(v))
			}
		}
		return
	},
}

func TestConstraintsAreSeeded(t *testing.T) {
	for name, f := range randomConstraints {
		genny.Seed(31)
		a := f()
		genny.Seed(31)
		b := f()
		genny.Seed(32)
		c := f()
		if !slices.Equal(a, b) {
			t.Errorf("%s: the same seed gave other values", name)
		}
		if slices.Equal(a, c) {
			t.Errorf("%s: another seed gave the same values", name)
		}
	}
}
