package line_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/line"
)

// A Line holds a value until To starts a ramp to a new target; then it holds
// the target. Done reports that the ramp has arrived.
func ExampleNew() {
	l := line.New(0)
	l.To(1, 4)
	for !l.Done() {
		fmt.Printf("%.2f ", l.Generate())
	}
	fmt.Printf("| %.2f\n", l.Generate())
	l.To(0.5, 2)
	fmt.Printf("%.2f %.2f %.2f\n", l.Generate(), l.Generate(), l.Generate())
	// Output:
	// 0.00 0.25 0.50 0.75 | 1.00
	// 1.00 0.75 0.50
}
