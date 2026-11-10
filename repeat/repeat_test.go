package repeat_test

import (
	"reflect"
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/constant"
	"github.com/almerlucke/genny/repeat"
	"github.com/almerlucke/genny/sequence"
)

func take(r *repeat.Repeat[int], n int) []int {
	out := make([]int, n)
	for i := range out {
		out[i] = r.Generate()
	}
	return out
}

func TestRepeatNonContinuousGenerator(t *testing.T) {
	r := repeat.New[int](sequence.New(1, 2), 2)
	if r.Continuous() || r.Done() {
		t.Fatal("repeat is never continuous, fresh is not done")
	}
	got := take(r, 4)
	if !reflect.DeepEqual(got, []int{1, 2, 1, 2}) {
		t.Fatalf("got %v", got)
	}
	if !r.Done() {
		t.Fatal("done after n passes")
	}
	if got := take(r, 2); !reflect.DeepEqual(got, []int{2, 2}) {
		t.Fatalf("done repeat returns the last value, got %v", got)
	}
}

func TestRepeatContinuousGenerator(t *testing.T) {
	r := repeat.New[int](constant.New(3), 3)
	for i := 0; i < 3; i++ {
		if r.Done() {
			t.Fatalf("done at %d", i)
		}
		if r.Generate() != 3 {
			t.Fatal("value")
		}
	}
	if !r.Done() {
		t.Fatal("done after 3 values")
	}
}

func TestRepeatReset(t *testing.T) {
	r := repeat.New[int](sequence.New(1, 2), 1)
	take(r, 2)
	if !r.Done() {
		t.Fatal("done")
	}
	r.Reset()
	if r.Done() {
		t.Fatal("reset clears done")
	}
	if got := take(r, 2); !reflect.DeepEqual(got, []int{1, 2}) || !r.Done() {
		t.Fatalf("got %v", got)
	}
}

func TestRepeatWithFunc(t *testing.T) {
	counts := []int{1, 2}
	i := 0
	r := repeat.NewWithFunc[int](constant.New(9), func() int { c := counts[i%2]; i++; return c })
	// n is read at construction (1) and again at every Reset (2)
	take(r, 1)
	if !r.Done() {
		t.Fatal("first n is 1")
	}
	r.Reset()
	take(r, 1)
	if r.Done() {
		t.Fatal("second n is 2")
	}
	take(r, 1)
	if !r.Done() {
		t.Fatal("done after 2")
	}
}

func TestRepeatNoRepeats(t *testing.T) {
	for _, n := range []int{0, -1, -5} {
		r := repeat.New[int](sequence.New(1, 2), n)
		if !r.Done() {
			t.Fatalf("n=%d must be done at once", n)
		}
		if r.Generate() != 0 {
			t.Fatalf("n=%d must not generate", n)
		}
	}
}

func TestRepeatRand(t *testing.T) {
	genny.Seed(10)
	seen := map[int]bool{}
	for i := 0; i < 100; i++ {
		r := repeat.NewRand[int](constant.New(1), 2, 4)
		n := 0
		for !r.Done() {
			r.Generate()
			n++
		}
		if n < 2 || n > 4 {
			t.Fatalf("n=%d out of [2,4]", n)
		}
		seen[n] = true
	}
	if len(seen) != 3 {
		t.Fatalf("expected 2, 3 and 4 to occur, got %v", seen)
	}
	// reversed bounds must not panic
	r := repeat.NewRand[int](constant.New(1), 3, 1)
	if r.Done() {
		t.Fatal("n must be in [1,3]")
	}
	// equal bounds
	r = repeat.NewRand[int](constant.New(1), 2, 2)
	r.Generate()
	r.Generate()
	if !r.Done() {
		t.Fatal("exactly 2")
	}
}
