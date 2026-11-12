package series_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/function"
	"github.com/almerlucke/genny/float/shape/shapers/linear"
	"github.com/almerlucke/genny/float/shape/shapers/series"
)

// Series chains shapers: here a phase is folded into a triangle in [-1, 1].
func ExampleNew() {
	s := series.New(function.NewTri(), linear.NewBipolar())
	for _, phase := range []float64{0, 0.25, 0.5, 0.75} {
		fmt.Printf("%.2f -> %.2f\n", phase, s.Shape(phase))
	}
	// Output:
	// 0.00 -> -1.00
	// 0.25 -> 0.00
	// 0.50 -> 1.00
	// 0.75 -> 0.00
}
