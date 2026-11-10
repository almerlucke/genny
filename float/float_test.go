package float_test

import (
	"reflect"
	"testing"

	"github.com/almerlucke/genny/float"
	"github.com/almerlucke/genny/float/iter"
	"github.com/almerlucke/genny/sequence"
)

func TestToFrame(t *testing.T) {
	f := float.ToFrame(sequence.New(1.0, 2.0))
	var _ float.FrameGenerator = f
	if f.Dimensions() != 1 || f.Continuous() || f.Done() {
		t.Fatal("one dimension, state follows the source")
	}
	if got := f.Generate(); !reflect.DeepEqual(got, []float64{1}) {
		t.Fatalf("got %v", got)
	}
	if got := f.Generate(); !reflect.DeepEqual(got, []float64{2}) {
		t.Fatalf("got %v", got)
	}
	if !f.Done() {
		t.Fatal("done follows the source")
	}
	f.Reset()
	if f.Done() || f.Generate()[0] != 1 {
		t.Fatal("reset follows the source")
	}
	if !float.ToFrame(sequence.NewLoop(1.0)).Continuous() {
		t.Fatal("continuous follows the source")
	}
}

type double struct{}

func (double) Update(v []float64) {
	for i := range v {
		v[i] *= 2
	}
}

func TestFromFrame(t *testing.T) {
	it := iter.New([]float64{1, 10}, double{})
	d0 := float.FromFrame(it, 0)
	if v := d0.Generate(); v != 1 {
		t.Fatalf("got %v", v)
	}
	// the iterator advanced once, dimension 1 now
	d1 := float.FromFrame(it, 1)
	if v := d1.Generate(); v != 20 {
		t.Fatalf("got %v", v)
	}
	if !d1.Continuous() || d1.Done() {
		t.Fatal("iterator is continuous")
	}
	d1.Reset()
	if v := d1.Generate(); v != 10 {
		t.Fatalf("reset must restart the iterator, got %v", v)
	}
}
