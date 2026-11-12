package shape_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/phasor"
	"github.com/almerlucke/genny/float/shape"
	"github.com/almerlucke/genny/float/shape/shapers/function"
	"github.com/almerlucke/genny/float/shape/shapers/linear"
	"github.com/almerlucke/genny/float/shape/shapers/series"
)

// A phasor shaped into a triangle wave in [-1, 1]; Single works on a plain
// float generator, Shape on frames.
func ExampleNewSingle() {
	tri := series.New(function.NewTri(), linear.NewBipolar())
	g := shape.NewSingle(phasor.New(1, 8, 0), tri)
	for i := 0; i < 8; i++ {
		fmt.Printf("%.2f ", g.Generate())
	}
	fmt.Println()
	// Output:
	// -1.00 -0.50 0.00 0.50 1.00 0.50 0.00 -0.50
}
