package sequence

import (
	"reflect"
	"sort"
	"testing"

	"github.com/almerlucke/genny"
)

func take(s *Sequence[int], n int) []int {
	out := make([]int, n)
	for i := range out {
		out[i] = s.Generate()
	}
	return out
}

func TestSequenceOnce(t *testing.T) {
	s := New(1, 2, 3)
	if s.Continuous() || s.Done() {
		t.Fatal("fresh sequence")
	}
	for i, want := range []int{1, 2, 3} {
		if s.Done() {
			t.Fatalf("done before value %d", i)
		}
		if got := s.Generate(); got != want {
			t.Fatalf("value %d: got %d want %d", i, got, want)
		}
	}
	if !s.Done() {
		t.Fatal("done after last value")
	}
	if got := take(s, 2); !reflect.DeepEqual(got, []int{3, 3}) {
		t.Fatalf("done sequence repeats the last value, got %v", got)
	}
	s.Reset()
	if s.Done() || s.Generate() != 1 {
		t.Fatal("reset starts over")
	}
}

func TestSequenceLoop(t *testing.T) {
	s := NewLoop("a", "b")
	if !s.Continuous() {
		t.Fatal("loop is continuous")
	}
	var got []string
	for i := 0; i < 5; i++ {
		got = append(got, s.Generate())
	}
	if !reflect.DeepEqual(got, []string{"a", "b", "a", "b", "a"}) || s.Done() {
		t.Fatalf("got %v done %v", got, s.Done())
	}
}

func TestSingleValue(t *testing.T) {
	s := New(4)
	if s.Generate() != 4 || !s.Done() {
		t.Fatal("single value sequence is done after one value")
	}
}

func TestEmpty(t *testing.T) {
	for _, s := range []*Sequence[int]{New[int](), NewLoop[int]()} {
		if !s.Done() || s.Generate() != 0 {
			t.Fatal("empty sequence is done and generates zero")
		}
		s.Reset()
		if !s.Done() || s.Generate() != 0 {
			t.Fatal("empty sequence stays done after Reset")
		}
		s.Randomize()
	}
}

func TestRandomize(t *testing.T) {
	run := func() []int {
		genny.Seed(11)
		s := New(1, 2, 3, 4, 5, 6, 7, 8)
		s.Randomize()
		return take(s, 8)
	}
	a, b := run(), run()
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("same seed, different order: %v %v", a, b)
	}
	c := append([]int(nil), a...)
	sort.Ints(c)
	if !reflect.DeepEqual(c, []int{1, 2, 3, 4, 5, 6, 7, 8}) {
		t.Fatalf("not a permutation: %v", a)
	}
}
