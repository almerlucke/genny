package supersaw

import (
	"math"
	"testing"
)

func expected(x, a1, m1, m2 float64) float64 {
	y := x * a1
	s := math.Mod(y, m1) + math.Mod(y, m2)
	return math.Sin(s)*2 - 1 // the sine takes radians, not a phase
}

func TestSuperSaw(t *testing.T) {
	saw := New(1.5, 0.25, 0.88)
	for _, x := range []float64{0, 0.2, 0.5, 0.77, 1} {
		if got, want := saw.Shape(x), expected(x, 1.5, 0.25, 0.88); math.Abs(got-want) > 1e-12 {
			t.Errorf("Shape(%v) = %v want %v", x, got, want)
		}
	}
	if got := saw.Shape(0); got != -1 {
		t.Errorf("Shape(0) = %v", got)
	}
	saw.SetA1(1)
	if got, want := saw.Shape(0.2), expected(0.2, 1, 0.25, 0.88); math.Abs(got-want) > 1e-12 {
		t.Errorf("SetA1: got %v want %v", got, want)
	}
	saw.SetM1(1)
	saw.SetM2(0.5)
	if got, want := saw.Shape(0.7), expected(0.7, 1, 1, 0.5); math.Abs(got-want) > 1e-12 {
		t.Errorf("SetM1/SetM2: got %v want %v", got, want)
	}
}
