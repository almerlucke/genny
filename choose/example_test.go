package choose_test

import (
	"fmt"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/choose"
	"github.com/almerlucke/genny/constant"
	"github.com/almerlucke/genny/sequence"
)

// Mostly a pattern, sometimes a fixed accent. Only the generator that is picked advances,
// so the pattern is not interrupted
func ExampleWeighted() {
	genny.Seed(4)
	pattern := sequence.NewLoop("a", "b", "c")
	g := choose.New(
		choose.Option[string]{Gen: pattern, Weight: 3},
		choose.Option[string]{Gen: constant.New("!"), Weight: 1},
	)
	for range 12 {
		fmt.Print(g.Generate(), " ")
	}
	fmt.Println()
	// Output: ! ! a b ! c a ! b c ! a
}
