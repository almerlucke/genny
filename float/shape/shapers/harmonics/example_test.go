package harmonics_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/harmonics"
)

func ExampleNew() {
	// a fundamental with a third of the second harmonic, scaled to a peak of 1
	h := harmonics.New([]float64{1, 1.0 / 3}, nil, true)
	for _, x := range []float64{-0.5, -0.25, 0, 0.25, 0.5} {
		fmt.Printf("%.4f ", h.Shape(x))
	}
	fmt.Println()
	// Output:
	// -0.8608 -0.8956 0.0000 0.8956 0.8608
}
