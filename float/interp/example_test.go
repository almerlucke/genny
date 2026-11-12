package interp_test

import (
	"fmt"

	"github.com/almerlucke/genny/float"
	"github.com/almerlucke/genny/float/interp"
	"github.com/almerlucke/genny/sequence"
)

// Interpolator reads a source at a different speed: dt = 0.5 gives two output
// frames per source frame, with linear interpolation in between.
func ExampleNew() {
	src := float.ToFrame(sequence.NewLoop(0.0, 1.0, 0.0, -1.0))
	ip := interp.New(src, interp.Linear, 0.5)
	for i := 0; i < 8; i++ {
		fmt.Printf("%.2f ", ip.Generate()[0])
	}
	fmt.Println()
	// Output:
	// 0.00 0.50 1.00 0.50 0.00 -0.50 -1.00 -0.50
}
