package curve_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/curve"
)

func ExampleNewUnipolar() {
	c := curve.NewUnipolar(0.5) // exponent 1/2
	fmt.Println(c.Shape(0.25), c.Shape(1))
	// Output:
	// 0.5 1
}

func ExampleNewBipolar() {
	c := curve.NewBipolar(-0.5) // exponent 2
	fmt.Println(c.Shape(-0.5), c.Shape(0.5))
	// Output:
	// -0.25 0.25
}
