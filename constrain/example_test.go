package constrain_test

import (
	"fmt"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/constrain"
	"github.com/almerlucke/genny/function"
	"github.com/almerlucke/genny/sequence"
)

// Random pitches that only move by a second or a third: a smooth line out of a wild source
func ExampleAllowInterval() {
	genny.Seed(1)
	wild := function.New(func() int { return 60 + genny.Intn(12) })
	line := constrain.AllowInterval(wild, []int{1, 2, 3, 4})
	fmt.Println(take(line, 8))
	// Output:
	// [66 68 70 67 70 66 67 68]
}

// No note twice in a row
func ExampleRemoveRepetitions() {
	line := constrain.RemoveRepetitions(sequence.NewLoop("a", "a", "b", "b", "c"))
	fmt.Println(take(line, 5))
	// Output:
	// [a b c a b]
}

// A group of four values without doubles: a short twelve tone like series
func ExampleDuplicates() {
	genny.Seed(2)
	src := function.New(func() int { return genny.Intn(4) })
	series := constrain.Duplicates(src, 4)
	for range 3 {
		fmt.Println(take(series, 4))
	}
	// Output:
	// [2 1 0 3]
	// [3 2 1 0]
	// [2 0 3 1]
}

// A walk through the notes of C major, stepping by the intervals 2, 2, 1, -3 over and over
func ExampleScaleIntervals() {
	cMajor := []int{55, 57, 59, 60, 62, 64, 65, 67, 69, 71, 72, 74, 76}
	walk := constrain.ScaleIntervals(60, []int{2, 2, 1, -3}, cMajor)
	fmt.Println(take(walk, 9))
	// Output:
	// [60 62 64 65 62 64 65 62 64]
}

// Moves of a fifth, a second or a third, up or down, inside one octave
func ExampleRandomIntervals() {
	genny.Seed(3)
	walk := constrain.RandomIntervals(60, 72, 2, 3, 7)
	fmt.Println(take(walk, 8))
	// Output:
	// [66 64 62 64 71 69 71 68]
}

// Half of the notes are shot away
func ExampleShotgun() {
	genny.Seed(4)
	line := constrain.Shotgun(sequence.NewLoop(1, 2, 3, 4, 5, 6, 7, 8), 0.5)
	fmt.Println(take(line, 8))
	// Output:
	// [1 2 4 5 8 3 5 6]
}
