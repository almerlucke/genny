package bitcrush

import (
	"math"
	"testing"

	"github.com/almerlucke/genny/float/shape"
)

var _ shape.Shaper = (*Crusher)(nil)

func TestQuantize(t *testing.T) {
	tests := []struct {
		x    float64
		bits int
		want float64
	}{
		{0, 1, 0}, {0.6, 1, 0}, {-0.6, 1, -1}, {-1, 1, -1}, {1, 1, 0},
		{0.3, 2, 0.5}, {0.2, 2, 0}, {1, 2, 0.5}, {-1, 2, -1}, {-0.8, 2, -1}, {0.75, 2, 0.5},
		{0.5, 8, 0.5}, {1, 8, 127.0 / 128}, {1.5, 8, 127.0 / 128}, {-9, 8, -1}, {math.NaN(), 8, 0},
		{0.3, 24, 0.3},
	}
	for _, tt := range tests {
		if got := Quantize(tt.x, tt.bits); math.Abs(got-tt.want) > 1e-6 {
			t.Errorf("Quantize(%v, %d) = %v want %v", tt.x, tt.bits, got, tt.want)
		}
	}
}

func TestLevelCount(t *testing.T) {
	for bits := 1; bits <= 8; bits++ {
		seen := map[float64]bool{}
		for i := -5000; i <= 5000; i++ {
			seen[Quantize(float64(i)/5000, bits)] = true
		}
		if len(seen) != 1<<bits {
			t.Errorf("bits %d: %d levels want %d", bits, len(seen), 1<<bits)
		}
	}
}

func TestSampleRateReduction(t *testing.T) {
	c := New(24, 3, 1)
	in := []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7}
	want := []float64{0.1, 0.1, 0.1, 0.4, 0.4, 0.4, 0.7}
	for i, x := range in {
		if got := c.Shape(x); math.Abs(got-want[i]) > 1e-6 {
			t.Errorf("step %d: got %v want %v", i, got, want[i])
		}
	}
	c.Reset()
	if got := c.Shape(0.9); math.Abs(got-0.9) > 1e-6 {
		t.Errorf("after Reset got %v", got)
	}
	if got := c.Shape(0.2); math.Abs(got-0.9) > 1e-6 {
		t.Errorf("held after Reset got %v", got)
	}
}

func TestDeterministic(t *testing.T) {
	run := func() []float64 {
		c := New(4, 2, 0.7)
		var out []float64
		for i := 0; i < 50; i++ {
			out = append(out, c.Shape(math.Sin(float64(i)/3)))
		}
		return out
	}
	a, b := run(), run()
	for i := range a {
		if a[i] != b[i] {
			t.Fatal("not deterministic")
		}
	}
}

func TestMix(t *testing.T) {
	c := New(1, 1, 0.5)
	// x=0.6 quantizes to 0 at one bit: 0.5*0.6 + 0.5*0
	if got := c.Shape(0.6); math.Abs(got-0.3) > 1e-12 {
		t.Errorf("got %v", got)
	}
	dry := New(1, 4, 0)
	for _, x := range []float64{0.3, -0.7, 0.9} {
		if dry.Shape(x) != x {
			t.Error("mix 0 must pass the input")
		}
	}
}

func TestFinite(t *testing.T) {
	c := New(3, 2, 0.8)
	for _, x := range []float64{0, 1e300, -1e300, math.NaN(), math.Inf(1), math.Inf(-1)} {
		if y := c.Shape(x); math.IsNaN(y) || math.IsInf(y, 0) && !math.IsInf(x, 0) {
			t.Errorf("Shape(%v) = %v", x, y)
		}
	}
}

func TestPanics(t *testing.T) {
	tests := map[string]func(){
		"bits 0":    func() { New(0, 1, 1) },
		"bits 25":   func() { New(25, 1, 1) },
		"factor 0":  func() { New(8, 0, 1) },
		"mix -1":    func() { New(8, 1, -1) },
		"mix 2":     func() { New(8, 1, 2) },
		"mix NaN":   func() { New(8, 1, math.NaN()) },
		"quantize0": func() { Quantize(0, 0) },
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
