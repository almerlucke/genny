package parallel

import (
	"testing"

	"github.com/almerlucke/genny/float/shape/shapers/add"
	"github.com/almerlucke/genny/float/shape/shapers/mult"
)

func TestParallelSumsByDefault(t *testing.T) {
	p := New(mult.New(2), add.New(1))
	if got := p.Shape(3); got != 10 { // 6 + 4
		t.Fatalf("got %v", got)
	}
	if got := New().Shape(3); got != 0 {
		t.Fatalf("empty: got %v", got)
	}
}

func TestParallelWithFunction(t *testing.T) {
	product := func(x, v float64) float64 { return x * v }
	p := NewWithFunction(1, product, mult.New(2), mult.New(3))
	if got := p.Shape(2); got != 24 { // 1 * 4 * 6
		t.Fatalf("got %v", got)
	}
	// the function gets the shaper output first and the accumulated value second
	first := NewWithFunction(0, func(x, v float64) float64 { return x - v }, mult.New(10), mult.New(1))
	if got := first.Shape(2); got != -18 { // (20-0) -> (2-20)
		t.Fatalf("got %v", got)
	}
	p.Start = 2
	if got := p.Shape(2); got != 48 {
		t.Fatalf("changed start: got %v", got)
	}
	if got := NewWithFunction(7, product).Shape(2); got != 7 {
		t.Fatalf("no shapers gives the start value, got %v", got)
	}
}
