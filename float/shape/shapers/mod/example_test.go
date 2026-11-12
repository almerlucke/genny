package mod_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/mod"
)

// Mod is the floating point remainder (it keeps the sign of the input), so it wraps a phase.
func ExampleNew() {
	s := mod.New(1)
	for _, x := range []float64{0.25, 1.25, 2.75, -0.25} {
		fmt.Printf("%.3f -> %.3f\n", x, s.Shape(x))
	}
	// Output:
	// 0.250 -> 0.250
	// 1.250 -> 0.250
	// 2.750 -> 0.750
	// -0.250 -> -0.250
}
