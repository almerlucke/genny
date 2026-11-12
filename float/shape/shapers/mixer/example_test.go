package mixer_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape"
	"github.com/almerlucke/genny/float/shape/shapers/mixer"
	"github.com/almerlucke/genny/float/shape/shapers/mult"
	"github.com/almerlucke/genny/float/shape/shapers/thru"
)

// Mixer sums the output of its shapers, each scaled by a level (default 1).
func ExampleNew() {
	m := mixer.New(thru.New(), mult.New(3))
	fmt.Printf("%.3f\n", m.Shape(0.5))
	// Output:
	// 2.000
}

func ExampleNewWithLevels() {
	m := mixer.NewWithLevels([]shape.Shaper{thru.New(), mult.New(3)}, []float64{1, 0.5})
	fmt.Printf("%.3f\n", m.Shape(0.5))
	// Output:
	// 1.250
}
