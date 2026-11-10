package ratchet_test

import (
	"fmt"

	"github.com/almerlucke/genny/ratchet"
)

// Offsets within the step: one hit, a roll of two, three, four
func ExampleOffsets() {
	for n := 1; n <= 4; n++ {
		fmt.Println(ratchet.Offsets(n))
	}
	// Output:
	// [0]
	// [0 0.5]
	// [0 0.3333333333333333 0.6666666666666666]
	// [0 0.25 0.5 0.75]
}

// Mostly single hits, with a roll on the 3rd and 5th step
func ExamplePattern() {
	p := ratchet.NewPattern(1, 1, 3, 1, 4)
	for range 5 {
		fmt.Print(len(p.Generate()), " ")
	}
	fmt.Println()
	// Output: 1 1 3 1 4
}
