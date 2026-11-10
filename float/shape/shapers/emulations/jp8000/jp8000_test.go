package jp8000

import (
	"math"
	"testing"
)

func TestTriMod(t *testing.T) {
	tm := NewTriMod(0.5)
	tests := []struct{ x, want float64 }{{0, 0}, {0.1, 0.6}, {0.25, 0}, {0.4, -0.6}, {0.5, 0}}
	for _, tt := range tests {
		if got := tm.Shape(tt.x); math.Abs(got-tt.want) > 1e-12 {
			t.Errorf("m=0.5 Shape(%v) = %v want %v", tt.x, got, tt.want)
		}
	}
	tm.SetMod(1)
	if got := tm.Shape(0.1); math.Abs(got+0.8) > 1e-12 {
		t.Errorf("m=1 Shape(0.1) = %v want -0.8", got)
	}
	tm.SetMod(0.5)
	if got := tm.Shape(0.1); math.Abs(got-0.6) > 1e-12 {
		t.Errorf("SetMod back: got %v", got)
	}
}
