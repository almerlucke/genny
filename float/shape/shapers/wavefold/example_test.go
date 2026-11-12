package wavefold_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/wavefold"
)

func ExampleNewSine() {
	s := wavefold.NewSine(2)
	for _, x := range []float64{0.25, 0.5, 1} {
		fmt.Printf("%.4f ", s.Shape(x))
	}
	fmt.Println()
	// Output:
	// 0.7071 1.0000 0.0000
}

func ExampleNewTriangle() {
	s := wavefold.NewTriangle(2)
	for _, x := range []float64{0.25, 0.5, 0.75, 1, 1.5} {
		fmt.Printf("%.2f ", s.Shape(x))
	}
	fmt.Println()
	// Output:
	// 0.50 1.00 0.50 0.00 -1.00
}
