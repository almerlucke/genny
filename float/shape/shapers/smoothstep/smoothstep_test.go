package smoothstep

import (
	"math"
	"testing"

	"github.com/almerlucke/genny/float/shape"
)

var _ shape.Shaper = (*Step)(nil)

func TestKnownValues(t *testing.T) {
	tests := []struct {
		name string
		s    *Step
		x    float64
		want float64
	}{
		{"smooth mid", New(Smoothstep, 0, 1), 0.5, 0.5},
		{"smooth quarter", New(Smoothstep, 0, 1), 0.25, 0.15625},
		{"smoother mid", New(Smootherstep, 0, 1), 0.5, 0.5},
		{"smoother quarter", New(Smootherstep, 0, 1), 0.25, 0.103515625},
		{"inverse mid", New(Inverse, 0, 1), 0.5, 0.5},
		{"below", New(Smoothstep, 2, 4), 1, 0},
		{"above", New(Smootherstep, 2, 4), 9, 1},
		{"scaled", New(Smoothstep, 2, 4), 2.5, 0.15625},
		{"negative edges", New(Smoothstep, -1, 1), 0, 0.5},
		{"inverse ends", New(Inverse, 0, 1), 1, 1},
	}
	for _, tt := range tests {
		if got := tt.s.Shape(tt.x); math.Abs(got-tt.want) > 1e-12 {
			t.Errorf("%s: got %v want %v", tt.name, got, tt.want)
		}
	}
}

func TestProperties(t *testing.T) {
	for _, k := range []Kind{Smoothstep, Smootherstep, Inverse} {
		s := New(k, -1, 1)
		prev := -1.0
		for x := -2.0; x <= 2; x += 0.005 {
			y := s.Shape(x)
			if y < prev-1e-12 || y < 0 || y > 1 {
				t.Fatalf("kind %v: bad at %v: %v", k, x, y)
			}
			prev = y
			// point symmetry about the middle
			if math.Abs(s.Shape(-x)+y-1) > 1e-9 {
				t.Fatalf("kind %v: not point symmetric at %v", k, x)
			}
		}
		for _, x := range []float64{1e300, -1e300, math.Inf(1), math.Inf(-1), math.NaN()} {
			if y := s.Shape(x); math.IsNaN(y) || y < 0 || y > 1 {
				t.Fatalf("kind %v: Shape(%v) = %v", k, x, y)
			}
		}
	}
	// Inverse really inverts Smoothstep
	f, g := New(Smoothstep, 0, 1), New(Inverse, 0, 1)
	for x := 0.0; x <= 1; x += 0.01 {
		if math.Abs(g.Shape(f.Shape(x))-x) > 1e-9 {
			t.Fatalf("inverse(smoothstep(%v)) = %v", x, g.Shape(f.Shape(x)))
		}
	}
}

func TestPanics(t *testing.T) {
	for name, f := range map[string]func(){
		"equal":    func() { New(Smoothstep, 1, 1) },
		"reversed": func() { New(Smoothstep, 1, 0) },
		"nan":      func() { New(Smoothstep, math.NaN(), 1) },
		"inf":      func() { New(Smoothstep, 0, math.Inf(1)) },
		"kind":     func() { New(Kind(5), 0, 1) },
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
