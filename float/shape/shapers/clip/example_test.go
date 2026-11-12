package clip_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/clip"
)

// Hard clipping limits values to a range.
func ExampleNew() {
	s := clip.New(clip.Hard, -0.5, 0.5)
	for _, x := range []float64{-1, -0.25, 0.25, 1} {
		fmt.Printf("%.3f -> %.3f\n", x, s.Shape(x))
	}
	// Output:
	// -1.000 -> -0.500
	// -0.250 -> -0.250
	// 0.250 -> 0.250
	// 1.000 -> 0.500
}
