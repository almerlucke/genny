package stochastic_test

import (
	"fmt"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/rest"
	"github.com/almerlucke/genny/sequence"
	"github.com/almerlucke/genny/stochastic"
)

// Poisson onset times: the gaps between the events are random but the rate is steady
func ExamplePoisson() {
	genny.Seed(4)
	for _, t := range stochastic.Poisson(1, 8) {
		fmt.Printf("%.2f ", t)
	}
	fmt.Println()
	// Output:
	// 2.13 6.01 6.47 7.26
}

// A cloud that thickens: the density rises from 0 to 4 events per step
func ExampleCloud() {
	genny.Seed(4)
	cloud := stochastic.NewCloud(stochastic.Config{
		Density: func(step int) float64 { return float64(step) },
		Pitch:   stochastic.Uniform(60, 72),
	})
	counts := make([]int, 5)
	for _, e := range cloud.Render(5) {
		counts[int(e.Start)]++
	}
	fmt.Println(counts)
	// Output:
	// [0 2 2 5 3]
}

// A gate whose chance of a hit rises, for rest.Gated: the notes get denser
func ExampleDensity() {
	genny.Seed(4)
	ramp := make([]float64, 16)
	for i := range ramp {
		ramp[i] = float64(i) / 15
	}
	notes := rest.NewGated[float64](sequence.NewLoop(60.0), stochastic.NewDensity(sequence.New(ramp...)))
	for range 16 {
		if notes.Generate().Played() {
			fmt.Print("x")
		} else {
			fmt.Print(".")
		}
	}
	fmt.Println()
	// Output:
	// .....xx.xx.xxxxx
}

// PoissonInt draws how many events fall into a step when they come at random
func ExamplePoissonInt() {
	genny.Seed(4)
	for range 8 {
		fmt.Print(stochastic.PoissonInt(3), " ")
	}
	fmt.Println()
	// Output:
	// 5 3 3 5 3 5 3 5
}
