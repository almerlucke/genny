package line_test

import (
	"fmt"

	"github.com/almerlucke/genny/line"
	"github.com/almerlucke/genny/note"
	"github.com/almerlucke/genny/scale"
)

func motif() []note.Event {
	var out []note.Event
	for i, p := range []float64{60, 62, 64, 67} {
		out = append(out, note.Event{Pitch: p, Velocity: 0.8, Start: float64(i), Duration: 1})
	}
	return out
}

func show(events []note.Event) {
	for i, e := range events {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Printf("%g@%g", e.Pitch, e.Start)
	}
	fmt.Println()
}

// A motif, its inversion in C major around E, and its retrograde
func ExampleInvert() {
	c := scale.New(60, scale.Major)
	show(motif())
	show(line.Invert(motif(), 64, c))
	show(line.Retrograde(motif()))
	// Output:
	// 60@0 62@1 64@2 67@3
	// 67@0 65@1 64@2 60@3
	// 67@0 64@1 62@2 60@3
}

// A canon at the fifth: the second voice enters 4 steps later, 7 semitones higher
func ExampleCanon() {
	for _, v := range line.Canon(motif(), 4, 7, 2) {
		show(v)
	}
	// Output:
	// 60@0 62@1 64@2 67@3
	// 67@4 69@5 71@6 74@7
}

// Augmentation doubles the durations, diminution halves them
func ExampleAugment() {
	show(line.Augment(motif(), 2))
	show(line.Diminish(motif(), 2))
	// Output:
	// 60@0 62@2 64@4 67@6
	// 60@0 62@0.5 64@1 67@1.5
}
