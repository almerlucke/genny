package wavefold

import (
	"math"
	"testing"

	"github.com/almerlucke/genny/float/shape"
)

var (
	_ shape.Shaper = (*Sine)(nil)
	_ shape.Shaper = (*Triangle)(nil)
)

func TestKnownValues(t *testing.T) {
	tests := []struct {
		name string
		s    shape.Shaper
		x    float64
		want float64
	}{
		{"sine 1 at 1", NewSine(1), 1, 1},
		{"sine 1 at 0.5", NewSine(1), 0.5, math.Sqrt2 / 2},
		{"sine 2 at 1", NewSine(2), 1, 0},
		{"sine 3 at 1", NewSine(3), 1, -1},
		{"sine 0", NewSine(0), 0.7, 0},
		{"tri 1 at 0.5", NewTriangle(1), 0.5, 0.5},
		{"tri 1 at 1", NewTriangle(1), 1, 1},
		{"tri 1 at 1.5", NewTriangle(1), 1.5, 0.5},
		{"tri 1 at 2", NewTriangle(1), 2, 0},
		{"tri 1 at 3", NewTriangle(1), 3, -1},
		{"tri 1 at 4", NewTriangle(1), 4, 0},
		{"tri 1 at -1.5", NewTriangle(1), -1.5, -0.5},
		{"tri 2 at 0.75", NewTriangle(2), 0.75, 0.5},
		{"tri 0.5 at 1", NewTriangle(0.5), 1, 0.5},
	}
	for _, tt := range tests {
		if got := tt.s.Shape(tt.x); math.Abs(got-tt.want) > 1e-12 {
			t.Errorf("%s: got %v want %v", tt.name, got, tt.want)
		}
	}
}

func TestPropertiesAndRobustness(t *testing.T) {
	for _, gain := range []float64{0, 0.3, 1, 2.5, 10, 1e6} {
		for name, s := range map[string]shape.Shaper{"sine": NewSine(gain), "tri": NewTriangle(gain)} {
			for x := -3.0; x <= 3; x += 0.013 {
				y := s.Shape(x)
				if math.IsNaN(y) || math.Abs(y) > 1+1e-12 {
					t.Fatalf("%s gain %v: Shape(%v) = %v", name, gain, x, y)
				}
				if math.Abs(s.Shape(-x)+y) > 1e-9 {
					t.Fatalf("%s gain %v: not odd at %v", name, gain, x)
				}
			}
			for _, x := range []float64{1e300, -1e300, math.MaxFloat64, math.Inf(1), math.SmallestNonzeroFloat64} {
				if y := s.Shape(x); math.IsNaN(y) || math.IsInf(y, 0) || math.Abs(y) > 1 {
					t.Fatalf("%s gain %v: Shape(%v) = %v", name, gain, x, y)
				}
			}
		}
	}
}

func TestMonotoneWithoutFold(t *testing.T) {
	for _, s := range []shape.Shaper{NewSine(1), NewTriangle(1)} {
		prev := math.Inf(-1)
		for x := -1.0; x <= 1; x += 0.01 {
			if y := s.Shape(x); y < prev {
				t.Fatalf("%T not monotone at %v", s, x)
			} else {
				prev = y
			}
		}
	}
}

func TestPanics(t *testing.T) {
	for name, f := range map[string]func(){
		"sine neg": func() { NewSine(-1) },
		"sine nan": func() { NewSine(math.NaN()) },
		"tri neg":  func() { NewTriangle(-0.1) },
		"tri inf":  func() { NewTriangle(math.Inf(1)) },
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: no panic", name)
				}
			}()
			f()
		}()
	}
}
