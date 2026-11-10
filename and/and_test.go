package and_test

import (
	"reflect"
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/and"
	"github.com/almerlucke/genny/constant"
	"github.com/almerlucke/genny/sequence"
)

func take[T any](g genny.Generator[T], n int) []T {
	out := make([]T, n)
	for i := range out {
		out[i] = g.Generate()
	}
	return out
}

func TestAndPlaysInOrderThenIsDone(t *testing.T) {
	a := and.New[int](sequence.New(1, 2), sequence.New(3), sequence.New(4, 5))
	if a.Continuous() || a.Done() {
		t.Fatal("fresh And must not be done or continuous")
	}
	var got []int
	for i := 0; i < 5; i++ {
		if a.Done() {
			t.Fatalf("done too early after %d values", i)
		}
		got = append(got, a.Generate())
	}
	if !reflect.DeepEqual(got, []int{1, 2, 3, 4, 5}) {
		t.Fatalf("got %v", got)
	}
	if !a.Done() {
		t.Fatal("should be done after the last generator")
	}
}

func TestAndLoop(t *testing.T) {
	a := and.NewLoop[int](sequence.New(1, 2), sequence.New(3))
	if !a.Continuous() {
		t.Fatal("loop must be continuous")
	}
	got := take[int](a, 7)
	if !reflect.DeepEqual(got, []int{1, 2, 3, 1, 2, 3, 1}) {
		t.Fatalf("got %v", got)
	}
	if a.Done() {
		t.Fatal("loop is never done")
	}
}

func TestAndContinuousGeneratorsTakeOneValueEach(t *testing.T) {
	a := and.New[int](constant.New(7), constant.New(8))
	got := take[int](a, 2)
	if !reflect.DeepEqual(got, []int{7, 8}) || !a.Done() {
		t.Fatalf("got %v done %v", got, a.Done())
	}
}

func TestAndReset(t *testing.T) {
	a := and.New[int](sequence.New(1, 2), sequence.New(3))
	take[int](a, 3)
	a.Reset()
	if a.Done() {
		t.Fatal("reset must clear done")
	}
	if got := take[int](a, 3); !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("got %v", got)
	}
	// reset in the middle
	a.Reset()
	a.Generate()
	a.Reset()
	if got := a.Generate(); got != 1 {
		t.Fatalf("got %d after mid reset", got)
	}
}

func TestAndEmpty(t *testing.T) {
	for _, a := range []*and.And[int]{and.New[int](), and.NewLoop[int]()} {
		if !a.Done() || a.Generate() != 0 {
			t.Fatal("empty And must be done and generate zero")
		}
		a.Reset()
		if !a.Done() || a.Generate() != 0 {
			t.Fatal("empty And must stay done after Reset")
		}
	}
}

func TestAndRandomizeKeepsAllGenerators(t *testing.T) {
	genny.Seed(1)
	a := and.New[int](constant.New(1), constant.New(2), constant.New(3))
	a.Randomize()
	got := take[int](a, 3)
	sum := got[0] + got[1] + got[2]
	if sum != 6 || got[0]*got[1]*got[2] != 6 {
		t.Fatalf("got %v", got)
	}
}
