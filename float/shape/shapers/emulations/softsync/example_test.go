package softsync_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/emulations/softsync"
)

// Triangle turns a phase into a soft synced triangle.
func ExampleNewTriangle() {
	s := softsync.NewTriangle(2)
	for _, x := range []float64{0.1, 0.3, 0.6, 0.9} {
		fmt.Printf("%.3f -> %.3f\n", x, s.Shape(x))
	}
	// Output:
	// 0.100 -> 0.600
	// 0.300 -> -0.200
	// 0.600 -> 0.600
	// 0.900 -> 0.600
}
