package rectify_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/rectify"
)

func ExampleNew() {
	half, full := rectify.New(rectify.HalfWave), rectify.New(rectify.FullWave)
	for _, x := range []float64{-0.5, 0.25} {
		fmt.Println(half.Shape(x), full.Shape(x))
	}
	// Output:
	// 0 0.5
	// 0.25 0.25
}

func ExampleNewDiode() {
	d := rectify.NewDiode(0.2, 0.1)
	for _, x := range []float64{0, 0.2, 0.5} {
		fmt.Printf("%.4f ", d.Shape(x))
	}
	fmt.Println()
	// Output:
	// 0.0000 0.0250 0.3000
}
