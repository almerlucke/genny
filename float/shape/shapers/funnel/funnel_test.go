package funnel

import (
	"math"
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/float/shape"
	"github.com/almerlucke/genny/scale"
	"github.com/almerlucke/genny/sieve"
)

var _ shape.Shaper = (*Funnel)(nil)

func TestList(t *testing.T) {
	f := New([]float64{4, 0, 1, 1, 10, math.NaN()})
	tests := []struct{ x, want float64 }{
		{-5, 0}, {0.4, 0}, {0.6, 1}, {1, 1}, {2, 1}, {2.5, 1}, {2.51, 4}, {4, 4}, {7, 4}, {7.5, 10}, {7.6, 10}, {100, 10},
	}
	for _, tt := range tests {
		if got := f.Shape(tt.x); got != tt.want {
			t.Errorf("Shape(%v) = %v want %v", tt.x, got, tt.want)
		}
	}
}

func TestTies(t *testing.T) {
	vals := []float64{0, 2}
	if got := New(vals).Shape(1); got != 0 {
		t.Errorf("Lower: %v", got)
	}
	if got := New(vals).WithTie(Upper).Shape(1); got != 2 {
		t.Errorf("Upper: %v", got)
	}
	f := New(vals).WithTie(Random)
	seen := map[float64]int{}
	genny.Seed(5)
	var first []float64
	for i := 0; i < 200; i++ {
		v := f.Shape(1)
		seen[v]++
		first = append(first, v)
	}
	if seen[0] < 50 || seen[2] < 50 || seen[0]+seen[2] != 200 {
		t.Errorf("random ties not spread: %v", seen)
	}
	genny.Seed(5)
	for i := 0; i < 200; i++ {
		if f.Shape(1) != first[i] {
			t.Fatal("random ties not repeatable with the same seed")
		}
	}
	// no tie: the stream is not touched
	genny.Seed(9)
	a := genny.Float64()
	genny.Seed(9)
	f.Shape(0.3)
	if genny.Float64() != a {
		t.Error("stream drawn without a tie")
	}
}

func TestEmpty(t *testing.T) {
	for name, f := range map[string]*Funnel{
		"nil list":   New(nil),
		"empty list": New([]float64{}),
		"only NaN":   New([]float64{math.NaN()}),
		"sieve":      FromSieve(sieve.Of(0, 10), 60),
	} {
		for _, x := range []float64{-3, 0.4, 1e6} {
			if got := f.Shape(x); got != x {
				t.Errorf("%s: Shape(%v) = %v", name, x, got)
			}
		}
	}
}

func TestSieve(t *testing.T) {
	f := FromSieve(sieve.Of(0, 11, 0, 4, 7), 60)
	tests := []struct{ x, want float64 }{{60, 60}, {61, 60}, {62.5, 64}, {65.4, 64}, {65.6, 67}, {80, 67}, {0, 60}}
	for _, tt := range tests {
		if got := f.Shape(tt.x); got != tt.want {
			t.Errorf("Shape(%v) = %v want %v", tt.x, got, tt.want)
		}
	}
}

func TestScaleMatchesQuantize(t *testing.T) {
	sc := scale.New(57, scale.Minor)
	f := FromScale(sc) // Lower ties, like Quantize
	for i := -300; i <= 3000; i++ {
		x := float64(i) / 20
		if got, want := f.Shape(x), sc.Quantize(x); got != want {
			t.Fatalf("Shape(%v) = %v want %v", x, got, want)
		}
	}
	u := FromScale(scale.New(0, []float64{0, 4})).WithTie(Upper) // steps 0, 4, 12, 16, 24 ... gap 8
	if got := u.Shape(8); got != 12 {
		t.Errorf("upper tie in the gap: %v", got)
	}
}

func TestFinite(t *testing.T) {
	fs := []*Funnel{New([]float64{1, 2}), FromScale(scale.New(0, scale.Major))}
	for _, f := range fs {
		for _, x := range []float64{1e300, -1e300, 1e13, -1e13, 12345.678} {
			if y := f.Shape(x); math.IsNaN(y) || math.IsInf(y, 0) {
				t.Errorf("Shape(%v) = %v", x, y)
			}
		}
		if !math.IsNaN(f.Shape(math.NaN())) || !math.IsInf(f.Shape(math.Inf(1)), 1) {
			t.Error("non finite input is returned unchanged")
		}
	}
}

func TestPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("FromScale(nil) must panic")
		}
	}()
	FromScale(nil)
}
