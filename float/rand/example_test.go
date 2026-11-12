package rand_test

import (
	"fmt"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/float/rand"
)

func ExampleNew() {
	genny.Seed(7)
	r := rand.New(-1, 1)
	fmt.Printf("%.3f %.3f %.3f\n", r.Generate(), r.Generate(), r.Generate())
	genny.Seed(7) // the same seed gives the same values
	fmt.Printf("%.3f\n", r.Generate())
	// Output:
	// -0.937 0.936 -0.243
	// -0.937
}
