package bucket

import (
	"reflect"
	"sort"
	"testing"

	"github.com/almerlucke/genny"
)

func drain(b *Bucket[int], n int) []int {
	out := make([]int, n)
	for i := range out {
		out[i] = b.Generate()
	}
	return out
}

func sorted(v []int) []int {
	c := append([]int(nil), v...)
	sort.Ints(c)
	return c
}

func TestIndexedTakesEachValueOnce(t *testing.T) {
	genny.Seed(3)
	b := New(1, 2, 3, 4, 5)
	if b.Continuous() || b.Done() {
		t.Fatal("fresh indexed bucket")
	}
	got := drain(b, 5)
	if !reflect.DeepEqual(sorted(got), []int{1, 2, 3, 4, 5}) {
		t.Fatalf("not a permutation: %v", got)
	}
	if !b.Done() {
		t.Fatal("done after all values were taken")
	}
	// done: keeps returning the last value
	for i := 0; i < 3; i++ {
		if v := b.Generate(); v != got[4] {
			t.Fatalf("got %d want last value %d", v, got[4])
		}
	}
}

func TestIndexedLoop(t *testing.T) {
	genny.Seed(4)
	b := NewLoop(1, 2, 3)
	if !b.Continuous() {
		t.Fatal("loop is continuous")
	}
	for round := 0; round < 3; round++ {
		got := drain(b, 3)
		if !reflect.DeepEqual(sorted(got), []int{1, 2, 3}) {
			t.Fatalf("round %d not a permutation: %v", round, got)
		}
		if b.Done() {
			t.Fatal("loop is never done")
		}
	}
}

func TestResetReshuffles(t *testing.T) {
	genny.Seed(5)
	b := New(1, 2, 3, 4, 5, 6, 7, 8)
	drain(b, 8)
	b.Reset()
	if b.Done() {
		t.Fatal("reset clears done")
	}
	got := drain(b, 8)
	if !reflect.DeepEqual(sorted(got), []int{1, 2, 3, 4, 5, 6, 7, 8}) {
		t.Fatalf("not a permutation: %v", got)
	}
}

func TestSeedRepeatable(t *testing.T) {
	run := func() []int {
		genny.Seed(99)
		b := New(1, 2, 3, 4, 5, 6)
		return drain(b, 6)
	}
	if a, b := run(), run(); !reflect.DeepEqual(a, b) {
		t.Fatalf("%v != %v", a, b)
	}
	run2 := func() []int {
		genny.Seed(99)
		b := New(1, 2, 3, 4, 5, 6).SetMode(Random)
		return drain(b, 20)
	}
	if a, b := run2(), run2(); !reflect.DeepEqual(a, b) {
		t.Fatalf("%v != %v", a, b)
	}
}

func TestRandomMode(t *testing.T) {
	genny.Seed(6)
	b := New(1, 2, 3).SetMode(Random)
	if !b.Continuous() || b.Done() {
		t.Fatal("random mode is continuous and never done")
	}
	seen := map[int]bool{}
	for i := 0; i < 200; i++ {
		v := b.Generate()
		if v < 1 || v > 3 {
			t.Fatalf("value %d out of bucket", v)
		}
		seen[v] = true
	}
	if len(seen) != 3 {
		t.Fatalf("expected all values to be drawn, got %v", seen)
	}
}

func TestLoopFlagAndSetValues(t *testing.T) {
	b := New(1, 2).Loop(true)
	if !b.Continuous() {
		t.Fatal("Loop(true) makes it continuous")
	}
	b.Loop(false)
	if b.Continuous() {
		t.Fatal("Loop(false) makes it non-continuous")
	}
	drain(b, 2)
	b.SetValues([]int{9})
	if b.Done() || b.Generate() != 9 || !b.Done() {
		t.Fatal("SetValues restarts the bucket")
	}
	if !reflect.DeepEqual(b.Values(), []int{9}) {
		t.Fatalf("values %v", b.Values())
	}
	b.SetValues(nil)
	if !b.Done() || b.Generate() != 0 {
		t.Fatal("empty after SetValues(nil)")
	}
}

func TestEmpty(t *testing.T) {
	for name, b := range map[string]*Bucket[int]{
		"New":      New[int](),
		"NewLoop":  NewLoop[int](),
		"random":   New[int]().SetMode(Random),
		"indexed2": New[int]().SetMode(Indexed),
	} {
		t.Run(name, func(t *testing.T) {
			if !b.Done() && b.mode == Indexed {
				t.Fatal("empty bucket is done at once")
			}
			if b.Generate() != 0 {
				t.Fatal("empty bucket generates zero")
			}
			b.Reset()
			if b.mode == Indexed && !b.Done() {
				t.Fatal("empty bucket stays done after Reset")
			}
			if b.Generate() != 0 {
				t.Fatal("still zero")
			}
		})
	}
}
