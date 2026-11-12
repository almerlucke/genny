package transform_test

import (
	"fmt"

	"github.com/almerlucke/genny/sequence"
	"github.com/almerlucke/genny/transform"
)

// Transform applies a function to every generated value, here a transposition.
func ExampleNew() {
	g := transform.New[int](sequence.New(60, 62, 64), func(p int) int { return p + 7 })
	for !g.Done() {
		fmt.Print(g.Generate(), " ")
	}
	fmt.Println()
	// Output: 67 69 71
}
