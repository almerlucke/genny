package curve

import (
	"math"
	"testing"

	"github.com/almerlucke/genny/float/shape"
)

var _ shape.Shaper = (*Curve)(nil)

func TestExponent(t *testing.T) {
	tests := []struct{ tension, want float64 }{{0, 1}, {0.5, 0.5}, {-0.5, 2}, {1.0 / 3, 0.6}, {0.99, 1 / 2.98}, {-0.99, 2.98}}
	for _, tt := range tests {
		if got := Exponent(tt.tension); math.Abs(got-tt.want) > 1e-12 {
			t.Errorf("Exponent(%v) = %v want %v", tt.tension, got, tt.want)
		}
	}
}

func TestKnownValues(t *testing.T) {
	tests := []struct {
		name string
		c    *Curve
		x    float64
		want float64
	}{
		{"uni linear", NewUnipolar(0), 0.3, 0.3},
		{"uni pos", NewUnipolar(0.5), 0.25, 0.5},
		{"uni neg", NewUnipolar(-0.5), 0.5, 0.25},
		{"uni clamp low", NewUnipolar(0.3), -2, 0},
		{"uni clamp high", NewUnipolar(0.3), 2, 1},
		{"bi linear", NewBipolar(0), -0.3, -0.3},
		{"bi pos", NewBipolar(0.5), -0.25, -0.5},
		{"bi neg", NewBipolar(-0.5), -0.5, -0.25},
		{"bi clamp", NewBipolar(-0.5), -5, -1},
	}
	for _, tt := range tests {
		if got := tt.c.Shape(tt.x); math.Abs(got-tt.want) > 1e-12 {
			t.Errorf("%s: got %v want %v", tt.name, got, tt.want)
		}
	}
}

func TestProperties(t *testing.T) {
	for _, tension := range []float64{-0.99, -0.5, 0, 0.2, 0.7, 0.99} {
		u, b := NewUnipolar(tension), NewBipolar(tension)
		if u.Shape(0) != 0 || u.Shape(1) != 1 || b.Shape(0) != 0 || b.Shape(1) != 1 || b.Shape(-1) != -1 {
			t.Fatalf("tension %v: anchors", tension)
		}
		pu, pb := math.Inf(-1), math.Inf(-1)
		for x := -1.0; x <= 1; x += 0.01 {
			yb := b.Shape(x)
			if yb < pb || math.Abs(b.Shape(-x)+yb) > 1e-12 {
				t.Fatalf("tension %v: bipolar not monotone/odd at %v", tension, x)
			}
			pb = yb
			yu := u.Shape(x)
			if yu < pu || yu < 0 || yu > 1 {
				t.Fatalf("tension %v: unipolar bad at %v", tension, x)
			}
			pu = yu
			// positive tension above the line, negative below
			if x > 0 && x < 1 && ((tension > 0 && u.Shape(x) < x) || (tension < 0 && u.Shape(x) > x)) {
				t.Fatalf("tension %v: wrong side of line at %v", tension, x)
			}
		}
		for _, x := range []float64{1e300, -1e300, math.Inf(1), math.Inf(-1)} {
			if y := b.Shape(x); math.IsNaN(y) || math.Abs(y) > 1 {
				t.Fatalf("Shape(%v) = %v", x, y)
			}
		}
	}
}

func TestPanics(t *testing.T) {
	for _, tension := range []float64{1, -1, 2, math.NaN(), math.Inf(1)} {
		for _, f := range []func(float64) *Curve{NewUnipolar, NewBipolar} {
			func() {
				defer func() {
					if recover() == nil {
						t.Errorf("tension %v: no panic", tension)
					}
				}()
				f(tension)
			}()
		}
	}
}
