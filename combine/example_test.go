package combine_test

import (
	"fmt"

	"github.com/almerlucke/genny/combine"
	"github.com/almerlucke/genny/constant"
	"github.com/almerlucke/genny/sequence"
)

// Combine takes one value from each generator per step, for example the
// pitches of a chord.
func ExampleNew() {
	g := combine.New[int](
		sequence.NewLoop(60, 62, 64),
		sequence.NewLoop(64, 65, 67),
		constant.New(72),
	)
	for i := 0; i < 3; i++ {
		fmt.Println(g.Generate())
	}
	// Output:
	// [60 64 72]
	// [62 65 72]
	// [64 67 72]
}
