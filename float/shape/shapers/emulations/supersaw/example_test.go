package supersaw_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/emulations/supersaw"
)

// SuperSaw approximates the JP-8000 supersaw from a phase.
func ExampleNew() {
	s := supersaw.New(2, 1, 0.5)
	for _, x := range []float64{0, 0.25, 0.5, 0.75} {
		fmt.Printf("%.3f -> %.3f\n", x, s.Shape(x))
	}
	// Output:
	// 0.000 -> -1.000
	// 0.250 -> -0.041
	// 0.500 -> -1.000
	// 0.750 -> -0.041
}
