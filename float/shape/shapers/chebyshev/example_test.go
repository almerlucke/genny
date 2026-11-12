package chebyshev_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/chebyshev"
)

// A Chebyshev shaper maps a sine of amplitude 1 to a mix of its harmonics (here the fundamental plus a third harmonic).
func ExampleNew() {
	s := chebyshev.New(map[int]float64{1: 1, 3: 0.5})
	for _, x := range []float64{-1, -0.5, 0, 0.5, 1} {
		fmt.Printf("%.3f -> %.3f\n", x, s.Shape(x))
	}
	// Output:
	// -1.000 -> -1.500
	// -0.500 -> 0.000
	// 0.000 -> 0.000
	// 0.500 -> 0.000
	// 1.000 -> 1.500
}
