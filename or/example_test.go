package or_test

import (
	"fmt"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/constant"
	"github.com/almerlucke/genny/or"
	"github.com/almerlucke/genny/sequence"
)

// Or plays one of its generators to the end, then picks the next one.
// In Indexed mode every generator is used once, in a shuffled order.
func ExampleNew() {
	genny.Seed(1)
	g := or.New[int](or.Indexed, sequence.New(1, 2), sequence.New(10, 20))
	n := 0
	sum := 0
	for !g.Done() {
		sum += g.Generate()
		n++
	}
	fmt.Println(n, sum)
	// Output: 4 33
}

// In Random mode it keeps picking forever.
func ExampleNewLoop() {
	genny.Seed(1)
	g := or.NewLoop[int](or.Random, constant.New(1), constant.New(2))
	fmt.Println(g.Continuous(), g.Done())
	// Output: true false
}
