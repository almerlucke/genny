package harmonics

import (
	"math"
	"testing"

	"github.com/almerlucke/genny/float/shape"
)

var _ shape.Shaper = (*Harmonics)(nil)

func TestKnownValues(t *testing.T) {
	tests := []struct {
		name string
		h    *Harmonics
		x    float64
		want float64
	}{
		{"fundamental", New([]float64{1}, nil, false), 0.5, 1},
		{"two", New([]float64{1, 0.5}, nil, false), 0.25, math.Sqrt2/2 + 0.5},
		{"third", New([]float64{0, 0, 2}, nil, false), 1.0 / 6, 2},
		{"phase", New([]float64{1}, []float64{math.Pi / 2}, false), 0, 1},
		{"short phases", New([]float64{1, 1}, []float64{math.Pi / 2}, false), 0, 1},
		{"zero", New([]float64{1, 1}, nil, false), 0, 0},
		{"normalised", New([]float64{3}, nil, true), 0.5, 1},
	}
	for _, tt := range tests {
		if got := tt.h.Shape(tt.x); math.Abs(got-tt.want) > 1e-12 {
			t.Errorf("%s: got %v want %v", tt.name, got, tt.want)
		}
	}
}

func TestProperties(t *testing.T) {
	sets := [][]float64{{1}, {1, 0.5, 0.25}, {1, -1, 1, -1, 1}, {0.2, 0, 0, 0.9}, {5, 5, 5, 5, 5, 5, 5, 5}}
	for _, amps := range sets {
		h := New(amps, nil, true)
		peak := 0.0
		for i := 0; i <= 20000; i++ {
			x := -1 + float64(i)/10000
			y := h.Shape(x)
			peak = math.Max(peak, math.Abs(y))
			if math.Abs(y) > 1+1e-6 || math.Abs(h.Shape(-x)+y) > 1e-9 { // odd without phases
				t.Fatalf("%v: bad at %v: %v", amps, x, y)
			}
		}
		if math.Abs(peak-1) > 1e-6 {
			t.Fatalf("%v: peak %v", amps, peak)
		}
		for _, x := range []float64{1e300, -1e300, math.MaxFloat64, math.Inf(1)} {
			if y := h.Shape(x); math.IsNaN(y) || math.IsInf(y, 0) {
				t.Fatalf("%v: Shape(%v) = %v", amps, x, y)
			}
		}
	}
	// the curve repeats every 2 in x
	h := New([]float64{1, 0.5, 0.3}, []float64{0.1, 0.2, 0.3}, false)
	if math.Abs(h.Shape(0.3)-h.Shape(2.3)) > 1e-9 {
		t.Fatal("not periodic")
	}
}

func TestInputNotAliased(t *testing.T) {
	amps := []float64{1, 1}
	h := New(amps, nil, false)
	amps[0] = 100
	if got := h.Shape(0.5); math.Abs(got-1) > 1e-12 {
		t.Fatalf("constructor kept the slice: %v", got)
	}
}

func TestPanics(t *testing.T) {
	for name, f := range map[string]func(){
		"empty":       func() { New(nil, nil, false) },
		"nan amp":     func() { New([]float64{math.NaN()}, nil, false) },
		"inf phase":   func() { New([]float64{1}, []float64{math.Inf(1)}, false) },
		"many phases": func() { New([]float64{1}, []float64{0, 0}, false) },
		"zero norm":   func() { New([]float64{0, 0}, nil, true) },
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
