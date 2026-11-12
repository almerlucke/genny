package rectify

import (
	"math"
	"testing"

	"github.com/almerlucke/genny/float/shape"
)

var _ shape.Shaper = (*Rectifier)(nil)

func TestKnownValues(t *testing.T) {
	tests := []struct {
		name string
		r    *Rectifier
		x    float64
		want float64
	}{
		{"half pos", New(HalfWave), 0.4, 0.4},
		{"half neg", New(HalfWave), -0.4, 0},
		{"full neg", New(FullWave), -0.4, 0.4},
		{"full pos", New(FullWave), 0.9, 0.9},
		{"diode below", NewDiode(0.2, 0.1), 0.05, 0},
		{"diode knee start", NewDiode(0.2, 0.1), 0.1, 0},
		{"diode knee mid", NewDiode(0.2, 0.1), 0.2, 0.025},
		{"diode knee end", NewDiode(0.2, 0.1), 0.3, 0.1},
		{"diode above", NewDiode(0.2, 0.1), 0.8, 0.6},
		{"diode hard", NewDiode(0.5, 0), 0.75, 0.25},
		{"diode hard below", NewDiode(0.5, 0), 0.25, 0},
		{"diode plain", NewDiode(0, 0), -1, 0},
	}
	for _, tt := range tests {
		if got := tt.r.Shape(tt.x); math.Abs(got-tt.want) > 1e-12 {
			t.Errorf("%s: got %v want %v", tt.name, got, tt.want)
		}
	}
}

func TestProperties(t *testing.T) {
	for name, r := range map[string]*Rectifier{
		"half": New(HalfWave), "full": New(FullWave), "diode": NewDiode(0.3, 0.2), "hard": NewDiode(0.3, 0),
	} {
		prev := math.Inf(-1)
		for x := 0.0; x <= 3; x += 0.01 { // non decreasing for x >= 0
			y := r.Shape(x)
			if y < prev-1e-15 || y < 0 || y > x+1e-12 {
				t.Fatalf("%s: bad value %v at %v", name, y, x)
			}
			prev = y
		}
		for x := -3.0; x <= 3; x += 0.01 { // never negative, even for full wave
			if r.Shape(x) < 0 {
				t.Fatalf("%s: negative at %v", name, x)
			}
		}
		for _, x := range []float64{1e300, -1e300, math.MaxFloat64, math.Inf(1)} {
			if y := r.Shape(x); math.IsNaN(y) {
				t.Fatalf("%s: NaN for %v", name, x)
			}
		}
	}
	// the diode is continuous across the knee
	d := NewDiode(0.4, 0.2)
	for _, edge := range []float64{0.2, 0.6} {
		if math.Abs(d.Shape(edge-1e-9)-d.Shape(edge+1e-9)) > 1e-8 {
			t.Fatalf("discontinuous at %v", edge)
		}
	}
	// full wave is even
	f := New(FullWave)
	for x := 0.0; x < 2; x += 0.1 {
		if f.Shape(x) != f.Shape(-x) {
			t.Fatal("full wave not even")
		}
	}
}

func TestPanics(t *testing.T) {
	for name, f := range map[string]func(){
		"new diode":     func() { New(Diode) },
		"bad mode":      func() { New(Mode(7)) },
		"neg threshold": func() { NewDiode(-0.1, 0) },
		"neg knee":      func() { NewDiode(0, -1) },
		"nan":           func() { NewDiode(math.NaN(), 0) },
		"inf knee":      func() { NewDiode(0, math.Inf(1)) },
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
