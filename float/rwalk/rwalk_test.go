package rwalk_test

import (
	"math"
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/float/rwalk"
)

func TestStaysInRange(t *testing.T) {
	genny.Seed(1)
	for _, mode := range []rwalk.Mode{rwalk.Clip, rwalk.Bounce, rwalk.Jump} {
		for _, pct := range []float64{0.05, 0.5, 3.7} {
			w := rwalk.New(0.5, 0, 1, pct, mode)
			if !w.Continuous() || w.Done() {
				t.Fatal("continuous, never done")
			}
			for i := 0; i < 3000; i++ {
				v := w.Generate()
				if v < 0 || v > 1 || math.IsNaN(v) {
					t.Fatalf("mode %d pct %v: %v out of range", mode, pct, v)
				}
			}
		}
	}
}

func TestStepSizeIsLimited(t *testing.T) {
	genny.Seed(2)
	w := rwalk.New(50, 0, 100, 1, rwalk.Clip)
	prev := 50.0
	moved := false
	for i := 0; i < 1000; i++ {
		v := w.Generate()
		if math.Abs(v-prev) > 1 {
			t.Fatalf("step %v -> %v is more than maxPct", prev, v)
		}
		moved = moved || v != prev
		prev = v
	}
	if !moved {
		t.Fatal("the walk never moved")
	}
}

func TestNoStepsStaysAtOrigin(t *testing.T) {
	w := rwalk.New(0.3, 0, 1, 0, rwalk.Bounce)
	for i := 0; i < 10; i++ {
		if w.Generate() != 0.3 {
			t.Fatal("maxPct 0 means no movement")
		}
	}
}

func TestClipSticksToTheEdges(t *testing.T) {
	genny.Seed(3)
	w := rwalk.New(0, 0, 1, 10, rwalk.Clip)
	edges := 0
	for i := 0; i < 200; i++ {
		if v := w.Generate(); v == 0 || v == 1 {
			edges++
		}
	}
	if edges < 150 {
		t.Fatalf("huge steps with clip should mostly sit on an edge, %d/200", edges)
	}
}

func TestResetAndSeed(t *testing.T) {
	run := func() []float64 {
		genny.Seed(5)
		w := rwalk.New(0.5, 0, 1, 0.3, rwalk.Bounce)
		out := make([]float64, 30)
		for i := range out {
			out[i] = w.Generate()
		}
		return out
	}
	a, b := run(), run()
	for i := range a {
		if a[i] != b[i] {
			t.Fatal("same seed must repeat")
		}
	}

	genny.Seed(6)
	w := rwalk.New(0.5, 0, 1, 0.3, rwalk.Clip)
	w.Generate()
	w.Reset()
	genny.Seed(6)
	first := rwalk.New(0.5, 0, 1, 0.3, rwalk.Clip).Generate()
	genny.Seed(6)
	if got := w.Generate(); got != first {
		t.Fatalf("after Reset the walk starts at the origin again: %v != %v", got, first)
	}
}

// a range without room, or given the wrong way around, must not hang or leave the range
func TestDegenerateRanges(t *testing.T) {
	genny.Seed(7)
	for _, mode := range []rwalk.Mode{rwalk.Clip, rwalk.Bounce, rwalk.Jump} {
		w := rwalk.New(2, 2, 2, 0.5, mode)
		for i := 0; i < 20; i++ {
			if v := w.Generate(); v != 2 {
				t.Fatalf("mode %d: got %v want 2", mode, v)
			}
		}
		r := rwalk.New(0.5, 1, 0, 0.4, mode) // min and max swapped
		for i := 0; i < 500; i++ {
			if v := r.Generate(); v < 0 || v > 1 {
				t.Fatalf("mode %d: %v out of range", mode, v)
			}
		}
	}
}
