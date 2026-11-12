package float_test

import (
	"fmt"

	"github.com/almerlucke/genny/float"
	"github.com/almerlucke/genny/float/line"
	"github.com/almerlucke/genny/sequence"
)

// ToFrame wraps a float generator as a one dimensional FrameGenerator, for
// APIs that work on frames (shape.Shape, interp.Interpolator, ...).
func ExampleToFrame() {
	f := float.ToFrame(sequence.New(1.0, 2.0, 3.0))
	fmt.Println(f.Dimensions(), f.Generate())
	// The frame slice is reused: copy it if you keep it past the next Generate.
	fmt.Println(f.Generate())
	// Output:
	// 1 [1]
	// [2]
}

// FromFrame picks one dimension out of a FrameGenerator.
func ExampleFromFrame() {
	l := line.New(0)
	l.To(1, 4)
	g := float.FromFrame(float.ToFrame(l), 0)
	for !g.Done() {
		fmt.Printf("%.2f ", g.Generate())
	}
	fmt.Println()
	// Output:
	// 0.00 0.25 0.50 0.75
}
