package listtools_test

import (
	"fmt"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/listtools"
	"github.com/almerlucke/genny/sequence"
)

func take[T any](g genny.Generator[T], n int) (out []T) {
	for range n {
		out = append(out, g.Generate())
	}
	return
}

// Jump plays a few values in a row, then leaps to somewhere else in the list
func ExampleNewJump() {
	g := listtools.NewJump([]int{0, 1, 2, 3, 4, 5, 6, 7}, 3, 1, 4)
	fmt.Println(take[int](g, 9))
	// Output:
	// [0 1 2 4 5 6 0 1 2]
}

// Stutter repeats the last value after every two values
func ExampleNewStutter() {
	g := listtools.NewStutter[int](sequence.NewLoop(1, 2, 3, 4), 2, 2)
	fmt.Println(take[int](g, 10))
	// Output:
	// [1 2 2 2 3 4 4 4 1 2]
}

// Group repeats every value, here once, twice, three times
func ExampleNewGroup() {
	g := listtools.NewGroup[string](sequence.NewLoop("a", "b", "c"), sequence.NewLoop(1, 2, 3))
	fmt.Println(take[string](g, 6))
	// Output:
	// [a b b c c c]
}

// Grow is a series that doubles until it hits the limit and starts again
func ExampleNewGrow() {
	g := listtools.NewGrow(0.25, listtools.Geometric, 2, 0, 2, listtools.Restart)
	fmt.Println(take[float64](g, 9))
	// Output:
	// [0.25 0.5 1 2 0.25 0.5 1 2 0.25]
}

// GatherUntil fills a bar of 4 with durations, the last one is trimmed to fit
func ExampleGatherUntil() {
	g := sequence.NewLoop(1.5, 1, 0.75)
	fmt.Println(listtools.GatherUntil(g, 4))
	// Output:
	// [1.5 1 0.75 0.75]
}

// PlusMin turns durations into rests (negative values) at random
func ExampleNewPlusMin() {
	genny.Seed(4)
	g := listtools.NewPlusMin(sequence.NewLoop(1.0), 0.5)
	fmt.Println(take[float64](g, 8))
	// Output:
	// [1 1 -1 1 1 -1 -1 1]
}

// PositiveNegative alternates notes and rests: two notes, one rest
func ExampleNewPositiveNegative() {
	g := listtools.NewPositiveNegative(sequence.NewLoop(1.0, 0.5), 2, 1)
	fmt.Println(take[float64](g, 7))
	// Output:
	// [1 0.5 -1 0.5 1 -0.5 1]
}
