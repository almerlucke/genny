package function_test

import (
	"fmt"
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/function"
)

func TestFunction(t *testing.T) {
	n := 0
	f := function.New(func() int { n++; return n })
	if f.Generate() != 1 || f.Generate() != 2 {
		t.Fatal("calls the function for every value")
	}
	if !f.Continuous() || f.Done() {
		t.Fatal("continuous, never done")
	}
	f.Reset()
	if f.Generate() != 3 {
		t.Fatal("Reset does nothing")
	}
}

func TestFunctionWithContext(t *testing.T) {
	f := function.NewWithContext("ab", func(ctx any) int { return len(ctx.(string)) })
	if f.Generate() != 2 {
		t.Fatal("context passed")
	}
	// a nil context still calls the context function
	g := function.NewWithContext(nil, func(ctx any) string { return fmt.Sprint(ctx) })
	if g.Generate() != "<nil>" {
		t.Fatal("nil context")
	}
}

func TestNewRandom(t *testing.T) {
	genny.Seed(5)
	r := function.NewRandom(2, 4)
	var first [10]float64
	for i := range first {
		v := r.Generate()
		if v < 2 || v >= 4 {
			t.Fatalf("%f out of range", v)
		}
		first[i] = v
	}
	genny.Seed(5)
	r = function.NewRandom(2, 4)
	for i := range first {
		if r.Generate() != first[i] {
			t.Fatal("not repeatable with the same seed")
		}
	}
}
