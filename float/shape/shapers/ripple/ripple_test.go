package ripple

import (
	"math"
	"testing"
)

func TestRipple(t *testing.T) {
	r := New(1)
	tests := []struct{ x, want float64 }{{2.5, 3}, {0.25, 0.5}, {3, 3}, {-0.25, -0.5}}
	for _, tt := range tests {
		if got := r.Shape(tt.x); math.Abs(got-tt.want) > 1e-12 {
			t.Errorf("Shape(%v) = %v want %v", tt.x, got, tt.want)
		}
	}
	r.M = 0.5
	if got := r.Shape(1.25); math.Abs(got-1.5) > 1e-12 {
		t.Errorf("got %v", got)
	}
}
