package progression

import (
	"testing"

	"github.com/almerlucke/genny"
)

func TestTablesOnlyUseDeclaredMoves(t *testing.T) {
	genny.Seed(5)
	for name, table := range map[string]Table{"pop": Pop, "classical": Classical, "jazz": Jazz, "blues": Blues} {
		g := New(table, 0)
		prev := g.Generate()
		if prev != 0 {
			t.Fatalf("%s starts on %d", name, prev)
		}
		seen := map[int]bool{}
		for range 2000 {
			next := g.Generate()
			ok := false
			for _, m := range table[prev] {
				ok = ok || m.To == next
			}
			if !ok {
				t.Fatalf("%s: %d -> %d is not a move", name, prev, next)
			}
			seen[next] = true
			prev = next
		}
		if len(seen) < 3 {
			t.Errorf("%s only visits %v", name, seen)
		}
		if g.Done() {
			t.Errorf("%s ended", name)
		}
	}
}

func TestEndingDegree(t *testing.T) {
	g := New(Table{0: {{1, 1}}, 1: nil}, 0)
	got := []int{g.Generate(), g.Generate()}
	if got[0] != 0 || got[1] != 1 || !g.Done() {
		t.Errorf("%v done=%v", got, g.Done())
	}
}

func TestSeedRepeats(t *testing.T) {
	run := func() (out []int) {
		genny.Seed(11)
		g := New(Jazz, 1)
		for range 50 {
			out = append(out, g.Generate())
		}
		return
	}
	a, b := run(), run()
	for i := range a {
		if a[i] != b[i] {
			t.Fatal("not repeatable")
		}
	}
}
