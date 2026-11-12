package add_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/add"
)

// Add shifts every value by a constant.
func ExampleNew() {
	s := add.New(0.25)
	for _, x := range []float64{-1, 0, 0.5} {
		fmt.Printf("%.3f -> %.3f\n", x, s.Shape(x))
	}
	// Output:
	// -1.000 -> -0.750
	// 0.000 -> 0.250
	// 0.500 -> 0.750
}
