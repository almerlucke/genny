package mirror

import (
	"math"
	"testing"
)

func TestMirror(t *testing.T) {
	tests := []struct {
		name        string
		bottom, top float64
		x, want     float64
	}{
		{"inside", 0, 1, 0.5, 0.5},
		{"on the bottom", 0, 1, 0, 0},
		{"on the top", 0, 1, 1, 1},
		{"above", 0, 1, 1.25, 0.75},
		{"below", 0, 1, -0.25, 0.25},
		{"twice above", 0, 1, 2.25, 0.25},
		{"three times", 0, 1, 3.5, 0.5},
		{"far below", 0, 1, -1.5, 0.5},
		{"on a multiple", 0, 1, 2, 0},
		{"symmetric range", -1, 1, 1.5, 0.5},
		{"offset range", 2, 4, 4.5, 3.5},
		{"offset range below", 2, 4, 1, 3},
		{"swapped bounds", 1, 0, 1.25, 0.75},
		{"no width", 2, 2, 5, 2},
		{"no width inside", 2, 2, 2, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.bottom, tt.top).Shape(tt.x); math.Abs(got-tt.want) > 1e-12 {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}

func TestMirrorStaysInRange(t *testing.T) {
	m := New(-0.3, 0.9)
	for x := -50.0; x <= 50; x += 0.37 {
		if v := m.Shape(x); v < -0.3-1e-12 || v > 0.9+1e-12 {
			t.Fatalf("Shape(%v) = %v", x, v)
		}
	}
	// must terminate for non finite input
	m.Shape(math.Inf(1))
	m.Shape(math.NaN())
}
