package lfo_test

import (
	"fmt"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/float/lfo"
)

// A 1 Hz triangle LFO, sampled 8 times per second
func ExampleLFO() {
	l := lfo.New(lfo.Triangle, 1, 8)
	for range 8 {
		fmt.Printf("%.2f ", l.Generate())
	}
	fmt.Println()
	// Output: -1.00 -0.50 0.00 0.50 1.00 0.50 0.00 -0.50
}

// Sample and hold with a slew limiter gives a smooth random pitch change instead of steps
func ExampleSlew() {
	genny.Seed(1)
	sh := lfo.New(lfo.SampleHold, 1, 8)
	smooth := lfo.NewSlew(sh, 0.25, 0.25)
	for range 16 {
		fmt.Printf("%.2f ", smooth.Generate())
	}
	fmt.Println()
	// Output: -0.52 -0.52 -0.52 -0.52 -0.52 -0.52 -0.52 -0.52 -0.27 -0.02 0.23 0.48 0.73 0.94 0.94 0.94
}
