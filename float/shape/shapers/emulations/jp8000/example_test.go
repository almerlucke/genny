package jp8000_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/emulations/jp8000"
)

// TriMod emulates the triangle modulation waveform.
func ExampleNewTriMod() {
	s := jp8000.NewTriMod(0.5)
	for _, x := range []float64{0.1, 0.25, 0.4, 0.6} {
		fmt.Printf("%.3f -> %.3f\n", x, s.Shape(x))
	}
	// Output:
	// 0.100 -> 0.600
	// 0.250 -> 0.000
	// 0.400 -> -0.600
	// 0.600 -> -0.600
}
