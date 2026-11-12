package sequence_test

import (
	"fmt"

	"github.com/almerlucke/genny/sequence"
)

func ExampleNew() {
	s := sequence.New(60, 62, 64)
	for !s.Done() {
		fmt.Print(s.Generate(), " ")
	}
	// A finished sequence keeps returning its last value.
	fmt.Println(s.Generate())
	// Output: 60 62 64 64
}

func ExampleNewLoop() {
	s := sequence.NewLoop("a", "b", "c")
	for i := 0; i < 7; i++ {
		fmt.Print(s.Generate(), " ")
	}
	fmt.Println()
	// Output: a b c a b c a
}
