package repeat_test

import (
	"fmt"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/repeat"
	"github.com/almerlucke/genny/sequence"
)

// Repeat plays a finite generator n times.
func ExampleNew() {
	g := repeat.New[int](sequence.New(1, 2, 3), 2)
	for !g.Done() {
		fmt.Print(g.Generate(), " ")
	}
	fmt.Println()
	// Output: 1 2 3 1 2 3
}

// NewRand picks the number of repetitions from a range.
func ExampleNewRand() {
	genny.Seed(4)
	g := repeat.NewRand[int](sequence.New(7), 2, 4)
	n := 0
	for !g.Done() {
		g.Generate()
		n++
	}
	fmt.Println(n >= 2 && n <= 4)
	// Output: true
}

func ExampleNewWithFunc() {
	rounds := 0
	g := repeat.NewWithFunc[int](sequence.New(1, 2), func() int { rounds++; return rounds })
	play := func() {
		var out []int
		for !g.Done() {
			out = append(out, g.Generate())
		}
		fmt.Println(out)
	}
	play()
	g.Reset() // the function is asked again: two repeats now
	play()
	// Output:
	// [1 2]
	// [1 2 1 2]
}
