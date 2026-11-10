package dist_test

import (
	"fmt"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/float/dist"
)

// Velocities that cluster around a mean are more natural than uniform random ones
func ExampleGaussian() {
	genny.Seed(1)
	velocity := dist.NewGaussian(0.7, 0.1).Clamp(0.1, 1)
	for range 8 {
		fmt.Printf("%.2f ", velocity.Generate())
	}
	fmt.Println()
	// Output: 0.77 0.53 0.66 0.53 0.38 0.64 0.73 0.84
}

// Many short values and a few long ones, like the gaps between random events
func ExampleExponential() {
	genny.Seed(1)
	gap := dist.NewExponential(1)
	for range 8 {
		fmt.Printf("%.2f ", gap.Generate())
	}
	fmt.Println()
	// Output: 0.27 3.58 1.52 0.75 0.71 0.36 1.54 0.80
}
