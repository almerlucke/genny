package window_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/window"
)

// Window a block of 5 samples with a Hann window (symmetric, so the input is n/(N-1))
func ExampleHann() {
	w := window.Hann()
	for n := 0; n < 5; n++ {
		fmt.Printf("%.2f ", w.Shape(float64(n)/4))
	}
	fmt.Println()
	// Output:
	// 0.00 0.50 1.00 0.50 0.00
}

// A periodic window is read with n/N and wraps around
func ExampleWindow_Periodic() {
	w := window.Hann().Periodic()
	for n := 0; n < 4; n++ {
		fmt.Printf("%.2f ", w.Shape(float64(n)/4))
	}
	fmt.Printf("%.2f\n", w.Shape(1.25))
	// Output:
	// 0.00 0.50 1.00 0.50 0.50
}

func ExampleKaiser() {
	w := window.Kaiser(5)
	fmt.Printf("%.4f %.4f %.4f\n", w.Shape(0), w.Shape(0.25), w.Shape(0.5))
	// Output:
	// 0.0367 0.5529 1.0000
}
