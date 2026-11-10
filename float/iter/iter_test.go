package iter_test

import (
	"reflect"
	"testing"

	"github.com/almerlucke/genny/float/iter"
)

type double struct{}

func (double) Update(v []float64) {
	for i := range v {
		v[i] *= 2
	}
}

func TestIterator(t *testing.T) {
	init := []float64{1, 3}
	it := iter.New(init, double{})
	if it.Dimensions() != 2 || !it.Continuous() || it.Done() {
		t.Fatal("two dimensional, continuous, never done")
	}
	for _, want := range [][]float64{{1, 3}, {2, 6}, {4, 12}} {
		if got := it.Generate(); !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v want %v", got, want)
		}
	}
	if !reflect.DeepEqual(init, []float64{1, 3}) {
		t.Fatalf("init values were modified: %v", init)
	}
	it.Reset()
	if got := it.Generate(); !reflect.DeepEqual(got, []float64{1, 3}) {
		t.Fatalf("after reset got %v", got)
	}
	it.SetValues([]float64{5, 6})
	if got := it.Generate(); !reflect.DeepEqual(got, []float64{5, 6}) {
		t.Fatalf("after SetValues got %v", got)
	}
	if got := it.Generate(); !reflect.DeepEqual(got, []float64{10, 12}) {
		t.Fatalf("got %v", got)
	}
}
