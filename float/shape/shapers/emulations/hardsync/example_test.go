package hardsync_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/emulations/hardsync"
)

// HardSync turns a phase into a sawtooth that restarts a1 times per cycle.
func ExampleNew() {
	s := hardsync.New(2)
	for _, x := range []float64{0, 0.25, 0.5, 0.75} {
		fmt.Printf("%.3f -> %.3f\n", x, s.Shape(x))
	}
	// Output:
	// 0.000 -> -1.000
	// 0.250 -> 0.000
	// 0.500 -> -1.000
	// 0.750 -> 0.000
}
