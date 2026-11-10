package ngram

import (
	"slices"
	"testing"

	"github.com/almerlucke/genny"
)

func take[T comparable](c *Chain[T], n int) (out []T) {
	for range n {
		out = append(out, c.Generate())
	}
	return
}

func TestDeterministicExample(t *testing.T) {
	// every note has exactly one follower, so the chain repeats the example
	c := Learn([]string{"a", "b", "c", "d", "a", "b", "c", "d", "a"}, 1)
	got := take(c, 8)
	if !slices.Equal(got, []string{"b", "c", "d", "a", "b", "c", "d", "a"}) {
		t.Errorf("%v", got)
	}
	c.Reset()
	if take(c, 1)[0] != "b" {
		t.Error("reset")
	}
}

func TestOrderTwoRemembersMore(t *testing.T) {
	// after "a" comes b or c, but the two before it decide: x a -> b, y a -> c
	seq := []string{"x", "a", "b", "y", "a", "c", "x", "a", "b", "y", "a", "c"}
	c := Learn(seq, 2)
	genny.Seed(1)
	// started on [x a]: b, then y a c, then x a b ...
	got := take(c, 9)
	if !slices.Equal(got, []string{"b", "y", "a", "c", "x", "a", "b", "y", "a"}) {
		t.Errorf("%v", got)
	}
}

func TestOnlyObservedTransitionsAndStatistics(t *testing.T) {
	genny.Seed(3)
	seq := []int{1, 2, 1, 2, 1, 3, 1, 2, 1, 2}
	c := Learn(seq, 1)
	seen := map[[2]int]int{}
	prev := c.Generate()
	for range 20000 {
		v := c.Generate()
		seen[[2]int{prev, v}]++
		prev = v
	}
	for tr := range seen {
		ok := false
		for i := 0; i+1 < len(seq); i++ {
			ok = ok || (seq[i] == tr[0] && seq[i+1] == tr[1])
		}
		if !ok {
			t.Fatalf("transition %v never occurs in the example", tr)
		}
	}
	// 1 is followed by 2 four times and by 3 once
	n1 := seen[[2]int{1, 2}] + seen[[2]int{1, 3}]
	if r := float64(seen[[2]int{1, 3}]) / float64(n1); r < 0.17 || r > 0.23 {
		t.Errorf("1->3 ratio %v, want 0.2", r)
	}
}

func TestDeadEndRestarts(t *testing.T) {
	// 3 is only at the end: the chain cannot stop
	c := Learn([]int{1, 2, 3}, 1)
	for _, v := range take(c, 50) {
		if v < 1 || v > 3 {
			t.Fatal(v)
		}
	}
}

func TestInvalid(t *testing.T) {
	for name, f := range map[string]func(){
		"order 0":       func() { Learn([]int{1, 2, 3}, 0) },
		"short example": func() { Learn([]int{1, 2}, 2) },
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
