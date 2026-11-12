package rwalk_test

import (
	"fmt"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/float/rwalk"
)

// A random walk takes steps of at most maxPct and stays within [min, max].
func ExampleNew() {
	genny.Seed(11)
	w := rwalk.New(0.5, 0, 1, 0.2, rwalk.Bounce)
	lo, hi := 1.0, 0.0
	for i := 0; i < 1000; i++ {
		v := w.Generate()
		lo, hi = min(lo, v), max(hi, v)
	}
	fmt.Println(lo >= 0, hi <= 1)
	for i := 0; i < 4; i++ {
		fmt.Printf("%.3f ", w.Generate())
	}
	fmt.Println()
	// Output:
	// true true
	// 0.976 0.904 0.737 0.597
}
