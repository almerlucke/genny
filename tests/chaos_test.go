package tests

import (
	"math"
	"testing"

	"github.com/almerlucke/genny/float/iter"
	"github.com/almerlucke/genny/float/iter/updaters/chaos"
)

func TestAttractorsStayBoundedAndDoNotSettle(t *testing.T) {
	for _, c := range []struct {
		name         string
		it           *iter.Iterator
		limit        float64
		minRangeOfXY float64
	}{
		{"lorenz", iter.New([]float64{1, 1, 1}, chaos.NewLorenz(10, 28, 8.0/3, 0.01)), 60, 20},
		{"rossler", iter.New([]float64{1, 1, 1}, chaos.NewRossler(0.2, 0.2, 5.7, 0.05)), 40, 10},
	} {
		lo, hi := math.Inf(1), math.Inf(-1)
		for i := range 20000 {
			v := c.it.Generate()
			for _, x := range v {
				if math.IsNaN(x) || math.Abs(x) > c.limit {
					t.Fatalf("%s: %v at %d", c.name, v, i)
				}
			}
			if i > 1000 {
				lo, hi = math.Min(lo, v[0]), math.Max(hi, v[0])
			}
		}
		if hi-lo < c.minRangeOfXY {
			t.Errorf("%s: x only covers [%v, %v]", c.name, lo, hi)
		}
	}
}

func TestLorenzIsSensitive(t *testing.T) {
	a := iter.New([]float64{1, 1, 1}, chaos.NewLorenz(10, 28, 8.0/3, 0.01))
	b := iter.New([]float64{1, 1, 1.000001}, chaos.NewLorenz(10, 28, 8.0/3, 0.01))
	var diff float64
	for range 5000 {
		diff = math.Abs(a.Generate()[0] - b.Generate()[0])
	}
	if diff < 0.1 {
		t.Errorf("paths still together: %v", diff)
	}
}
