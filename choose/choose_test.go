package choose

import (
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/constant"
	"github.com/almerlucke/genny/sequence"
)

func TestWeights(t *testing.T) {
	genny.Seed(9)
	w := New(Option[int]{constant.New(1), 3}, Option[int]{constant.New(2), 1}, Option[int]{constant.New(3), 0})
	count := map[int]int{}
	for range 20000 {
		count[w.Generate()]++
	}
	if count[3] != 0 {
		t.Error("zero weight picked")
	}
	if r := float64(count[1]) / 20000; r < 0.72 || r > 0.78 {
		t.Errorf("ratio %v", r)
	}
	if !w.Continuous() || w.Done() {
		t.Error("flags")
	}
}

func TestOnlyPickedAdvances(t *testing.T) {
	genny.Seed(1)
	a := sequence.NewLoop(1, 2, 3)
	w := New(Option[int]{a, 1}, Option[int]{constant.New(0), 1})
	want := 1
	for range 200 {
		if v := w.Generate(); v != 0 {
			if v != want {
				t.Fatalf("got %d, want %d", v, want)
			}
			want = want%3 + 1
		}
	}
}

func TestDoneOptionsAreSkipped(t *testing.T) {
	w := New(Option[int]{sequence.New(1), 1}, Option[int]{sequence.New(2, 2), 1})
	var got []int
	for !w.Done() {
		got = append(got, w.Generate())
	}
	if len(got) != 3 {
		t.Errorf("%v", got)
	}
	w.Reset()
	if w.Done() {
		t.Error("reset")
	}
}
