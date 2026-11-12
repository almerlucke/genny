package deadzone

import (
	"math"
	"testing"

	"github.com/almerlucke/genny/float/shape"
)

var (
	_ shape.Shaper = (*DeadZone)(nil)
	_ shape.Shaper = (*Expander)(nil)
)

func TestKnownValues(t *testing.T) {
	tests := []struct {
		name string
		s    shape.Shaper
		x    float64
		want float64
	}{
		{"jump inside", NewDeadZone(0.2, Jump), 0.1, 0},
		{"jump edge", NewDeadZone(0.2, Jump), -0.2, 0},
		{"jump outside", NewDeadZone(0.2, Jump), 0.5, 0.5},
		{"jump outside neg", NewDeadZone(0.2, Jump), -0.5, -0.5},
		{"cont inside", NewDeadZone(0.2, Continuous), -0.15, 0},
		{"cont outside", NewDeadZone(0.2, Continuous), 0.5, 0.3},
		{"cont outside neg", NewDeadZone(0.2, Continuous), -0.5, -0.3},
		{"zero width", NewDeadZone(0, Continuous), 0.7, 0.7},
		{"exp above", NewExpander(0.5, 2), 0.8, 0.8},
		{"exp at", NewExpander(0.5, 2), 0.5, 0.5},
		{"exp below", NewExpander(0.5, 2), 0.25, 0.125},
		{"exp below neg", NewExpander(0.5, 2), -0.25, -0.125},
		{"exp ratio 1", NewExpander(0.5, 1), 0.3, 0.3},
		{"exp zero", NewExpander(0.5, 3), 0, 0},
	}
	for _, tt := range tests {
		if got := tt.s.Shape(tt.x); math.Abs(got-tt.want) > 1e-12 {
			t.Errorf("%s: got %v want %v", tt.name, got, tt.want)
		}
	}
}

func TestProperties(t *testing.T) {
	shapers := map[string]shape.Shaper{
		"jump": NewDeadZone(0.3, Jump), "cont": NewDeadZone(0.3, Continuous),
		"exp2": NewExpander(0.4, 2), "exp8": NewExpander(0.4, 8),
	}
	for name, s := range shapers {
		prev := math.Inf(-1)
		for x := -2.0; x <= 2; x += 0.005 {
			y := s.Shape(x)
			if math.IsNaN(y) || math.Abs(y) > math.Abs(x)+1e-12 || y*x < 0 {
				t.Fatalf("%s: bad value %v at %v", name, y, x)
			}
			if y < prev || math.Abs(s.Shape(-x)+y) > 1e-12 {
				t.Fatalf("%s: not monotone/odd at %v", name, x)
			}
			prev = y
		}
		for _, x := range []float64{1e300, -1e300, math.Inf(1), math.Inf(-1)} {
			if y := s.Shape(x); math.IsNaN(y) {
				t.Fatalf("%s: Shape(%v) = NaN", name, x)
			}
		}
	}
	// continuity at the edge
	c, e := NewDeadZone(0.3, Continuous), NewExpander(0.4, 3)
	if math.Abs(c.Shape(0.3+1e-9)) > 1e-8 || math.Abs(e.Shape(0.4-1e-9)-0.4) > 1e-8 {
		t.Fatal("discontinuous")
	}
}

func TestPanics(t *testing.T) {
	for name, f := range map[string]func(){
		"neg width": func() { NewDeadZone(-1, Jump) },
		"nan width": func() { NewDeadZone(math.NaN(), Jump) },
		"bad mode":  func() { NewDeadZone(1, Mode(4)) },
		"zero thr":  func() { NewExpander(0, 2) },
		"inf thr":   func() { NewExpander(math.Inf(1), 2) },
		"low ratio": func() { NewExpander(0.5, 0.5) },
		"nan ratio": func() { NewExpander(0.5, math.NaN()) },
		"inf ratio": func() { NewExpander(0.5, math.Inf(1)) },
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
