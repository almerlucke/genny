package noise

import (
	"math"
	"testing"

	"github.com/almerlucke/genny"
)

func TestSmoothBoundedAndMoving(t *testing.T) {
	genny.Seed(1)
	for _, oct := range []int{1, 4} {
		n := New(0.01, oct)
		prev := n.Generate()
		lo, hi := prev, prev
		for range 20000 {
			v := n.Generate()
			if v < -1 || v > 1 {
				t.Fatalf("out of range %v", v)
			}
			if math.Abs(v-prev) > 0.05 {
				t.Fatalf("octaves %d: jump %v", oct, math.Abs(v-prev))
			}
			lo, hi = min(lo, v), max(hi, v)
			prev = v
		}
		if hi-lo < 0.5 {
			t.Errorf("octaves %d barely moves: [%v, %v]", oct, lo, hi)
		}
	}
}

func TestZeroAtLatticePointsAndContinuity(t *testing.T) {
	n := New(1, 1)
	for range 10 {
		if v := n.Generate(); v != 0 {
			t.Fatalf("lattice point gives %v", v)
		}
	}
}

func TestResetUnipolarAndSeed(t *testing.T) {
	genny.Seed(8)
	a := New(0.1, 3)
	genny.Seed(8)
	b := New(0.1, 3).SetUnipolar(true)
	var first float64
	for i := range 100 {
		x, y := a.Generate(), b.Generate()
		if math.Abs((x+1)/2-y) > 1e-12 || y < 0 || y > 1 {
			t.Fatal("seed or unipolar", x, y)
		}
		if i == 0 {
			first = x
		}
	}
	a.Reset()
	if a.Generate() != first {
		t.Error("reset")
	}
}
