package lookup_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/lookup"
)

// A Table is indexed by an input in [0, 1] and interpolates linearly between entries.
func ExampleNewSineTable() {
	t := lookup.NewSineTable(8)
	for _, x := range []float64{0, 0.125, 0.25, 0.5, 0.75} {
		fmt.Printf("%.3f -> %.3f\n", x, t.Shape(x))
	}
	// Output:
	// 0.000 -> 0.000
	// 0.125 -> 0.707
	// 0.250 -> 1.000
	// 0.500 -> 0.000
	// 0.750 -> -1.000
}

func ExampleNewWithRampFunc() {
	t := lookup.NewWithRampFunc(4, lookup.Custom, func(r float64) float64 { return r * r }, 1)
	fmt.Printf("%.3f %.3f %.3f\n", t.Shape(0), t.Shape(0.5), t.Shape(1))
	// Output:
	// 0.000 0.250 1.000
}
