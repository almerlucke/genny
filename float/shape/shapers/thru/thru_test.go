package thru

import "testing"

func TestThru(t *testing.T) {
	for _, x := range []float64{0, 1.5, -3} {
		if got := New().Shape(x); got != x {
			t.Errorf("Shape(%v) = %v", x, got)
		}
	}
}
