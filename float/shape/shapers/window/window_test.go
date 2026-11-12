package window

import (
	"math"
	"testing"

	"github.com/almerlucke/genny/float/shape"
)

var _ shape.Shaper = (*Window)(nil)

func TestKnownValues(t *testing.T) {
	tests := []struct {
		name string
		w    *Window
		x    float64
		want float64
	}{
		{"hann start", Hann(), 0, 0},
		{"hann mid", Hann(), 0.5, 1},
		{"hann quarter", Hann(), 0.25, 0.5},
		{"hamming start", Hamming(), 0, 0.08},
		{"hamming mid", Hamming(), 0.5, 1},
		{"blackman start", Blackman(), 0, 0},
		{"blackman mid", Blackman(), 0.5, 1},
		{"tukey 0 is rect", Tukey(0), 0, 1},
		{"tukey flat", Tukey(0.5), 0.5, 1},
		{"tukey start", Tukey(0.5), 0, 0},
		{"tukey half taper", Tukey(0.5), 0.125, 0.5},
		{"tukey 1 is hann", Tukey(1), 0.25, 0.5},
		{"gaussian mid", Gaussian(0.4), 0.5, 1},
		{"gaussian edge", Gaussian(0.5), 0, math.Exp(-2)},
		{"kaiser mid", Kaiser(8), 0.5, 1},
		{"kaiser 0 is rect", Kaiser(0), 0.1, 1},
		{"kaiser edge", Kaiser(5), 0, 1 / I0(5)},
		{"welch edge", Welch(), 0, 0},
		{"welch quarter", Welch(), 0.25, 0.75},
		{"triangle quarter", Triangle(), 0.25, 0.5},
		{"triangle mid", Triangle(), 0.5, 1},
		{"outside low", Hann(), -0.1, 0},
		{"outside high", Hann(), 1.1, 0},
		{"end of range", Hamming(), 1, 0.08},
	}
	for _, tt := range tests {
		if got := tt.w.Shape(tt.x); math.Abs(got-tt.want) > 1e-12 {
			t.Errorf("%s: Shape(%v) = %v want %v", tt.name, tt.x, got, tt.want)
		}
	}
}

func TestI0(t *testing.T) {
	tests := []struct{ x, want float64 }{
		{0, 1}, {1, 1.2660658777520082}, {-1, 1.2660658777520082}, {5, 27.239871823604442}, {20, 4.355828255955353e7},
	}
	for _, tt := range tests {
		if got := I0(tt.x); math.Abs(got-tt.want) > 1e-12*tt.want {
			t.Errorf("I0(%v) = %v want %v", tt.x, got, tt.want)
		}
	}
}

func TestSymmetricAndBounded(t *testing.T) {
	ws := map[string]*Window{
		"hann": Hann(), "hamming": Hamming(), "blackman": Blackman(), "tukey": Tukey(0.3),
		"gaussian": Gaussian(0.3), "kaiser": Kaiser(9), "kaiser500": Kaiser(500), "welch": Welch(), "triangle": Triangle(),
	}
	for name, w := range ws {
		for i := 0; i <= 200; i++ {
			x := float64(i) / 200
			y := w.Shape(x)
			if math.IsNaN(y) || math.IsInf(y, 0) || y < -1e-12 || y > 1+1e-12 {
				t.Errorf("%s: Shape(%v) = %v out of [0,1]", name, x, y)
			}
			if m := w.Shape(1 - x); math.Abs(y-m) > 1e-9 {
				t.Errorf("%s: not symmetric at %v: %v vs %v", name, x, y, m)
			}
		}
		for _, x := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
			if w.Shape(x) != 0 || w.Periodic().Shape(x) != 0 {
				t.Errorf("%s: non finite input should give 0", name)
			}
		}
	}
}

func TestPeriodic(t *testing.T) {
	w := Hann().Periodic()
	tests := []struct{ x, want float64 }{
		{0, 0}, {1, 0}, {1.5, 1}, {-0.5, 1}, {0.25, 0.5}, {2.25, 0.5},
	}
	for _, tt := range tests {
		if got := w.Shape(tt.x); math.Abs(got-tt.want) > 1e-12 {
			t.Errorf("periodic Shape(%v) = %v want %v", tt.x, got, tt.want)
		}
	}
	// the original stays symmetric (0 outside the range)
	h := Hann()
	_ = h.Periodic()
	if h.Shape(1.5) != 0 {
		t.Error("Periodic must not change the original")
	}
}

func TestPanics(t *testing.T) {
	tests := map[string]func(){
		"tukey -1":     func() { Tukey(-1) },
		"tukey 2":      func() { Tukey(2) },
		"tukey NaN":    func() { Tukey(math.NaN()) },
		"gaussian 0":   func() { Gaussian(0) },
		"gaussian -1":  func() { Gaussian(-1) },
		"gaussian inf": func() { Gaussian(math.Inf(1)) },
		"kaiser -1":    func() { Kaiser(-1) },
		"kaiser 501":   func() { Kaiser(501) },
		"kaiser NaN":   func() { Kaiser(math.NaN()) },
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
