package phasedist_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/phasedist"
)

// A phase distorted saw falls quickly and rises slowly, d=0 would be a cosine
func ExampleNewSaw() {
	s := phasedist.NewSaw(1)
	for _, p := range []float64{0, 0.005, 0.25, 0.5, 0.75, 1} {
		fmt.Printf("%.3f ", s.Shape(p))
	}
	fmt.Println()
	// Output:
	// 1.000 -1.000 -0.715 -0.008 0.704 1.000
}

// The square rests at -1 and 1
func ExampleNewSquare() {
	s := phasedist.NewSquare(1)
	for _, p := range []float64{0.1, 0.4, 0.6, 0.9} {
		fmt.Printf("%.1f ", s.Shape(p))
	}
	fmt.Println()
	// Output:
	// -1.0 -1.0 1.0 1.0
}

func ExampleNewResonant() {
	s := phasedist.NewResonant(1, 4)
	for _, p := range []float64{0, 0.125, 0.25, 0.5, 1} {
		fmt.Printf("%.3f ", s.Shape(p))
	}
	fmt.Println()
	// Output:
	// 1.000 -0.875 0.750 0.500 1.000
}
