package automaton

import (
	"testing"

	"github.com/almerlucke/genny"
)

func row(r []bool) string {
	b := make([]byte, len(r))
	for i, v := range r {
		b[i] = '.'
		if v {
			b[i] = '#'
		}
	}
	return string(b)
}

func TestRule30(t *testing.T) {
	a := New(30, 9)
	for i, want := range []string{"....#....", "...###...", "..##..#..", ".##.####."} {
		if got := row(a.Generate()); got != want {
			t.Errorf("gen %d: %s, want %s", i, got, want)
		}
	}
	a.Reset()
	if got := row(a.Generate()); got != "....#...." {
		t.Errorf("reset: %s", got)
	}
}

func TestRule110(t *testing.T) {
	// rule 110 grows to the left
	a := NewWithCells(110, []bool{false, false, false, false, false, false, true})
	for i, want := range []string{"......#", ".....##", "....###", "...##.#"} {
		if got := row(a.Generate()); got != want {
			t.Errorf("gen %d: %s, want %s", i, got, want)
		}
	}
}

func TestWrapAround(t *testing.T) {
	// rule 2 moves a cell one step left, from the left edge to the right edge
	a := NewWithCells(2, []bool{true, false, false, false})
	a.Generate()
	if got := row(a.Generate()); got != "...#" {
		t.Errorf("got %s", got)
	}
}

func TestGateReadsRows(t *testing.T) {
	g := NewGate(New(30, 5))
	var got string
	for range 10 {
		if g.Generate() {
			got += "#"
		} else {
			got += "."
		}
	}
	if got != "..#...###." {
		t.Errorf("got %s", got)
	}
	g.Reset()
	if g.Generate() || g.Generate() || !g.Generate() {
		t.Error("reset")
	}
}

func TestRandomStartIsSeeded(t *testing.T) {
	genny.Seed(2)
	a := row(NewRandom(90, 32).Generate())
	genny.Seed(2)
	b := row(NewRandom(90, 32).Generate())
	if a != b {
		t.Error("not repeatable")
	}
}
