package genny_test

import (
	"fmt"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/sequence"
)

// Every random generator draws from one stream; Seed makes a whole piece reproducible.
func ExampleSeed() {
	genny.Seed(42)
	a := []int{genny.Intn(100), genny.Intn(100), genny.Intn(100)}
	genny.Seed(42)
	b := []int{genny.Intn(100), genny.Intn(100), genny.Intn(100)}
	fmt.Println(a[0] == b[0] && a[1] == b[1] && a[2] == b[2])
	// Output: true
}

func ExampleFloat64() {
	genny.Seed(1)
	v := genny.Float64()
	fmt.Println(v >= 0 && v < 1)
	// Output: true
}

func ExampleShuffle() {
	genny.Seed(3)
	s := []int{1, 2, 3, 4, 5}
	genny.Shuffle(len(s), func(i, j int) { s[i], s[j] = s[j], s[i] })
	sum := 0
	for _, v := range s {
		sum += v
	}
	fmt.Println(len(s), sum)
	// Output: 5 15
}

// A Generator is a stream of values that can end (Done) and restart (Reset).
func ExampleGenerator() {
	var g genny.Generator[string] = sequence.New("a", "b")
	for !g.Done() {
		fmt.Print(g.Generate(), " ")
	}
	g.Reset()
	fmt.Println(g.Generate())
	// Output: a b a
}
