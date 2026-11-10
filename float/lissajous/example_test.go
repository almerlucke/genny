package lissajous_test

import (
	"fmt"
	"math"

	"github.com/almerlucke/genny/float/lissajous"
)

// round to 2 decimals, and no negative zero
func round(v float64) float64 { return math.Round(v*100)/100 + 0 }

// Two related modulation sources: x and y as a 3:2 knot, 12 steps per period
func ExampleLissajous() {
	l := lissajous.New(3, 2, 0, 12)
	for range 6 {
		v := l.Generate()
		fmt.Printf("%5.2f %5.2f\n", round(v[0]), round(v[1]))
	}
	// Output:
	//  0.00  0.00
	//  1.00  0.87
	//  0.00  0.87
	// -1.00  0.00
	//  0.00 -0.87
	//  1.00 -0.87
}
