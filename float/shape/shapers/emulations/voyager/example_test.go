package voyager_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/emulations/voyager"
)

// Sawtooth emulates the sawtooth of the Minimoog Voyager from a phase.
func ExampleNewSawtooth() {
	s := voyager.NewSawtooth()
	for _, x := range []float64{0, 0.25, 0.5, 0.75} {
		fmt.Printf("%.3f -> %.3f\n", x, s.Shape(x))
	}
	// Output:
	// 0.000 -> -1.000
	// 0.250 -> -0.235
	// 0.500 -> 0.414
	// 0.750 -> 0.848
}
