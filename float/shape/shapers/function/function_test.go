package function

import (
	"math"
	"testing"
)

func TestFunctions(t *testing.T) {
	tests := []struct {
		name string
		f    *Function
		x    float64
		want float64
	}{
		{"mod 2.5", NewMod(2), 2.5, 0.5},
		{"mod sign follows x", NewMod(2), -2.5, -0.5},
		{"mod1", NewMod1(), 1.25, 0.25},
		{"mod1 whole", NewMod1(), 3, 0},
		{"abs neg", NewAbs(), -3, 3},
		{"abs pos", NewAbs(), 3, 3},
		{"sin 0", NewSin(), 0, 0},
		{"sin quarter", NewSin(), 0.25, 1},
		{"sin half", NewSin(), 0.5, 0},
		{"sin three quarters", NewSin(), 0.75, -1},
		{"tri 0", NewTri(), 0, 0},
		{"tri quarter", NewTri(), 0.25, 0.5},
		{"tri half", NewTri(), 0.5, 1},
		{"tri three quarters", NewTri(), 0.75, 0.5},
		{"tri 1", NewTri(), 1, 0},
		{"custom", New(func(x float64) float64 { return x * x }), 3, 9},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.Shape(tt.x); math.Abs(got-tt.want) > 1e-12 {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}
