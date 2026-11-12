package thru_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/thru"
)

// Thru passes its input unchanged, a placeholder in a chain.
func ExampleNew() {
	s := thru.New()
	for _, x := range []float64{-1, 0.3, 2} {
		fmt.Printf("%.3f -> %.3f\n", x, s.Shape(x))
	}
	// Output:
	// -1.000 -> -1.000
	// 0.300 -> 0.300
	// 2.000 -> 2.000
}
