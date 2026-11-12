package and_test

import (
	"fmt"

	"github.com/almerlucke/genny/and"
	"github.com/almerlucke/genny/sequence"
)

// And plays its generators one after the other; here two phrases are chained.
func ExampleNew() {
	g := and.New[int](sequence.New(1, 2), sequence.New(3, 4, 5))
	for !g.Done() {
		fmt.Print(g.Generate(), " ")
	}
	fmt.Println()
	// Output: 1 2 3 4 5
}

// NewLoop starts over after the last generator.
func ExampleNewLoop() {
	g := and.NewLoop[int](sequence.New(1, 2), sequence.New(3))
	for i := 0; i < 7; i++ {
		fmt.Print(g.Generate(), " ")
	}
	fmt.Println()
	// Output: 1 2 3 1 2 3 1
}
