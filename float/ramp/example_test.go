package ramp_test

import (
	"fmt"
	"strings"

	"github.com/almerlucke/genny/float/ramp"
)

// A ramp runs from min to max in the given number of steps, including both
// ends; exp curves it (1 is linear).
func ExampleNew() {
	for _, exp := range []float64{1, 2} {
		r := ramp.New(5, 0, 1, exp)
		var out []string
		for !r.Done() {
			out = append(out, fmt.Sprintf("%.3f", r.Generate()))
		}
		fmt.Println(strings.Join(out, " "))
	}
	// Output:
	// 0.000 0.250 0.500 0.750 1.000
	// 0.000 0.062 0.250 0.562 1.000
}

// NewForever stays on the end value instead of finishing.
func ExampleNewForever() {
	r := ramp.NewForever(3, 0, 10, 1)
	for i := 0; i < 5; i++ {
		fmt.Printf("%.0f ", r.Generate())
	}
	fmt.Println(r.Continuous(), r.Done())
	// Output:
	// 0 5 10 10 10 true false
}
