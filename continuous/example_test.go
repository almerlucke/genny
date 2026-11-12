package continuous_test

import (
	"fmt"

	"github.com/almerlucke/genny/continuous"
	"github.com/almerlucke/genny/sequence"
)

// Continuous restarts a finite generator whenever it is done.
func ExampleNew() {
	g := continuous.New[int](sequence.New(1, 2, 3))
	for i := 0; i < 7; i++ {
		fmt.Print(g.Generate(), " ")
	}
	fmt.Println(g.Continuous(), g.Done())
	// Output: 1 2 3 1 2 3 1 true false
}
