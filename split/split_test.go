package split

import (
	"math"
	"slices"
	"testing"

	"github.com/almerlucke/genny"
)

func sum(v []float64) (s float64) {
	for _, x := range v {
		s += x
	}
	return
}

func TestSplitSumsAndLimits(t *testing.T) {
	genny.Seed(5)
	for _, c := range []struct {
		value, lower, prob float64
		branches           int
	}{
		{4, 0.25, 0.5, 2}, {4, 0.25, 1, 2}, {3, 0.1, 0.7, 3}, {1, 0.01, 0.9, 5}, {10, 0.5, 0.3, 4},
	} {
		for range 50 {
			parts := Split(c.value, c.lower, c.prob, c.branches)
			if math.Abs(sum(parts)-c.value) > 1e-9 {
				t.Fatalf("%+v: sum %v", c, sum(parts))
			}
			for _, p := range parts {
				// a part is only shorter than lower if it was never split
				if p < c.lower-1e-12 && p != c.value {
					t.Fatalf("%+v: part %v below lower", c, p)
				}
			}
		}
	}
}

func TestSplitExtremes(t *testing.T) {
	genny.Seed(1)
	if got := Split(4, 0.5, 0, 2); !slices.Equal(got, []float64{4}) {
		t.Errorf("prob 0 = %v", got)
	}
	if got := Split(4, 0.5, -3, 2); !slices.Equal(got, []float64{4}) {
		t.Errorf("negative prob = %v", got)
	}
	// prob 1 splits down to the smallest parts that are not below lower
	if got := Split(4, 0.5, 1, 2); len(got) != 8 || got[0] != 0.5 {
		t.Errorf("prob 1 = %v", got)
	}
	if got := Split(4, 0.5, 9, 2); len(got) != 8 {
		t.Errorf("prob above 1 = %v", got)
	}
	// lower above the value: no split possible
	if got := Split(1, 2, 1, 2); !slices.Equal(got, []float64{1}) {
		t.Errorf("lower > value = %v", got)
	}
	// lower exactly reachable
	if got := Split(1, 0.5, 1, 2); !slices.Equal(got, []float64{0.5, 0.5}) {
		t.Errorf("exact = %v", got)
	}
	if Split(0, 0.5, 1, 2) != nil || Split(-1, 0.5, 1, 2) != nil || Split(math.NaN(), 0.5, 1, 2) != nil {
		t.Error("no duration should give no parts")
	}
}

func TestSplitPanics(t *testing.T) {
	for name, f := range map[string]func(){
		"branches": func() { Split(1, 0.1, 0.5, 1) },
		"lower 0":  func() { Split(1, 0, 0.5, 2) },
		"lower <0": func() { Split(1, -1, 0.5, 2) },
		"metric":   func() { Metric(1, 0) },
		"gen":      func() { NewGen(0, 0.1, 0.5, 2) },
		"gen br":   func() { NewGen(1, 0.1, 0.5, 0) },
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s did not panic", name)
				}
			}()
			f()
		}()
	}
}

func TestMetric(t *testing.T) {
	if got := Metric(3, 4); !slices.Equal(got, []float64{0.75, 0.75, 0.75, 0.75}) {
		t.Errorf("Metric = %v", got)
	}
	if got := Metric(2, 1); !slices.Equal(got, []float64{2}) {
		t.Errorf("Metric n=1 = %v", got)
	}
}

func TestGenCycles(t *testing.T) {
	genny.Seed(9)
	g := NewGen(4, 0.25, 0.6, 2)
	if !g.Continuous() || g.Done() {
		t.Error("gen loops forever")
	}
	// every cycle of parts sums to the value; collect cycles by summing
	cycles, acc := 0, 0.0
	var all []float64
	for range 400 {
		v := g.Generate()
		all = append(all, v)
		acc += v
		if math.Abs(acc-4) < 1e-9 {
			cycles++
			acc = 0
		} else if acc > 4 {
			t.Fatal("a cycle overshoots the value")
		}
	}
	if cycles < 10 {
		t.Errorf("only %d cycles", cycles)
	}
	// the cycles differ
	genny.Seed(9)
	g = NewGen(4, 0.25, 0.6, 2)
	first := make([]float64, 0, 20)
	for range 20 {
		first = append(first, g.Generate())
	}
	if !slices.Equal(first, all[:20]) {
		t.Error("same seed, other values")
	}
	g.Reset()
	if v := g.Generate(); v <= 0 || v > 4 {
		t.Errorf("after reset: %v", v)
	}
}
