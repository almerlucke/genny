package ramp

import (
	"math"
	"testing"
)

func check(t *testing.T, got, want []float64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range got {
		if math.Abs(got[i]-want[i]) > 1e-12 {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}

func take(r *Ramp, n int) []float64 {
	out := make([]float64, n)
	for i := range out {
		out[i] = r.Generate()
	}
	return out
}

func TestRampLinear(t *testing.T) {
	r := New(5, 0, 1, 1)
	if r.Continuous() || r.Done() {
		t.Fatal("fresh ramp")
	}
	check(t, take(r, 4), []float64{0, 0.25, 0.5, 0.75})
	if r.Done() {
		t.Fatal("not done before the last value")
	}
	check(t, take(r, 1), []float64{1})
	if !r.Done() {
		t.Fatal("done after steps values")
	}
	check(t, take(r, 2), []float64{1, 1}) // holds the end value
}

func TestRampMinMaxAndExponent(t *testing.T) {
	check(t, take(New(3, 10, 20, 1), 3), []float64{10, 15, 20})
	check(t, take(New(3, 20, 10, 1), 3), []float64{20, 15, 10})
	check(t, take(New(5, 0, 1, 2), 5), []float64{0, 0.0625, 0.25, 0.5625, 1})
	check(t, take(New(3, 0, 4, 0.5), 3), []float64{0, 4 * math.Sqrt(0.5), 4})
}

func TestRampReset(t *testing.T) {
	r := New(3, 0, 1, 1)
	take(r, 3)
	r.Reset()
	if r.Done() {
		t.Fatal("reset clears done")
	}
	check(t, take(r, 3), []float64{0, 0.5, 1})
}

func TestRampForever(t *testing.T) {
	r := NewForever(3, 0, 1, 1)
	if !r.Continuous() {
		t.Fatal("forever is continuous")
	}
	check(t, take(r, 6), []float64{0, 0.5, 1, 1, 1, 1})
	if r.Done() {
		t.Fatal("a continuous ramp is never done")
	}
	r.MakeContinuous(false)
	if r.Continuous() || !r.Done() {
		t.Fatal("MakeContinuous(false) makes it end")
	}
}

func TestRampDegenerateSteps(t *testing.T) {
	for _, steps := range []int{1, 0, -2} {
		r := New(steps, 3, 9, 1)
		for i := 0; i < 3; i++ {
			v := r.Generate()
			if math.IsNaN(v) || math.IsInf(v, 0) || v != 3 {
				t.Fatalf("steps=%d: got %v, want the start value", steps, v)
			}
		}
		if !r.Done() {
			t.Fatalf("steps=%d must be done after one value", steps)
		}
	}
}
