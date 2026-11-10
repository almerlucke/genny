package dist

import (
	"math"
	"testing"

	"github.com/almerlucke/genny"
)

func stats(g interface{ Generate() float64 }, n int) (mean, std, lo, hi float64) {
	var sum, sq float64
	lo, hi = math.Inf(1), math.Inf(-1)
	for range n {
		v := g.Generate()
		sum += v
		sq += v * v
		lo, hi = min(lo, v), max(hi, v)
	}
	mean = sum / float64(n)
	return mean, math.Sqrt(sq/float64(n) - mean*mean), lo, hi
}

func TestGaussian(t *testing.T) {
	genny.Seed(1)
	mean, std, _, _ := stats(NewGaussian(5, 2), 200000)
	if math.Abs(mean-5) > 0.03 || math.Abs(std-2) > 0.03 {
		t.Errorf("mean %v std %v", mean, std)
	}
	_, _, lo, hi := stats(NewGaussian(0.5, 1).Clamp(0, 1), 10000)
	if lo != 0 || hi != 1 {
		t.Errorf("clamp [%v, %v]", lo, hi)
	}
	// about 68% within one standard deviation
	g := NewGaussian(0, 1)
	in := 0
	for range 100000 {
		if math.Abs(g.Generate()) < 1 {
			in++
		}
	}
	if f := float64(in) / 100000; f < 0.67 || f > 0.70 {
		t.Errorf("%v within one sigma", f)
	}
}

func TestExponential(t *testing.T) {
	genny.Seed(2)
	mean, _, lo, _ := stats(NewExponential(3), 200000)
	if math.Abs(mean-3) > 0.05 || lo < 0 {
		t.Errorf("mean %v min %v", mean, lo)
	}
	_, _, lo, hi := stats(NewExponential(3).Clamp(1, 4), 10000)
	if lo != 1 || hi != 4 {
		t.Errorf("clamp [%v, %v]", lo, hi)
	}
}

func TestInvalid(t *testing.T) {
	for name, f := range map[string]func(){
		"negative stddev": func() { NewGaussian(0, -1) },
		"zero mean":       func() { NewExponential(0) },
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
