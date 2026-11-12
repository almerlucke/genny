package iter_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/iter"
)

// halve is an Updater: every step it halves each value.
type halve struct{}

func (halve) Update(v []float64) {
	for i := range v {
		v[i] /= 2
	}
}

// An Iterator emits the successive states of an Updater, starting with the
// initial values. Chaotic systems live in iter/updaters/chaos.
func ExampleNew() {
	it := iter.New([]float64{1, 8}, halve{})
	for i := 0; i < 4; i++ {
		fmt.Println(it.Generate())
	}
	// Output:
	// [1 8]
	// [0.5 4]
	// [0.25 2]
	// [0.125 1]
}
