package hardsync

import (
	"math"
	"testing"
)

func TestHardSync(t *testing.T) {
	hs := New(2)
	tests := []struct{ x, want float64 }{
		{0, -1}, {0.3, 0.2}, {0.5, -1}, {0.75, 0}, {0.25, 0},
	}
	for _, tt := range tests {
		if got := hs.Shape(tt.x); math.Abs(got-tt.want) > 1e-12 {
			t.Errorf("Shape(%v) = %v want %v", tt.x, got, tt.want)
		}
	}
	hs.SetA1(3)
	if got := hs.Shape(0.5); math.Abs(got) > 1e-12 { // 1.5 -> .5 -> 0
		t.Errorf("after SetA1: got %v", got)
	}
}
