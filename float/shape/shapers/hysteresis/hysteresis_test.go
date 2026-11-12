package hysteresis

import (
	"math"
	"testing"

	"github.com/almerlucke/genny/float/shape"
)

var (
	_ shape.Shaper = (*Schmitt)(nil)
	_ shape.Shaper = (*Play)(nil)
)

func TestSchmitt(t *testing.T) {
	s := NewSchmitt(0.3, 0.7, -1, 1)
	in := []float64{0, 0.5, 0.69, 0.7, 0.5, 0.31, 0.3, 0.5, 0.9, 0.2, math.NaN(), 0.5}
	want := []float64{-1, -1, -1, 1, 1, 1, -1, -1, 1, -1, -1, -1}
	for i, x := range in {
		if got := s.Shape(x); got != want[i] {
			t.Errorf("step %d (x=%v): got %v want %v", i, x, got, want[i])
		}
	}
	s.Shape(0.9)
	if !s.On() {
		t.Error("expected on")
	}
	s.Reset()
	if s.On() || s.Shape(0.5) != -1 {
		t.Error("Reset must switch off")
	}
}

func TestSchmittEqualThresholds(t *testing.T) {
	s := NewSchmitt(0.5, 0.5, 0, 1)
	if s.Shape(0.6) != 1 || s.Shape(0.4) != 0 || s.Shape(0.5) != 1 {
		t.Error("equal thresholds: x==threshold turns on")
	}
}

func TestPlay(t *testing.T) {
	p := NewPlay(1)
	in := []float64{0, 0.3, 0.5, 1, 2, 1.8, 1.5, 1, 0, -1}
	want := []float64{0, 0, 0, 0.5, 1.5, 1.5, 1.5, 1.5, 0.5, -0.5}
	for i, x := range in {
		if got := p.Shape(x); math.Abs(got-want[i]) > 1e-12 {
			t.Errorf("step %d (x=%v): got %v want %v", i, x, got, want[i])
		}
	}
	p.Reset()
	if p.Shape(0.4) != 0 {
		t.Error("Reset")
	}
}

func TestPlayProperties(t *testing.T) {
	// the output stays within half the width of the input and the width 0 play is the identity
	p, id := NewPlay(0.4), NewPlay(0)
	for i := 0; i < 1000; i++ {
		x := math.Sin(float64(i)/17) * 3
		y := p.Shape(x)
		if math.Abs(x-y) > 0.2+1e-12 && i > 0 {
			t.Fatalf("output %v further than w/2 from input %v", y, x)
		}
		if id.Shape(x) != x {
			t.Fatal("width 0 must pass the input")
		}
	}
	if y := p.Shape(math.Inf(1)); math.IsInf(y, 0) || math.IsNaN(y) {
		t.Error("non finite input must be ignored")
	}
}

func TestPanics(t *testing.T) {
	tests := map[string]func(){
		"low>high":  func() { NewSchmitt(1, 0, 0, 1) },
		"NaN":       func() { NewSchmitt(math.NaN(), 0, 0, 1) },
		"inf out":   func() { NewSchmitt(0, 1, 0, math.Inf(1)) },
		"width < 0": func() { NewPlay(-1) },
		"width NaN": func() { NewPlay(math.NaN()) },
		"width inf": func() { NewPlay(math.Inf(1)) },
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
