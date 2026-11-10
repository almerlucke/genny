package rand_test

import (
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/float/rand"
)

func TestRandRange(t *testing.T) {
	genny.Seed(1)
	tests := []struct{ low, high float64 }{{0, 1}, {-5, 5}, {10, 11}, {3, -3}}
	for _, tt := range tests {
		r := rand.New(tt.low, tt.high)
		lo, hi := tt.low, tt.high
		if lo > hi {
			lo, hi = hi, lo
		}
		var min, max = hi, lo
		for i := 0; i < 2000; i++ {
			v := r.Generate()
			if v < lo || v >= hi {
				t.Fatalf("[%v,%v]: %v out of range", tt.low, tt.high, v)
			}
			min, max = minf(min, v), maxf(max, v)
		}
		if min > lo+(hi-lo)*0.05 || max < hi-(hi-lo)*0.05 {
			t.Fatalf("[%v,%v]: range barely used (%v..%v)", tt.low, tt.high, min, max)
		}
	}
}

func minf(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxf(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func TestRandSettersAndState(t *testing.T) {
	r := rand.New(0, 1)
	if !r.Continuous() || r.Done() {
		t.Fatal("continuous, never done")
	}
	r.Reset()
	r.SetLow(5)
	r.SetHigh(5)
	if v := r.Generate(); v != 5 {
		t.Fatalf("zero width range gives that value, got %v", v)
	}
	r.SetLow(100)
	r.SetHigh(200)
	if v := r.Generate(); v < 100 || v >= 200 {
		t.Fatalf("got %v", v)
	}
}

func TestRandSeedRepeatable(t *testing.T) {
	run := func() []float64 {
		genny.Seed(31)
		r := rand.New(-1, 1)
		out := make([]float64, 20)
		for i := range out {
			out[i] = r.Generate()
		}
		return out
	}
	a, b := run(), run()
	for i := range a {
		if a[i] != b[i] {
			t.Fatal("same seed must repeat")
		}
	}
	genny.Seed(32)
	if rand.New(-1, 1).Generate() == a[0] {
		t.Fatal("different seed should differ")
	}
}
