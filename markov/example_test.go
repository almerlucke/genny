package markov_test

import (
	"fmt"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/markov"
)

// A chain of ProbabilityStates: after "tonic" comes "dominant" with 70% and
// "subdominant" with 30%; both return to the tonic, and the tonic may end the
// chain (a nil successor ends the generator).
func ExampleNew() {
	genny.Seed(5)
	tonic := markov.NewProbabilityState("I")
	dom := markov.NewProbabilityState("V")
	sub := markov.NewProbabilityState("IV")
	tonic.SetProbabilities(dom, 0.7, sub, 0.3)
	dom.SetProbabilities(tonic, 1.0)
	sub.SetProbabilities(tonic, 1.0)

	m := markov.New[string](tonic, 0)
	var out []string
	for i := 0; i < 6; i++ {
		out = append(out, m.Generate())
	}
	fmt.Println(out[0], out[2], out[4], m.Done())
	// Output: I I I false
}
