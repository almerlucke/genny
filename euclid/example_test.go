package euclid_test

import (
	"fmt"

	"github.com/almerlucke/genny/euclid"
)

func show(p []bool) string {
	s := ""
	for _, v := range p {
		if v {
			s += "x"
		} else {
			s += "."
		}
	}
	return s
}

// Pulses spread as evenly as possible. Many world rhythms are Euclidean rhythms
func ExamplePattern() {
	fmt.Println("tresillo  ", show(euclid.Pattern(8, 3, 0)))
	fmt.Println("cinquillo ", show(euclid.Pattern(8, 5, 0)))
	fmt.Println("four floor", show(euclid.Pattern(16, 4, 0)))
	fmt.Println("rotated   ", show(euclid.Pattern(8, 3, 2)))
	// Output:
	// tresillo   x..x..x.
	// cinquillo  x.xx.xx.
	// four floor x...x...x...x...
	// rotated    .x..x.x.
}

// Euclid is a looping gate generator
func ExampleEuclid() {
	kick := euclid.New(4, 1, 0)
	for range 8 {
		fmt.Print(kick.Generate(), " ")
	}
	fmt.Println()
	// Output: true false false false true false false false
}
