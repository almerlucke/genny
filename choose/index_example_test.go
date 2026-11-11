package choose_test

import (
	"fmt"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/choose"
)

func words[T any](g interface{ Generate() T }, n int) {
	for i := range n {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Print(g.Generate())
	}
	fmt.Println()
}

// Notes of a scale, the low ones more often than the high ones
func ExampleIndex() {
	genny.Seed(2)
	pitches := []int{57, 59, 60, 62, 64, 65, 67, 69}
	words[int](choose.NewIndex(pitches, choose.Linear()), 12)
	// Output:
	// 64 57 60 60 57 60 60 62 59 60 65 60
}

// Mostly around the third element, 1 element of spread
func ExampleGaussian() {
	genny.Seed(2)
	g := choose.NewIndex([]string{"a", "b", "c", "d", "e", "f"}, choose.Gaussian(2, 1))
	words[string](g, 12)
	// Output:
	// b d c c e b b b c d c d
}

// The weights move on every 4 picks: first the start of the list, then the end
func ExampleChangingWeights() {
	genny.Seed(2)
	g := choose.NewChangingWeights([]string{"low", "mid", "high"}, 4,
		[]float64{8, 2, 0}, []float64{0, 2, 8})
	words[string](g, 16)
	// Output:
	// low low low low high high high high mid low low low high high high high
}
