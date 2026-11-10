package groove_test

import (
	"fmt"

	"github.com/almerlucke/genny/groove"
)

// Offsets are fractions of a step: add them to the time of each step
func ExampleSwing() {
	swing := groove.NewSwing(1.0 / 3) // triplet feel
	for step := range 4 {
		fmt.Printf("step %d at %.3f\n", step, float64(step)+swing.Generate())
	}
	// Output:
	// step 0 at 0.000
	// step 1 at 1.333
	// step 2 at 2.000
	// step 3 at 3.333
}

// Velocities for a bar of 7/8 in groups of 2+2+3 eighths: the downbeat is strongest, the other
// groups get a medium accent
func ExampleAccent() {
	a := groove.NewAccent(2, 2, 3)
	for range 7 {
		fmt.Print(a.Generate(), " ")
	}
	fmt.Println()
	// Output: 1 0.55 0.8 0.55 0.8 0.55 0.55
}
