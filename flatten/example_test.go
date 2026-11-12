package flatten_test

import (
	"fmt"

	"github.com/almerlucke/genny/flatten"
	"github.com/almerlucke/genny/sequence"
)

// Flatten turns a generator of slices into a generator of their elements.
func ExampleNew() {
	g := flatten.New[int](sequence.New([]int{1, 2}, []int{3}, []int{4, 5, 6}))
	for !g.Done() {
		fmt.Print(g.Generate(), " ")
	}
	fmt.Println()
	// Output: 1 2 3 4 5 6
}
