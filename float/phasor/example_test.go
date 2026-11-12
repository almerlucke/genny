package phasor_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/phasor"
)

// A phasor at 1 kHz with a sample rate of 8 kHz advances 1/8 per sample.
func ExampleNew() {
	p := phasor.New(1000, 8000, 0)
	for i := 0; i < 10; i++ {
		fmt.Printf("%.3f ", p.Generate())
	}
	fmt.Println()
	// Output:
	// 0.000 0.125 0.250 0.375 0.500 0.625 0.750 0.875 0.000 0.125
}
