package deadzone_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/deadzone"
)

func ExampleNewDeadZone() {
	jump, cont := deadzone.NewDeadZone(0.2, deadzone.Jump), deadzone.NewDeadZone(0.2, deadzone.Continuous)
	for _, x := range []float64{0.1, 0.5, -0.5} {
		fmt.Printf("%.2f %.2f\n", jump.Shape(x), cont.Shape(x))
	}
	// Output:
	// 0.00 0.00
	// 0.50 0.30
	// -0.50 -0.30
}

func ExampleNewExpander() {
	e := deadzone.NewExpander(0.5, 2)
	for _, x := range []float64{0.1, 0.25, 0.5, 0.9} {
		fmt.Printf("%.4f ", e.Shape(x))
	}
	fmt.Println()
	// Output:
	// 0.0200 0.1250 0.5000 0.9000
}
