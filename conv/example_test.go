package conv_test

import (
	"fmt"
	"strconv"

	"github.com/almerlucke/genny/conv"
	"github.com/almerlucke/genny/sequence"
)

type toString struct{}

func (toString) Convert(i int) string { return "n" + strconv.Itoa(i) }

func ExampleNew() {
	g := conv.New[int, string](sequence.New(60, 62, 64), toString{})
	for !g.Done() {
		fmt.Print(g.Generate(), " ")
	}
	fmt.Println()
	// Output: n60 n62 n64
}

// ToAny erases the value type so generators of different types can be mixed.
func ExampleToAny() {
	g := conv.ToAny[int](sequence.New(1, 2))
	fmt.Printf("%T %v\n", g.Generate(), g.Generate())
	// Output: int 2
}
