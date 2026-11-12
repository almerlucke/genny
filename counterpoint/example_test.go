package counterpoint_test

import (
	"fmt"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/counterpoint"
	"github.com/almerlucke/genny/scale"
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

// Second species: two notes against each note of the cantus (the last note gets one). The result passes
// CheckSecond: consonant on the beat, passing and neighbour tones off the beat, the last note by step
func ExampleGenerateSecond() {
	genny.Seed(3)
	cf := []int{62, 65, 64, 62}
	cp, err := counterpoint.GenerateSecond(cf, counterpoint.Config{Scale: scale.New(62, scale.Dorian)})
	fmt.Println(err, len(cp), counterpoint.CheckSecond(cf, cp))
	// Output: <nil> 7 []
}
