package linear_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/linear"
)

// Linear computes x*Scale+Shift; NewBipolar maps [0, 1] to [-1, 1].
func ExampleNew() {
	s := linear.New(2, -1)
	for _, x := range []float64{0, 0.25, 0.5, 1} {
		fmt.Printf("%.3f -> %.3f\n", x, s.Shape(x))
	}
	// Output:
	// 0.000 -> -1.000
	// 0.250 -> -0.500
	// 0.500 -> 0.000
	// 1.000 -> 1.000
}
