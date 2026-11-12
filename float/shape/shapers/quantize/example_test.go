package quantize_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/quantize"
)

// Quantizer rounds to a multiple of its quantum.
func ExampleNew() {
	s := quantize.New(0.25)
	for _, x := range []float64{0.1, 0.2, 0.4, 0.9} {
		fmt.Printf("%.3f -> %.3f\n", x, s.Shape(x))
	}
	// Output:
	// 0.100 -> 0.000
	// 0.200 -> 0.250
	// 0.400 -> 0.500
	// 0.900 -> 1.000
}
