package route_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/mult"
	"github.com/almerlucke/genny/float/shape/shapers/route"
	"github.com/almerlucke/genny/float/shape/shapers/thru"
)

// Route sends the input through one selected shaper; change Index to switch.
func ExampleNew() {
	r := route.New(0, thru.New(), mult.New(10))
	fmt.Printf("%.1f\n", r.Shape(2))
	r.Index = 1
	fmt.Printf("%.1f\n", r.Shape(2))
	// Output:
	// 2.0
	// 20.0
}
