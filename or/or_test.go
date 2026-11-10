package or_test

import (
	"reflect"
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/constant"
	"github.com/almerlucke/genny/or"
	"github.com/almerlucke/genny/sequence"
)

func take(o *or.Or[int], n int) []int {
	out := make([]int, n)
	for i := range out {
		out[i] = o.Generate()
	}
	return out
}

func TestIndexedRunsEachGeneratorOnce(t *testing.T) {
	genny.Seed(1)
	o := or.New(or.Indexed, sequence.New(1, 2), sequence.New(10, 20), sequence.New(100))
	if o.Continuous() || o.Done() {
		t.Fatal("fresh indexed Or")
	}
	got := take(o, 5)
	if !o.Done() {
		t.Fatal("done after every generator ran")
	}
	// each sequence is played in full, in some order
	sum := 0
	for _, v := range got {
		sum += v
	}
	if sum != 133 {
		t.Fatalf("got %v", got)
	}
	// 1 must be directly followed by 2, 10 by 20
	for i, v := range got {
		if v == 1 && (i+1 >= len(got) || got[i+1] != 2) {
			t.Fatalf("sequence broken: %v", got)
		}
		if v == 10 && (i+1 >= len(got) || got[i+1] != 20) {
			t.Fatalf("sequence broken: %v", got)
		}
	}
}

func TestIndexedSeedRepeatableAndReset(t *testing.T) {
	run := func() []int {
		genny.Seed(8)
		return take(or.New(or.Indexed, sequence.New(1, 2), sequence.New(3), sequence.New(4, 5)), 5)
	}
	if a, b := run(), run(); !reflect.DeepEqual(a, b) {
		t.Fatalf("%v != %v", a, b)
	}

	genny.Seed(8)
	o := or.New(or.Indexed, sequence.New(1, 2), sequence.New(3))
	take(o, 3)
	if !o.Done() {
		t.Fatal("should be done")
	}
	o.Reset()
	if o.Done() {
		t.Fatal("reset clears done")
	}
	got := take(o, 3)
	if !o.Done() || (got[0] != 3 && got[0] != 1) {
		t.Fatalf("got %v done %v", got, o.Done())
	}
}

func TestLoopNeverDone(t *testing.T) {
	genny.Seed(2)
	o := or.NewLoop(or.Indexed, sequence.New(1), sequence.New(2))
	if !o.Continuous() {
		t.Fatal("loop is continuous")
	}
	got := take(o, 8)
	for i := 0; i+1 < len(got); i += 2 {
		if got[i]+got[i+1] != 3 {
			t.Fatalf("each round plays both generators once: %v", got)
		}
	}
	if o.Done() {
		t.Fatal("loop is never done")
	}
}

func TestRandomMode(t *testing.T) {
	genny.Seed(3)
	o := or.New(or.Random, constant.New(1), constant.New(2))
	if !o.Continuous() {
		t.Fatal("random Or is continuous")
	}
	seen := map[int]bool{}
	for _, v := range take(o, 100) {
		seen[v] = true
	}
	if !seen[1] || !seen[2] || len(seen) != 2 || o.Done() {
		t.Fatalf("seen %v", seen)
	}

	run := func() []int {
		genny.Seed(77)
		return take(or.New(or.Random, constant.New(1), constant.New(2), constant.New(3)), 30)
	}
	if a, b := run(), run(); !reflect.DeepEqual(a, b) {
		t.Fatal("same seed must repeat")
	}
}

func TestEmpty(t *testing.T) {
	for _, o := range []*or.Or[int]{or.New[int](or.Indexed), or.NewLoop[int](or.Random), or.New[int](or.Random)} {
		if !o.Done() || o.Generate() != 0 {
			t.Fatal("empty Or is done and generates zero")
		}
		o.Reset()
		if !o.Done() || o.Generate() != 0 {
			t.Fatal("empty Or stays done after Reset")
		}
	}
}
