package ripple_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/ripple"
)

// Ripple adds the remainder x mod M to x, giving a staircase-like ripple.
func ExampleNew() {
	s := ripple.New(0.5)
	for _, x := range []float64{0.2, 0.6, 1.3} {
		fmt.Printf("%.3f -> %.3f\n", x, s.Shape(x))
	}
	// Output:
	// 0.200 -> 0.400
	// 0.600 -> 0.700
	// 1.300 -> 1.600
}
