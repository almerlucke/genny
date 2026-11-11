package dist

import (
	"math"
	"slices"
	"testing"

	"github.com/almerlucke/genny"
)

func TestBeta(t *testing.T) {
	for _, c := range []struct{ a, b float64 }{{1, 1}, {2, 5}, {5, 2}, {0.5, 0.5}, {3, 3}} {
		genny.Seed(3)
		mean, _, lo, hi := stats(NewBeta(c.a, c.b), 100000)
		want := c.a / (c.a + c.b)
		if math.Abs(mean-want) > 0.01 || lo < 0 || hi > 1 {
			t.Errorf("beta(%v,%v): mean %v want %v range [%v, %v]", c.a, c.b, mean, want, lo, hi)
		}
	}
	_, _, lo, hi := stats(NewBeta(2, 2).Scale(10, 20), 10000)
	if lo < 10 || hi > 20 {
		t.Errorf("scaled range [%v, %v]", lo, hi)
	}
}

func TestSampleBetaInvalid(t *testing.T) {
	if SampleBeta(0, 1) != 0.5 || SampleBeta(1, -1) != 0.5 {
		t.Error("invalid parameters should give 0.5")
	}
}

func TestCauchy(t *testing.T) {
	genny.Seed(4)
	g := NewCauchy(10, 2)
	var below, within int
	for range 100000 {
		v := g.Generate()
		if v < 10 {
			below++
		}
		if math.Abs(v-10) < 2 {
			within++
		}
	}
	if f := float64(below) / 100000; math.Abs(f-0.5) > 0.01 {
		t.Errorf("median off: %v below the location", f)
	}
	if f := float64(within) / 100000; math.Abs(f-0.5) > 0.01 {
		t.Errorf("%v within one scale, want 0.5", f)
	}
	_, _, lo, hi := stats(NewCauchy(0, 1).Clamp(-1, 1), 10000)
	if lo != -1 || hi != 1 {
		t.Errorf("clamp [%v, %v]", lo, hi)
	}
	_, _, lo, hi = stats(NewCauchy(0, 1).Retry(-1, 1, 1000), 10000)
	if lo < -1 || hi > 1 || lo == -1 || hi == 1 {
		t.Errorf("retry [%v, %v]", lo, hi)
	}
	// without retries it is a clamp
	_, _, lo, hi = stats(NewCauchy(0, 1).Retry(-1, 1, 0), 10000)
	if lo != -1 || hi != 1 {
		t.Errorf("retry 0 [%v, %v]", lo, hi)
	}
}

func TestWeibull(t *testing.T) {
	genny.Seed(5)
	// shape 1 is exponential: mean threshold + spread
	mean, _, lo, _ := stats(NewWeibull(2, 3, 1), 200000)
	if math.Abs(mean-5) > 0.05 || lo < 2 {
		t.Errorf("mean %v min %v", mean, lo)
	}
	// shape 2: mean spread*sqrt(pi)/2
	mean, _, _, _ = stats(NewWeibull(0, 2, 2), 200000)
	if want := 2 * math.Sqrt(math.Pi) / 2; math.Abs(mean-want) > 0.02 {
		t.Errorf("mean %v want %v", mean, want)
	}
	_, _, lo, hi := stats(NewWeibull(0, 5, 1).Clamp(1, 3), 10000)
	if lo != 1 || hi != 3 {
		t.Errorf("clamp [%v, %v]", lo, hi)
	}
	_, _, lo, hi = stats(NewWeibull(0, 5, 1).Retry(1, 3, 100), 10000)
	if lo < 1 || hi > 3 {
		t.Errorf("retry [%v, %v]", lo, hi)
	}
}

func TestLehmer(t *testing.T) {
	genny.Seed(1)
	a := NewLehmer(42)
	var first []float64
	for range 20 {
		first = append(first, a.Generate())
	}
	genny.Seed(99) // the genny stream is not used
	b := NewLehmer(42)
	var second []float64
	for range 20 {
		second = append(second, b.Generate())
	}
	if !slices.Equal(first, second) {
		t.Error("same seed should give the same values")
	}
	a.Reset()
	if a.Generate() != first[0] {
		t.Error("reset should restart")
	}
	for _, v := range first {
		if v < 0 || v >= 1 {
			t.Errorf("value %v out of range", v)
		}
	}
	if slices.Equal(first, take20(NewLehmer(43))) {
		t.Error("another seed gave the same values")
	}
	// X(n+1) = (5*X + 3) mod 16 from 7: 6, 1, 8, 11
	g := NewLehmerParams(7, 5, 3, 16)
	for _, want := range []float64{6, 1, 8, 11} {
		if v := g.Generate(); v != want/16 {
			t.Errorf("got %v want %v", v, want/16)
		}
	}
	// no overflow with large constants
	big := NewLehmerParams(1, 1<<62-1, 1<<62-1, 1<<62)
	for range 10 {
		if v := big.Generate(); v < 0 || v >= 1 {
			t.Errorf("big value %v", v)
		}
	}
}

func take20(g *Lehmer) (out []float64) {
	for range 20 {
		out = append(out, g.Generate())
	}
	return
}

func TestPink(t *testing.T) {
	genny.Seed(6)
	p := NewPink(8)
	const n = 1 << 16
	vals := make([]float64, n)
	var sum float64
	for i := range vals {
		vals[i] = p.Generate()
		sum += vals[i]
		if vals[i] < 0 || vals[i] >= 1 {
			t.Fatalf("value %v out of range", vals[i])
		}
	}
	if m := sum / n; math.Abs(m-0.5) > 0.05 {
		t.Errorf("mean %v", m)
	}
	// more correlated than white noise: neighbours are closer than random pairs (1/3 apart on average)
	var diff float64
	for i := 1; i < n; i++ {
		diff += math.Abs(vals[i] - vals[i-1])
	}
	if d := diff / (n - 1); d > 0.25 {
		t.Errorf("mean step %v, pink noise should be smoother than white (0.33)", d)
	}
	p.Reset()
	if p.counter != 0 {
		t.Error("reset")
	}
}

func TestMoreInvalid(t *testing.T) {
	for name, f := range map[string]func(){
		"beta a":         func() { NewBeta(0, 1) },
		"beta b":         func() { NewBeta(1, -1) },
		"cauchy scale":   func() { NewCauchy(0, 0) },
		"weibull spread": func() { NewWeibull(0, 0, 1) },
		"weibull shape":  func() { NewWeibull(0, 1, 0) },
		"lehmer m":       func() { NewLehmerParams(1, 1, 1, 0) },
		"pink rows":      func() { NewPink(0) },
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
