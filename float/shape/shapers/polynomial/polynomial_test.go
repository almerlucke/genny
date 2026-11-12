package polynomial

import (
	"math"
	"testing"

	"github.com/almerlucke/genny/float/shape"
)

var (
	_ shape.Shaper = (*Polynomial)(nil)
	_ shape.Shaper = (*Normalized)(nil)
)

func TestHorner(t *testing.T) {
	tests := []struct {
		c    []float64
		x    float64
		want float64
	}{
		{[]float64{5}, 3, 5},
		{[]float64{1, 2, 3}, 2, 17},     // 1 + 4 + 12
		{[]float64{1, 0, -2}, 0.5, 0.5}, // 1 - 2*0.25
		{[]float64{0, 0, 0, 1}, -2, -8},
		{[]float64{1, 1}, math.NaN(), 0},
	}
	for _, tt := range tests {
		p := New(tt.c...)
		if got := p.Shape(tt.x); math.Abs(got-tt.want) > 1e-12 {
			t.Errorf("%v at %v = %v want %v", tt.c, tt.x, got, tt.want)
		}
	}
	p := New(1, 2, 3)
	if p.Degree() != 2 {
		t.Error("Degree")
	}
	c := p.Coefficients()
	c[0] = 99
	if p.Shape(0) != 1 {
		t.Error("coefficients must be copied")
	}
}

func TestLegendreCoefficients(t *testing.T) {
	tests := []struct {
		n    int
		want []float64
	}{
		{0, []float64{1}},
		{1, []float64{0, 1}},
		{2, []float64{-0.5, 0, 1.5}},
		{3, []float64{0, -1.5, 0, 2.5}},
		{4, []float64{3.0 / 8, 0, -30.0 / 8, 0, 35.0 / 8}},
	}
	for _, tt := range tests {
		got := NewLegendre(tt.n).Coefficients()
		if len(got) != len(tt.want) {
			t.Fatalf("P%d: %v want %v", tt.n, got, tt.want)
		}
		for i := range got {
			if math.Abs(got[i]-tt.want[i]) > 1e-12 {
				t.Errorf("P%d: %v want %v", tt.n, got, tt.want)
			}
		}
	}
}

func TestLegendreProperties(t *testing.T) {
	for n := 0; n <= 20; n++ {
		p := NewLegendre(n)
		if math.Abs(p.Shape(1)-1) > 1e-8 {
			t.Errorf("P%d(1) = %v", n, p.Shape(1))
		}
		sign := 1.0
		if n%2 == 1 {
			sign = -1
		}
		if math.Abs(p.Shape(-1)-sign) > 1e-8 {
			t.Errorf("P%d(-1) = %v", n, p.Shape(-1))
		}
		for i := 0; i <= 100; i++ {
			x := -1 + float64(i)/50
			if y := p.Shape(x); math.Abs(y) > 1+1e-8 || math.IsNaN(y) {
				t.Fatalf("P%d(%v) = %v", n, x, y)
			}
			if d := p.Shape(-x) - sign*p.Shape(x); math.Abs(d) > 1e-8 {
				t.Fatalf("P%d parity at %v", n, x)
			}
		}
	}
}

func TestNormalized(t *testing.T) {
	n := NewNormalized(New(0, 0, 4)) // 4x^2, peak 4
	if math.Abs(n.Gain()-0.25) > 1e-12 {
		t.Errorf("gain %v", n.Gain())
	}
	if math.Abs(n.Shape(1)-1) > 1e-12 || math.Abs(n.Shape(0.5)-0.25) > 1e-12 {
		t.Error("scaled values")
	}
	// peak inside the range: x - x^3 has peak 2/(3 sqrt 3) at x = 1/sqrt 3
	m := NewNormalized(New(0, 1, 0, -1))
	peak := 2 / (3 * math.Sqrt(3))
	if math.Abs(m.Gain()-1/peak) > 1e-4 {
		t.Errorf("gain %v want %v", m.Gain(), 1/peak)
	}
	for i := 0; i <= 1000; i++ {
		if math.Abs(m.Shape(-1+float64(i)/500)) > 1+1e-6 {
			t.Fatal("normalized peak above 1")
		}
	}
	if g := NewNormalized(New(0)).Gain(); g != 1 {
		t.Errorf("zero polynomial gain %v", g)
	}
}

func TestOverflowClamped(t *testing.T) {
	p := New(0, 0, 0, 1e300, -1e300)
	for _, x := range []float64{1e10, -1e10, 1e300} {
		if y := p.Shape(x); math.IsNaN(y) || math.IsInf(y, 0) {
			t.Errorf("Shape(%v) = %v", x, y)
		}
	}
}

func TestPanics(t *testing.T) {
	tests := map[string]func(){
		"no coefficients":  func() { New() },
		"NaN coefficient":  func() { New(1, math.NaN()) },
		"Inf coefficient":  func() { New(math.Inf(1)) },
		"legendre -1":      func() { NewLegendre(-1) },
		"legendre 21":      func() { NewLegendre(21) },
		"normalized nil":   func() { NewNormalized(nil) },
		"normalized n < 2": func() { NewNormalizedN(New(1), 1) },
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
