package series

import (
	"testing"

	"github.com/almerlucke/genny/float/shape/shapers/add"
	"github.com/almerlucke/genny/float/shape/shapers/mult"
)

func TestSeries(t *testing.T) {
	if got := New(mult.New(2), add.New(1)).Shape(3); got != 7 {
		t.Fatalf("got %v", got)
	}
	if got := New(add.New(1), mult.New(2)).Shape(3); got != 8 {
		t.Fatalf("order matters: got %v", got)
	}
	if got := New().Shape(3); got != 3 {
		t.Fatalf("empty series is the identity, got %v", got)
	}
}
