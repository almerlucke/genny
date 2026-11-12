package parallel_test

import (
	"fmt"
	"math"

	"github.com/almerlucke/genny/float/shape/shapers/mult"
	"github.com/almerlucke/genny/float/shape/shapers/parallel"
	"github.com/almerlucke/genny/float/shape/shapers/thru"
)

// Parallel feeds its input to every shaper and sums the results.
func ExampleNew() {
	p := parallel.New(thru.New(), mult.New(2))
	fmt.Printf("%.3f\n", p.Shape(0.5))
	// Output:
	// 1.500
}

// NewWithFunction folds the outputs with another function, here the maximum.
func ExampleNewWithFunction() {
	p := parallel.NewWithFunction(math.Inf(-1), math.Max, mult.New(2), mult.New(-3))
	fmt.Printf("%.3f %.3f\n", p.Shape(1), p.Shape(-1))
	// Output:
	// 2.000 3.000
}
