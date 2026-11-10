package combine_test

import (
	"reflect"
	"testing"

	"github.com/almerlucke/genny/combine"
	"github.com/almerlucke/genny/constant"
	"github.com/almerlucke/genny/sequence"
)

func TestCombine(t *testing.T) {
	c := combine.New[int](sequence.New(1, 2), sequence.NewLoop(10, 20, 30), constant.New(5))
	if c.Continuous() {
		t.Fatal("not continuous when one generator is not")
	}
	var got [][]int
	for i := 0; i < 3; i++ {
		if c.Done() && i < 2 {
			t.Fatalf("done too early at %d", i)
		}
		got = append(got, c.Generate())
	}
	want := [][]int{{1, 10, 5}, {2, 20, 5}, {2, 30, 5}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	if !c.Done() {
		t.Fatal("done as soon as one generator is done")
	}
	c.Reset()
	if c.Done() || !reflect.DeepEqual(c.Generate(), []int{1, 10, 5}) {
		t.Fatal("reset resets all generators")
	}
}

func TestCombineContinuous(t *testing.T) {
	c := combine.New[int](sequence.NewLoop(1), constant.New(2))
	if !c.Continuous() || c.Done() {
		t.Fatal("all continuous")
	}
}

func TestCombineEmpty(t *testing.T) {
	c := combine.New[int]()
	if got := c.Generate(); len(got) != 0 {
		t.Fatalf("got %v", got)
	}
	if c.Done() || !c.Continuous() {
		t.Fatal("nothing to end: not done, continuous")
	}
	c.Reset()
}
