package section_test

import (
	"fmt"

	"github.com/almerlucke/genny/constant"
	"github.com/almerlucke/genny/section"
	"github.com/almerlucke/genny/sequence"
)

// A song form: A for 4 steps, B for 2 steps, and around again. With Restart the melody of A starts
// from its first note each time, without it A carries on where it stopped
func ExampleNewLoop() {
	a := sequence.NewLoop("a1", "a2", "a3", "a4", "a5", "a6")
	b := constant.New("B")
	song := section.NewLoop(
		section.Part[string]{Gen: a, Steps: 4, Restart: true},
		section.Part[string]{Gen: b, Steps: 2},
	)
	for range 12 {
		fmt.Print(song.Generate(), " ")
	}
	fmt.Println()
	// Output: a1 a2 a3 a4 B B a1 a2 a3 a4 B B
}

// With Steps 0 a part plays until its generator is done
func ExampleNew() {
	intro := sequence.New("i1", "i2", "i3")
	song := section.New(
		section.Part[string]{Gen: intro},
		section.Part[string]{Gen: constant.New("verse"), Steps: 2},
	)
	for !song.Done() {
		fmt.Print(song.Generate(), " ")
	}
	fmt.Println()
	// Output: i1 i2 i3 verse verse
}
