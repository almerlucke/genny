package evolve_test

import (
	"fmt"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/evolve"
)

// A phrase of 12 scale degrees is evolved towards an arch: it rises and falls, in small steps
func ExampleGA() {
	genny.Seed(2)
	ga := evolve.New(evolve.Config{
		Population: 40, Length: 12, Min: 0, Max: 14,
		Fitness: evolve.Weighted(
			evolve.Weight(evolve.Contour(evolve.Arch(12)), 2),
			evolve.Weight(evolve.Smoothness(), 1)),
		MutationRate: 0.1, Elitism: 2,
	})
	for _, gen := range []int{0, 10, 50} {
		b := ga.Run(gen - ga.Generation())
		fmt.Printf("generation %2d: fitness %.2f %v\n", b.Generation, b.Fitness, b.Phrase)
	}
	// Output:
	// generation  0: fitness 0.56 [7 5 14 8 14 11 9 12 12 8 2 9]
	// generation 10: fitness 0.73 [3 3 6 6 8 10 9 10 10 5 3 0]
	// generation 50: fitness 0.80 [2 4 6 7 9 9 9 8 7 6 4 2]
}

// The player plays the best phrase and evolves further after each pass
func ExamplePlayer() {
	genny.Seed(3)
	ga := evolve.New(evolve.Config{
		Population: 30, Length: 6, Min: 0, Max: 9,
		Fitness:      evolve.Target([]int{0, 2, 4, 6, 8, 9}),
		MutationRate: 0.15, Elitism: 1,
	})
	p := evolve.NewPlayer(ga, 8)
	for range 4 {
		var notes []int
		for range 6 {
			notes = append(notes, p.Generate())
		}
		fmt.Println(notes)
	}
	// Output:
	// [0 2 4 7 3 5]
	// [0 2 4 6 6 9]
	// [0 2 4 6 8 9]
	// [0 2 4 6 8 9]
}

// Weighted combines fitness functions
func ExampleWeighted() {
	f := evolve.Weighted(evolve.Weight(evolve.Smoothness(), 1), evolve.Weight(evolve.IntervalVariety(), 1))
	fmt.Printf("%.2f %.2f\n", f([]int{0, 1, 2, 3}), f([]int{0, 1, 3, 6}))
	// Output:
	// 0.42 0.67
}
