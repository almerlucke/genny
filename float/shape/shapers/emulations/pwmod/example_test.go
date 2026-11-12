package pwmod_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/emulations/pwmod"
)

// PWMod turns a phase into a pulse wave in [-1, 1]; SetWidth changes the pulse width.
func ExampleNew() {
	s := pwmod.New()
	for _, x := range []float64{0, 0.2, 0.4, 0.6, 0.8} {
		fmt.Printf("%.3f -> %.3f\n", x, s.Shape(x))
	}
	// Output:
	// 0.000 -> 1.000
	// 0.200 -> 1.000
	// 0.400 -> -1.000
	// 0.600 -> -1.000
	// 0.800 -> 1.000
}
