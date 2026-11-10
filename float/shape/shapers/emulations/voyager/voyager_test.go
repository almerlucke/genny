package voyager

import (
	"math"
	"testing"
)

func TestSawtooth(t *testing.T) {
	s := NewSawtooth()
	tests := []struct{ x, want float64 }{{0, -1}, {1, 1}, {2, -1}, {3, -3}, {-1, -3}}
	for _, tt := range tests {
		if got := s.Shape(tt.x); math.Abs(got-tt.want) > 1e-9 {
			t.Errorf("Shape(%v) = %v want %v", tt.x, got, tt.want)
		}
	}
}
