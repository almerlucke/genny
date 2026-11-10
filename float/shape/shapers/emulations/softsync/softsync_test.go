package softsync

import (
	"math"
	"testing"
)

func TestTriangle(t *testing.T) {
	tr := NewTriangle(1.25)
	tests := []struct{ x, want float64 }{
		{0.5, -1}, {0, 0}, {1, 0}, {0.9, -1}, {0.75, 0.5},
	}
	for _, tt := range tests {
		if got := tr.Shape(tt.x); math.Abs(got-tt.want) > 1e-12 {
			t.Errorf("Shape(%v) = %v want %v", tt.x, got, tt.want)
		}
	}
	tr.SetA1(1)
	if got := tr.Shape(0.75); math.Abs(got-1) > 1e-12 { // .5 -> tri 1 -> 1
		t.Errorf("after SetA1: got %v", got)
	}
}
