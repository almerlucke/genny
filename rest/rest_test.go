package rest

import (
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/clock"
	"github.com/almerlucke/genny/sequence"
)

func TestGated(t *testing.T) {
	g := NewGated[int](sequence.NewLoop(1, 2, 3, 4), clock.NewDivider(2, 0))
	var got []Optional[int]
	for range 4 {
		got = append(got, g.Generate())
	}
	want := []Optional[int]{Some(1), None[int](), Some(3), None[int]()}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("consume: %v, want %v", got, want)
		}
	}

	g = NewGated[int](sequence.NewLoop(1, 2, 3, 4), clock.NewDivider(2, 0)).Consume(false)
	got = got[:0]
	for range 4 {
		got = append(got, g.Generate())
	}
	want = []Optional[int]{Some(1), None[int](), Some(2), None[int]()}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("no consume: %v, want %v", got, want)
		}
	}
}

func TestProbability(t *testing.T) {
	genny.Seed(1)
	for _, c := range []struct{ p, lo, hi float64 }{{0, 0, 0}, {1, 1, 1}, {0.3, 0.25, 0.35}} {
		g := NewProbability[int](sequence.NewLoop(1, 2, 3), c.p)
		played := 0
		const n = 10000
		for range n {
			if g.Generate().Played() {
				played++
			}
		}
		if f := float64(played) / n; f < c.lo || f > c.hi {
			t.Errorf("p=%v played %v", c.p, f)
		}
	}
}
