package progression_test

import (
	"fmt"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/progression"
)

var numerals = []string{"I", "ii", "iii", "IV", "V", "vi", "vii°"}

// The generator gives scale degrees (0 is I). Seed makes the progression repeatable
func ExampleNew() {
	genny.Seed(3)
	prog := progression.New(progression.Pop, 0)
	for range 8 {
		fmt.Print(numerals[prog.Generate()], " ")
	}
	fmt.Println()
	// Output: I V IV V I V I vi
}

// Your own table: a degree lists the moves that can follow it, with weights
func ExampleTable() {
	genny.Seed(1)
	twoChords := progression.Table{
		0: {{To: 5, Weight: 1}},
		5: {{To: 0, Weight: 1}},
	}
	prog := progression.New(twoChords, 0)
	for range 6 {
		fmt.Print(numerals[prog.Generate()], " ")
	}
	fmt.Println()
	// Output: I vi I vi I vi
}
