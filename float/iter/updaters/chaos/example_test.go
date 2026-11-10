package chaos_test

import (
	"fmt"

	"github.com/almerlucke/genny/float"
	"github.com/almerlucke/genny/float/iter"
	"github.com/almerlucke/genny/float/iter/updaters/chaos"
)

// The Lorenz attractor as a control source: x and y stay roughly within [-20, 20], z within
// [0, 50]. dt sets the speed. Scale the dimensions to what you need to control
func ExampleLorenz() {
	lorenz := iter.New([]float64{1, 1, 1}, chaos.NewLorenz(10, 28, 8.0/3, 0.01))
	x := float.FromFrame(lorenz, 0)
	for range 8 {
		for range 99 {
			x.Generate() // skip: dt is small
		}
		fmt.Printf("%.1f ", x.Generate())
	}
	fmt.Println()
	// Output: -9.5 -8.0 -7.6 -10.0 -6.5 -9.9 -7.5 -7.3
}
