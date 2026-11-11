package dist_test

import (
	"fmt"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/float/dist"
)

func show(g interface{ Generate() float64 }, n int) {
	for i := range n {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Printf("%.2f", g.Generate())
	}
	fmt.Println()
}

// With a below b the values lean towards 0, scaled here to velocities between 0.2 and 1
func ExampleBeta() {
	genny.Seed(1)
	show(dist.NewBeta(2, 5).Scale(0.2, 1), 8)
	// Output:
	// 0.62 0.24 0.38 0.50 0.41 0.24 0.37 0.69
}

// Mostly near the center, but now and then a leap far away. Retry keeps the leaps within a range
func ExampleCauchy() {
	genny.Seed(1)
	show(dist.NewCauchy(60, 2).Retry(48, 72, 10), 8)
	// Output:
	// 57.85 62.43 60.19 60.06 58.57 62.51 60.33 60.16
}

// Durations of at least 0.25 with a long tail
func ExampleWeibull() {
	genny.Seed(1)
	show(dist.NewWeibull(0.25, 1, 0.8).Clamp(0.25, 4), 8)
	// Output:
	// 0.45 4.00 1.93 0.95 0.91 0.53 1.97 1.01
}

// A Lehmer generator is repeatable by its own seed, without genny.Seed
func ExampleLehmer() {
	show(dist.NewLehmer(1), 5)
	show(dist.NewLehmer(1), 5)
	// Output:
	// 0.24 0.37 0.50 0.70 0.05
	// 0.24 0.37 0.50 0.70 0.05
}

// Pink noise wanders slowly with some fast detail
func ExamplePink() {
	genny.Seed(1)
	show(dist.NewPink(6), 8)
	// Output:
	// 0.63 0.63 0.55 0.50 0.57 0.37 0.26 0.23
}
