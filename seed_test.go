package genny_test

import (
	"slices"
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/bucket"
	"github.com/almerlucke/genny/float/rand"
	"github.com/almerlucke/genny/float/rwalk"
)

// draw takes values of generators that use the random stream
func draw() []float64 {
	r := rand.New(0, 1)
	w := rwalk.New(0.5, 0, 1, 0.1, rwalk.Bounce)
	b := bucket.NewLoop(1.0, 2.0, 3.0, 4.0).SetMode(bucket.Indexed)
	var out []float64
	for range 16 {
		out = append(out, r.Generate(), w.Generate(), b.Generate())
	}
	return out
}

func TestSeedRepeats(t *testing.T) {
	genny.Seed(7)
	a := draw()
	genny.Seed(7)
	b := draw()
	if !slices.Equal(a, b) {
		t.Fatal("the same seed gave other values")
	}
	genny.Seed(8)
	if slices.Equal(a, draw()) {
		t.Fatal("another seed gave the same values")
	}
}
