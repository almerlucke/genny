package function_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/function"
)

// Function wraps a Go function; NewAbs, NewSin, NewTri and NewMod are ready made.
func ExampleNew() {
	s := function.NewAbs()
	for _, x := range []float64{-2, -0.5, 1} {
		fmt.Printf("%.3f -> %.3f\n", x, s.Shape(x))
	}
	// Output:
	// -2.000 -> 2.000
	// -0.500 -> 0.500
	// 1.000 -> 1.000
}
