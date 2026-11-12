package mult_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/mult"
)

// Mult scales every value by a constant.
func ExampleNew() {
	s := mult.New(0.5)
	for _, x := range []float64{-1, 0, 0.5, 1} {
		fmt.Printf("%.3f -> %.3f\n", x, s.Shape(x))
	}
	// Output:
	// -1.000 -> -0.500
	// 0.000 -> 0.000
	// 0.500 -> 0.250
	// 1.000 -> 0.500
}
