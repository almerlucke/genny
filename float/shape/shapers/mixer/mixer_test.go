package mixer

import (
	"testing"

	"github.com/almerlucke/genny/float/shape"
	"github.com/almerlucke/genny/float/shape/shapers/mult"
	"github.com/almerlucke/genny/float/shape/shapers/thru"
)

func TestMixer(t *testing.T) {
	if got := New(thru.New(), mult.New(2)).Shape(3); got != 9 {
		t.Fatalf("unit levels: got %v", got)
	}
	m := NewWithLevels([]shape.Shaper{thru.New(), mult.New(2)}, []float64{0.5, 2})
	if got := m.Shape(3); got != 13.5 {
		t.Fatalf("levels: got %v", got)
	}
	m.Levels[0] = 0
	if got := m.Shape(3); got != 12 {
		t.Fatalf("changed level: got %v", got)
	}
	if got := New().Shape(3); got != 0 {
		t.Fatalf("empty mixer: got %v", got)
	}
}
