package chebyshev

import (
	"math"
	"testing"
)

func TestChebyshev(t *testing.T) {
	tests := []struct {
		name string
		h    map[int]float64
		x    float64
		want float64
	}{
		{"T1 is identity", map[int]float64{1: 1}, 0.3, 0.3},
		{"T2 = 2x^2-1 at .5", map[int]float64{2: 1}, 0.5, -0.5},
		{"T2 at 1", map[int]float64{2: 1}, 1, 1},
		{"T2 at 0", map[int]float64{2: 1}, 0, -1},
		{"T3 = 4x^3-3x at .5", map[int]float64{3: 1}, 0.5, -1},
		{"T5 = cos(5 acos x) at .5", map[int]float64{5: 1}, 0.5, 0.5},
		{"weighted mix", map[int]float64{2: 0.5, 3: 0.25}, 0.5, -0.5},
		{"gap in harmonics", map[int]float64{1: 2, 4: 1}, 0.5, 2*0.5 + (8*0.0625 - 8*0.25 + 1)},
		{"harmonic 0 is ignored", map[int]float64{0: 5}, 0.5, 0},
		{"empty", map[int]float64{}, 0.5, 0},
		{"nil", nil, 0.5, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.h).Shape(tt.x); math.Abs(got-tt.want) > 1e-12 {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}

func TestChebyshevOfSineGivesHarmonic(t *testing.T) {
	c := New(map[int]float64{3: 1})
	for _, phase := range []float64{0, 0.1, 0.37, 0.8} {
		x := math.Cos(2 * math.Pi * phase)
		if got, want := c.Shape(x), math.Cos(2*math.Pi*3*phase); math.Abs(got-want) > 1e-9 {
			t.Fatalf("phase %v: got %v want %v", phase, got, want)
		}
	}
}

func TestSetHarmonics(t *testing.T) {
	c := New(map[int]float64{1: 1})
	c.SetHarmonics(map[int]float64{2: 1})
	if got := c.Shape(0.5); math.Abs(got+0.5) > 1e-12 {
		t.Fatalf("got %v", got)
	}
}
