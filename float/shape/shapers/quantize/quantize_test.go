package quantize

import (
	"math"
	"testing"
)

func TestQuantizer(t *testing.T) {
	q := New(0.25)
	tests := []struct{ x, want float64 }{
		{0, 0}, {0.3, 0.25}, {0.4, 0.5}, {0.125, 0.25}, {-0.125, -0.25}, {-0.3, -0.25}, {1, 1}, {0.1, 0},
	}
	for _, tt := range tests {
		if got := q.Shape(tt.x); math.Abs(got-tt.want) > 1e-12 {
			t.Errorf("Shape(%v) = %v want %v", tt.x, got, tt.want)
		}
	}
	if q.Quantum() != 0.25 {
		t.Error("Quantum")
	}
	q.SetQuantum(1)
	if q.Quantum() != 1 || q.Shape(2.6) != 3 {
		t.Error("SetQuantum")
	}
}
