package funnel_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/funnel"
	"github.com/almerlucke/genny/scale"
	"github.com/almerlucke/genny/sieve"
)

func ExampleNew() {
	f := funnel.New([]float64{0, 1, 5})
	for _, x := range []float64{-2, 0.4, 0.6, 3, 3.1, 9} {
		fmt.Print(f.Shape(x), " ")
	}
	fmt.Println()
	// Output:
	// 0 0 1 1 5 5
}

// Snap a pitch to the notes of a sieve: a major triad on C4
func ExampleFromSieve() {
	f := funnel.FromSieve(sieve.Of(0, 11, 0, 4, 7), 60)
	for _, x := range []float64{61, 62.5, 66, 70} {
		fmt.Print(f.Shape(x), " ")
	}
	fmt.Println()
	// Output:
	// 60 64 67 67
}

// Snap to A minor in every octave
func ExampleFromScale() {
	f := funnel.FromScale(scale.New(57, scale.Minor))
	for _, x := range []float64{58, 61.4, 70, 45.6} {
		fmt.Print(f.Shape(x), " ")
	}
	fmt.Println()
	// Output:
	// 57 62 69 45
}
