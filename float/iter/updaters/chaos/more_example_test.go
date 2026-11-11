package chaos_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/iter"
	"github.com/almerlucke/genny/float/iter/updaters/chaos"
)

// The standard map folds the state into [0, 2 pi) on every step
func ExampleStandardMap() {
	it := iter.New([]float64{1, 0.5}, chaos.NewStandardMap(1))
	for range 3 {
		f := it.Generate()
		fmt.Printf("%.3f %.3f\n", f[0], f[1])
	}
	// Output:
	// 1.000 0.500
	// 2.341 1.341
	// 4.400 2.059
}

// Take the angle of a chaotic pendulum as a modulation signal
func ExampleProject() {
	it := iter.New([]float64{0.2, 0}, chaos.NewPendulum())
	angle := chaos.Project(it, chaos.AxisX)
	for range 4 {
		fmt.Printf("%.3f\n", angle.Generate()[0])
	}
	// Output:
	// 0.200
	// 0.201
	// 0.205
	// 0.211
}

// A Duffing oscillator in its double well
func ExampleDuffing() {
	it := iter.New([]float64{1, 0}, chaos.NewDuffing())
	for range 3 {
		f := it.Generate()
		fmt.Printf("x %.3f v %.3f\n", f[0], f[1])
	}
	// Output:
	// x 1.000 v 0.000
	// x 1.001 v 0.025
	// x 1.002 v 0.049
}
