package unwrap_test

import (
	"fmt"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/sequence"
	"github.com/almerlucke/genny/unwrap"
)

// Unwrap flattens a generator of generators: each inner generator is played to
// its end before the next is taken.
func ExampleNew() {
	phrases := sequence.New[genny.Generator[int]](
		sequence.New(1, 2),
		sequence.New(3),
		sequence.New(4, 5),
	)
	g := unwrap.New[int](phrases)
	for !g.Done() {
		fmt.Print(g.Generate(), " ")
	}
	fmt.Println()
	// Output: 1 2 3 4 5
}
