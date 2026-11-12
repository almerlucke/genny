package saturate_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/saturate"
)

func ExampleNew() {
	s := saturate.New(saturate.Tanh, 3)
	for _, x := range []float64{0.1, 0.5, 1, 2} {
		fmt.Printf("%.4f ", s.Shape(x))
	}
	fmt.Println()
	// Output:
	// 0.2928 0.9096 1.0000 1.0050
}

func ExampleNewTube() {
	s := saturate.NewTube(2, 0.3)
	fmt.Printf("%.4f %.4f %.4f\n", s.Shape(-1), s.Shape(0), s.Shape(1))
	// Output:
	// -1.0000 0.0000 0.3178
}
