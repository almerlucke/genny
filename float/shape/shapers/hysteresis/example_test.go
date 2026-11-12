package hysteresis_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/hysteresis"
)

// A noisy signal around 0.5 does not make the trigger flutter
func ExampleSchmitt() {
	s := hysteresis.NewSchmitt(0.3, 0.7, 0, 1)
	for _, x := range []float64{0.1, 0.5, 0.8, 0.6, 0.4, 0.5, 0.2, 0.5} {
		fmt.Printf("%.0f ", s.Shape(x))
	}
	fmt.Println()
	// Output:
	// 0 0 1 1 1 1 0 0
}

// The output only follows after the input has left the dead band
func ExamplePlay() {
	p := hysteresis.NewPlay(1)
	for _, x := range []float64{0, 0.4, 1, 2, 1.5, 1, 0} {
		fmt.Printf("%.1f ", p.Shape(x))
	}
	fmt.Println()
	// Output:
	// 0.0 0.0 0.5 1.5 1.5 1.5 0.5
}
