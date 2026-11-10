package ngram_test

import (
	"fmt"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/ngram"
)

// Learn from a tune and play new music in its style: only moves that occur in the tune are made,
// as often as they occur. Order 2 follows the example more closely than order 1
func ExampleLearn() {
	genny.Seed(5)
	tune := []string{"C", "D", "E", "C", "C", "D", "E", "C", "E", "F", "G", "E", "F", "G"}
	chain := ngram.Learn(tune, 1)
	for range 16 {
		fmt.Print(chain.Generate(), " ")
	}
	fmt.Println()
	// Output: C D E C C D E F G E F G E C C C
}
