package mirror_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/mirror"
)

// Mirror reflects values that leave the range back into it.
func ExampleNew() {
	s := mirror.New(-1, 1)
	for _, x := range []float64{-1.5, -0.5, 0.5, 1.25} {
		fmt.Printf("%.3f -> %.3f\n", x, s.Shape(x))
	}
	// Output:
	// -1.500 -> -0.500
	// -0.500 -> -0.500
	// 0.500 -> 0.500
	// 1.250 -> 0.750
}
