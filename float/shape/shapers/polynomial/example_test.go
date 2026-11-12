package polynomial_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/polynomial"
)

func ExampleNew() {
	p := polynomial.New(1, 0, -2) // 1 - 2x^2
	fmt.Println(p.Shape(0), p.Shape(0.5), p.Shape(1))
	// Output:
	// 1 0.5 -1
}

func ExampleNewLegendre() {
	p := polynomial.NewLegendre(3) // (5x^3 - 3x) / 2
	fmt.Println(p.Coefficients(), p.Shape(1), p.Shape(0.5))
	// Output:
	// [0 -1.5 0 2.5] 1 -0.4375
}

func ExampleNewNormalized() {
	n := polynomial.NewNormalized(polynomial.New(0, 0, 4)) // 4x^2 peaks at 4
	fmt.Println(n.Gain(), n.Shape(1))
	// Output:
	// 0.25 1
}
