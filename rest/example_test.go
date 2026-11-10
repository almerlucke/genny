package rest_test

import (
	"fmt"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/euclid"
	"github.com/almerlucke/genny/rest"
	"github.com/almerlucke/genny/sequence"
)

func show[T any](g interface{ Generate() rest.Optional[T] }, n int) {
	for range n {
		if o := g.Generate(); o.Played() {
			fmt.Print(o.Value, " ")
		} else {
			fmt.Print("_ ")
		}
	}
	fmt.Println()
}

// A Euclidean rhythm decides which steps of a melody are played. A rest also takes a note of the
// melody, so the melody keeps its place in time
func ExampleGated() {
	melody := sequence.NewLoop(60, 62, 64, 65, 67, 69, 71, 72)
	g := rest.NewGated[int](melody, euclid.New(8, 3, 0))
	show[int](g, 8)
	// Output: 60 _ _ 65 _ _ 71 _
}

// Consume(false) lets a rest leave the note for the next played step
func ExampleGated_Consume() {
	melody := sequence.NewLoop(60, 62, 64, 65, 67, 69, 71, 72)
	g := rest.NewGated[int](melody, euclid.New(8, 3, 0)).Consume(false)
	show[int](g, 8)
	// Output: 60 _ _ 62 _ _ 64 _
}

// Probability thins out a melody at random
func ExampleProbability() {
	genny.Seed(1)
	melody := sequence.NewLoop(60, 62, 64, 65)
	g := rest.NewProbability[int](melody, 0.5)
	show[int](g, 12)
	// Output: 60 _ _ _ _ 62 _ _ _ _ 64 _
}
