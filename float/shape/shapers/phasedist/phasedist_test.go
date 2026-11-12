package phasedist

import (
	"math"
	"testing"

	"github.com/almerlucke/genny/float/shape"
)

var _ shape.Shaper = (*Shaper)(nil)

func TestSineAtZero(t *testing.T) {
	shapers := map[string]*Shaper{
		"saw": NewSaw(0), "square": NewSquare(0), "pulse": NewPulse(0), "resonant": NewResonant(0, 5),
	}
	for name, s := range shapers {
		for i := 0; i <= 40; i++ {
			p := float64(i) / 40
			if got, want := s.Shape(p), math.Cos(2*math.Pi*p); math.Abs(got-want) > 1e-12 {
				t.Errorf("%s d=0 Shape(%v) = %v want %v", name, p, got, want)
			}
		}
	}
}

func TestKnownValues(t *testing.T) {
	tests := []struct {
		name string
		s    *Shaper
		x    float64
		want float64
	}{
		{"saw start", NewSaw(1), 0, 1},
		{"saw break", NewSaw(1), 0.005, -1},
		{"saw late", NewSaw(1), 0.9995, 1},
		{"saw wrap", NewSaw(1), 1.005, -1},
		{"saw negative wrap", NewSaw(1), -0.995, -1},
		{"square low", NewSquare(1), 0.25, -1},
		{"square high", NewSquare(1), 0.75, 1},
		{"square start", NewSquare(1), 0, 1},
		{"square width low", NewSquareWidth(1, 0.25), 0.2, -1},
		{"square width high", NewSquareWidth(1, 0.25), 0.3, 1},
		{"pulse inside", NewPulse(1), 0.005, -1},
		{"pulse rest", NewPulse(1), 0.5, 1},
		{"resonant peak", NewResonant(1, 4), 0, 1},
		{"resonant window end", NewResonant(1, 4), 0.99999, 0},
		{"resonant half", NewResonant(1, 4), 0.5, 0.5},
	}
	for _, tt := range tests {
		if got := tt.s.Shape(tt.x); math.Abs(got-tt.want) > 1e-3 {
			t.Errorf("%s: Shape(%v) = %v want %v", tt.name, tt.x, got, tt.want)
		}
	}
}

func TestWarpMonotonic(t *testing.T) {
	for _, s := range []*Shaper{NewSaw(0.7), NewSquare(0.7), NewSquareWidth(0.3, 0.3), NewPulse(0.7)} {
		prev := -1.0
		for i := 0; i < 1000; i++ {
			g := s.Warp(float64(i) / 1000)
			if g < prev-1e-12 || g < 0 || g > 1 {
				t.Fatalf("warp not monotonic in [0,1]: %v after %v", g, prev)
			}
			prev = g
		}
	}
}

func TestRangeAndFinite(t *testing.T) {
	inputs := []float64{0, 0.3, 0.999999, 1, -0.2, 7.77, -1e12, 1e15, math.NaN(), math.Inf(1), math.Inf(-1)}
	for _, d := range []float64{0, 0.1, 0.5, 0.99, 1} {
		for _, s := range []*Shaper{NewSaw(d), NewSquare(d), NewSquareWidth(d, 0.2), NewPulse(d), NewResonant(d, 7.5)} {
			for _, x := range inputs {
				y := s.Shape(x)
				if math.IsNaN(y) || math.IsInf(y, 0) || y < -1-1e-12 || y > 1+1e-12 {
					t.Errorf("kind %d d=%v Shape(%v) = %v", s.kind, d, x, y)
				}
			}
		}
	}
}

func TestPanics(t *testing.T) {
	tests := map[string]func(){
		"saw d<0":        func() { NewSaw(-0.1) },
		"saw d>1":        func() { NewSaw(1.1) },
		"square NaN":     func() { NewSquare(math.NaN()) },
		"pulse d>1":      func() { NewPulse(2) },
		"width 0":        func() { NewSquareWidth(0.5, 0) },
		"width 1":        func() { NewSquareWidth(0.5, 1) },
		"resonant d":     func() { NewResonant(2, 3) },
		"resonant h<1":   func() { NewResonant(0.5, 0.5) },
		"resonant h inf": func() { NewResonant(0.5, math.Inf(1)) },
	}
	for name, f := range tests {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: expected panic", name)
				}
			}()
			f()
		}()
	}
}
