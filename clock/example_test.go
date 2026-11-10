package clock_test

import (
	"fmt"

	"github.com/almerlucke/genny/clock"
)

// A hi-hat on every step, a snare on every 4th step
func ExampleDivider() {
	snare := clock.NewDivider(4, 2)
	for range 8 {
		if snare.Generate() {
			fmt.Print("S ")
		} else {
			fmt.Print(". ")
		}
	}
	fmt.Println()
	// Output: . . S . . . S .
}

// Cycles of 3 and 4 steps drift against each other and meet again after 12 steps
func ExamplePoly() {
	p := clock.NewPoly(3, 4)
	for range p.Cycle() {
		t := p.Generate()
		a, b := ".", "."
		if t[0] {
			a = "a"
		}
		if t[1] {
			b = "b"
		}
		fmt.Print(a, b, " ")
	}
	fmt.Println()
	// Output: ab .. .. a. .b .. a. .. .b a. .. ..
}
