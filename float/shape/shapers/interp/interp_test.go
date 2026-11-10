package interp

import (
	"math"
	"testing"

	"github.com/almerlucke/genny/float/shape/shapers/mult"
	"github.com/almerlucke/genny/float/shape/shapers/thru"
)

func TestInterpolate(t *testing.T) {
	ip := New(thru.New(), mult.New(2), mult.New(4))
	tests := []struct{ index, want float64 }{
		{0, 1}, {0.25, 1.5}, {0.5, 2}, {0.75, 3}, {1, 4},
		{-1, 1}, // clamped
		{2, 4},  // clamped
	}
	for _, tt := range tests {
		ip.Index = tt.index
		if got := ip.Shape(1); math.Abs(got-tt.want) > 1e-12 {
			t.Errorf("index %v: got %v want %v", tt.index, got, tt.want)
		}
	}
	ip.Index = 0.25
	if got := ip.Shape(3); math.Abs(got-4.5) > 1e-12 {
		t.Errorf("input is scaled by the shapers: got %v", got)
	}
}

func TestInterpolateSingleShaper(t *testing.T) {
	ip := New(mult.New(3))
	for _, index := range []float64{0, 0.5, 1} {
		ip.Index = index
		if got := ip.Shape(2); got != 6 {
			t.Fatalf("index %v: got %v", index, got)
		}
	}
}
