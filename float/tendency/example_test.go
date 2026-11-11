package tendency_test

import (
	"fmt"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/float/tendency"
)

func show[T any](g interface{ Generate() T }, n int, format string) {
	for i := range n {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Printf(format, g.Generate())
	}
	fmt.Println()
}

// A band that opens from a single pitch to two octaves, and closes again
func ExampleValue() {
	genny.Seed(1)
	mask := tendency.NewMask(false,
		tendency.Segment{N: 6, A1: 60, A2: 60, Z1: 53, Z2: 77},
		tendency.Segment{N: 6, A1: 53, A2: 77, Z1: 65, Z2: 67},
	)
	g := tendency.NewValue(mask)
	show[float64](g, 12, "%.0f")
	fmt.Println(g.Done())
	// Output:
	// 60 63 65 63 64 60 72 66 73 66 64 66
	// true
}

// The bounds of a mask for a list are percentages of the list: the first half, then the second
func ExampleChoice() {
	genny.Seed(1)
	mask := tendency.NewMask(true,
		tendency.Segment{N: 4, A1: 0, A2: 50, Z1: 0, Z2: 50},
		tendency.Segment{N: 4, A1: 50, A2: 100, Z1: 50, Z2: 100},
	)
	g := tendency.NewChoice(mask, []string{"c", "d", "e", "f", "g", "a", "b", "C"})
	show[string](g, 8, "%s")
	// Output:
	// c f f e b a C b
}

// With a beta distribution below 1 the values cling to the bounds of the band
func ExampleNewBetaValue() {
	genny.Seed(1)
	mask := tendency.NewMask(true, tendency.Segment{N: 1, A1: 0, A2: 10, Z1: 0, Z2: 10})
	show[float64](tendency.NewBetaValue(mask, 0.3, 0.3), 10, "%.1f")
	// Output:
	// 7.8 0.0 9.8 1.3 1.8 6.1 8.4 9.9 9.9 9.5
}

// Values of the outer band, but not inside the band of the inner mask
func ExampleTwoMasks() {
	genny.Seed(1)
	outer := tendency.NewMask(false, tendency.Segment{N: 10, A1: 0, A2: 10, Z1: 0, Z2: 10})
	inner := tendency.NewMask(false, tendency.Segment{N: 10, A1: 3, A2: 7, Z1: 3, Z2: 7})
	show[float64](tendency.NewTwoMasks(outer, inner, 20), 10, "%.1f")
	// Output:
	// 2.4 9.7 7.8 7.9 10.0 1.6 0.7 9.0 9.6 1.9
}

// A rising arch of 8 values that starts at 60 and ends at 64, within 60 and 72
func ExampleAnchor() {
	for _, v := range tendency.Anchor([]float64{0, 1, 0.6}, 8, 60, 64, 60, 72) {
		fmt.Printf("%.1f ", v)
	}
	fmt.Println()
	// Output:
	// 60.0 63.0 65.9 68.9 69.5 67.7 65.8 64.0
}

// Spray scatters a line of values within a spread
func ExampleSpray() {
	genny.Seed(1)
	for _, v := range tendency.Spray([]float64{60, 62, 64, 65}, 1) {
		fmt.Printf("%.1f ", v)
	}
	fmt.Println()
	// Output:
	// 59.5 62.9 64.6 65.1
}
