package bitcrush_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/bitcrush"
)

// Three bits: 8 levels in steps of 0.25
func ExampleQuantize() {
	for _, x := range []float64{-1, -0.3, 0.1, 0.4, 1} {
		fmt.Printf("%.2f ", bitcrush.Quantize(x, 3))
	}
	fmt.Println()
	// Output:
	// -1.00 -0.25 0.00 0.50 0.75
}

// Hold every value for 2 calls
func ExampleCrusher() {
	c := bitcrush.New(24, 2, 1)
	for _, x := range []float64{0.1, 0.2, 0.3, 0.4} {
		fmt.Printf("%.1f ", c.Shape(x))
	}
	fmt.Println()
	// Output:
	// 0.1 0.1 0.3 0.3
}
