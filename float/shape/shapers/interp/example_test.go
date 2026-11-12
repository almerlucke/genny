package interp_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/add"
	"github.com/almerlucke/genny/float/shape/shapers/interp"
	"github.com/almerlucke/genny/float/shape/shapers/mult"
	"github.com/almerlucke/genny/float/shape/shapers/thru"
)

// Interpolate crossfades between neighbouring shapers; Index 0 is the first
// shaper and 1 the last.
func ExampleNew() {
	ip := interp.New(thru.New(), mult.New(2), add.New(10))
	for _, index := range []float64{0, 0.25, 0.5, 1} {
		ip.Index = index
		fmt.Printf("%.2f -> %.3f\n", index, ip.Shape(1))
	}
	// Output:
	// 0.00 -> 1.000
	// 0.25 -> 1.500
	// 0.50 -> 2.000
	// 1.00 -> 11.000
}
