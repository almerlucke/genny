package split_test

import (
	"fmt"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/split"
)

// Split cuts a bar of 4 into halves, recursively and at random, and always fills the bar
func ExampleSplit() {
	genny.Seed(3)
	for range 3 {
		parts := split.Split(4, 0.25, 0.6, 2)
		sum := 0.0
		for _, p := range parts {
			sum += p
		}
		fmt.Println(parts, sum)
	}
	// Output:
	// [2 0.25 0.25 0.25 0.25 1] 4
	// [0.25 0.25 0.5 0.5 0.5 2] 4
	// [4] 4
}

// With probability 1 the splitting goes on until the parts would get shorter than lower
func ExampleSplit_complete() {
	fmt.Println(split.Split(3, 0.5, 1, 3))
	// Output:
	// [1 1 1]
}

// Metric is an equal division, for instance a triplet
func ExampleMetric() {
	fmt.Println(split.Metric(1, 3))
	fmt.Println(split.Metric(1, 4))
	// Output:
	// [0.3333333333333333 0.3333333333333333 0.3333333333333333]
	// [0.25 0.25 0.25 0.25]
}

// Gen loops and splits anew every cycle
func ExampleNewGen() {
	genny.Seed(8)
	g := split.NewGen(2, 0.25, 0.6, 2)
	for range 12 {
		fmt.Print(g.Generate(), " ")
	}
	fmt.Println()
	// Output:
	// 0.5 0.5 0.25 0.25 0.25 0.25 0.25 0.25 0.25 0.25 1 2
}
