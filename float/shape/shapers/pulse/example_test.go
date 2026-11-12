package pulse_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/pulse"
)

// Pulse turns a phase in [0, 1) into a pulse wave of width W.
func ExampleNew() {
	s := pulse.New(0.25)
	for _, x := range []float64{0, 0.1, 0.3, 0.9} {
		fmt.Printf("%.3f -> %.3f\n", x, s.Shape(x))
	}
	// Output:
	// 0.000 -> 1.000
	// 0.100 -> 1.000
	// 0.300 -> 0.000
	// 0.900 -> 0.000
}
