package counterpoint_test

import (
	"fmt"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/counterpoint"
)

var cantus = []int{62, 65, 64, 67, 65, 69, 67, 65, 64, 62}

// A line above a cantus firmus in D Dorian: it passes the rule check
func ExampleGenerate() {
	genny.Seed(3)
	cp, err := counterpoint.Generate(cantus, counterpoint.Config{})
	fmt.Println(err, len(cp), counterpoint.Check(cantus, cp))
	// Output: <nil> 10 []
}

// Check names what is wrong with a line: here parallel fifths between the first two notes
func ExampleCheck() {
	for _, v := range counterpoint.Check([]int{62, 65, 64, 62}, []int{69, 72, 71, 74}) {
		fmt.Println(v)
	}
	// Output:
	// note 1: parallel: parallel fifth
	// note 2: parallel: parallel fifth
}
