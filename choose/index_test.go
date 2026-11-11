package choose

import (
	"math"
	"testing"

	"github.com/almerlucke/genny"
)

func histogram(d Distribution, n, draws int) []float64 {
	items := make([]int, n)
	for i := range items {
		items[i] = i
	}
	x := NewIndex(items, d)
	h := make([]float64, n)
	for range draws {
		h[x.Generate()]++
	}
	for i := range h {
		h[i] /= float64(draws)
	}
	return h
}

func TestIndexDistributions(t *testing.T) {
	genny.Seed(11)
	const draws = 100000
	tests := []struct {
		name string
		d    Distribution
		want []float64 // expected chance per index for a list of 4
		tol  float64
	}{
		{"uniform", Uniform(), []float64{.25, .25, .25, .25}, 0.01},
		{"linear", Linear(), []float64{7. / 16, 5. / 16, 3. / 16, 1. / 16}, 0.01},
		{"ratio", Ratio(4, 2, 1, 1), []float64{.5, .25, .125, .125}, 0.01},
		{"ratio short", Ratio(1, 1), []float64{.5, .5, 0, 0}, 0.01},
		{"triangle start", Triangle(0), []float64{7. / 16, 5. / 16, 3. / 16, 1. / 16}, 0.01},
		{"triangle end", Triangle(1), []float64{1. / 16, 3. / 16, 5. / 16, 7. / 16}, 0.01},
		{"beta 1 1", Beta(1, 1), []float64{.25, .25, .25, .25}, 0.01},
		{"gaussian center", Gaussian(1, 0), []float64{0, 1, 0, 0}, 0},
	}
	for _, tc := range tests {
		got := histogram(tc.d, 4, draws)
		for i := range got {
			if math.Abs(got[i]-tc.want[i]) > tc.tol {
				t.Errorf("%s: %v, want %v", tc.name, got, tc.want)
				break
			}
		}
	}
}

func TestIndexShapes(t *testing.T) {
	genny.Seed(12)
	increasing := func(h []float64) bool {
		for i := 1; i < len(h); i++ {
			if h[i] > h[i-1] {
				return false
			}
		}
		return true
	}
	if h := histogram(Exponential(3), 6, 50000); !increasing(h) || h[0] < 2*h[5] {
		t.Errorf("exponential should fall: %v", h)
	}
	if h := histogram(Weibull(0.5, 0.7), 6, 50000); !increasing(h) {
		t.Errorf("weibull below 1 should fall: %v", h)
	}
	h := histogram(Gaussian(4, 1), 9, 50000)
	if h[4] < h[3] || h[4] < h[5] || h[0] > 0.01 {
		t.Errorf("gaussian around 4: %v", h)
	}
	h = histogram(Cauchy(0.5, 0.05), 10, 50000)
	if h[4]+h[5] < 0.5 {
		t.Errorf("cauchy around the middle: %v", h)
	}
	h = histogram(Beta(0.4, 0.4), 10, 50000)
	if h[0] < 2*h[5] || h[9] < 2*h[4] {
		t.Errorf("beta below 1 favours the ends: %v", h)
	}
	for _, d := range []Distribution{Gaussian(100, 1), Gaussian(-100, 1), Cauchy(5, 1), Weibull(9, 1)} {
		histogram(d, 3, 1000) // never out of range (would panic)
	}
}

func TestIndexEdges(t *testing.T) {
	x := NewIndex([]string(nil), Uniform())
	if !x.Done() || x.Continuous() || x.Generate() != "" || x.Pick() != -1 {
		t.Error("empty list should be done and give the zero value")
	}
	one := NewIndex([]string{"a"}, Linear())
	for range 10 {
		if one.Generate() != "a" {
			t.Error("single element")
		}
	}
	// out of range and NaN positions are clamped
	for _, p := range []float64{-5, 5, math.NaN(), math.Inf(1)} {
		x := NewIndex([]int{1, 2, 3}, func(int) float64 { return p })
		if v := x.Generate(); v < 1 || v > 3 {
			t.Errorf("position %v gave %v", p, v)
		}
	}
	items := []int{1, 2}
	y := NewIndex(items, Uniform())
	items[0] = 99
	for range 20 {
		if y.Generate() == 99 {
			t.Error("items should be copied")
		}
	}
	y.Reset()
}

func TestChangingWeights(t *testing.T) {
	genny.Seed(13)
	c := NewChangingWeights([]string{"a", "b", "c"}, 5, []float64{1, 0, 0}, []float64{0, 0, 1}, []float64{0, 1})
	var got []string
	for range 30 {
		got = append(got, c.Generate())
	}
	for i, v := range got {
		want := []string{"a", "c", "b"}[(i/5)%3]
		if v != want {
			t.Fatalf("pick %d is %s, want %s: %v", i, v, want, got)
		}
	}
	c.Reset()
	if c.Generate() != "a" {
		t.Error("reset should use the first set again")
	}

	// weights are respected
	w := NewChangingWeights([]int{0, 1}, 1000000, []float64{3, 1})
	n := 0
	for range 40000 {
		n += w.Generate()
	}
	if f := float64(n) / 40000; math.Abs(f-0.25) > 0.015 {
		t.Errorf("chance of 1 is %v, want 0.25", f)
	}

	// a set longer than the list, and one with weight only past the end
	l := NewChangingWeights([]int{7}, 1, []float64{1, 5}, []float64{0, 1})
	for range 6 {
		if l.Generate() != 7 {
			t.Error("single element")
		}
	}

	e := NewChangingWeights[int](nil, 1, []float64{1})
	if !e.Done() || e.Continuous() || e.Generate() != 0 {
		t.Error("empty list")
	}
}

func TestIndexInvalid(t *testing.T) {
	for name, f := range map[string]func(){
		"nil distribution": func() { NewIndex([]int{1}, nil) },
		"exponential rate": func() { Exponential(0) },
		"gaussian sd":      func() { Gaussian(0, -1) },
		"beta":             func() { Beta(0, 1) },
		"cauchy":           func() { Cauchy(0.5, 0) },
		"weibull":          func() { Weibull(0, 1) },
		"triangle":         func() { Triangle(1.5) },
		"ratio negative":   func() { Ratio(1, -1) },
		"ratio none":       func() { Ratio() },
		"ratio zeros":      func() { Ratio(0, 0) },
		"every":            func() { NewChangingWeights([]int{1}, 0, []float64{1}) },
		"no sets":          func() { NewChangingWeights([]int{1}, 1) },
		"negative weight":  func() { NewChangingWeights([]int{1}, 1, []float64{-1}) },
		"zero weights":     func() { NewChangingWeights([]int{1}, 1, []float64{0, 0}) },
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s should panic", name)
				}
			}()
			f()
		}()
	}
}
