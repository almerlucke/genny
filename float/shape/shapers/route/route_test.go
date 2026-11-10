package route

import (
	"testing"

	"github.com/almerlucke/genny/float/shape/shapers/mult"
)

func TestRoute(t *testing.T) {
	a, b := mult.New(2), mult.New(3)
	r := New(1, a, b)
	if got := r.Shape(2); got != 6 || r.Selected() != b {
		t.Fatalf("index 1: got %v", got)
	}
	r.Index = 0
	if got := r.Shape(2); got != 4 || r.Selected() != a {
		t.Fatalf("index 0: got %v", got)
	}
}
