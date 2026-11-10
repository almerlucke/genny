package continuous_test

import (
	"reflect"
	"testing"

	"github.com/almerlucke/genny/continuous"
	"github.com/almerlucke/genny/sequence"
)

func TestContinuousRestartsWhenDone(t *testing.T) {
	c := continuous.New[int](sequence.New(1, 2, 3))
	if !c.Continuous() {
		t.Fatal("must be continuous")
	}
	var got []int
	for i := 0; i < 8; i++ {
		got = append(got, c.Generate())
		if c.Done() {
			t.Fatal("never done")
		}
	}
	if want := []int{1, 2, 3, 1, 2, 3, 1, 2}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestContinuousReset(t *testing.T) {
	c := continuous.New[int](sequence.New(1, 2, 3))
	c.Generate()
	c.Reset()
	if c.Generate() != 1 {
		t.Fatal("Reset must reset the wrapped generator")
	}
}
