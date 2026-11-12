package onepole_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/onepole"
)

// The step response of a lowpass with coefficient 0.5 halves the distance to the target each call
func ExampleNew() {
	f := onepole.New(0.5)
	for i := 0; i < 4; i++ {
		fmt.Printf("%.4f ", f.Shape(1))
	}
	fmt.Println()
	// Output:
	// 0.5000 0.7500 0.8750 0.9375
}

// A 1 ms time constant at 1 kHz is a single sample: 63% after one call
func ExampleNewTime() {
	f := onepole.NewTime(0.001, 1000)
	fmt.Printf("%.3f\n", f.Shape(1))
	// Output:
	// 0.632
}

// The highpass removes the constant part
func ExampleNewHighpass() {
	f := onepole.NewHighpass(0.5)
	for i := 0; i < 4; i++ {
		fmt.Printf("%.4f ", f.Shape(1))
	}
	fmt.Println()
	// Output:
	// 0.5000 0.2500 0.1250 0.0625
}
