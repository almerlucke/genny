package isorhythm_test

import (
	"fmt"

	"github.com/almerlucke/genny/isorhythm"
	"github.com/almerlucke/genny/note"
)

// A talea of 3 durations against a color of 2 pitches: the pattern only starts again after 6 notes
func ExampleEvents() {
	for i, e := range isorhythm.Events([]float64{2, 1, -1}, []float64{60, 67}, 12) {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Printf("%g@%g(%g)", e.Pitch, e.Start, e.Duration)
	}
	fmt.Println()
	fmt.Println(isorhythm.Period([]float64{2, 1, -1}, []float64{60, 67}))
	// Output:
	// 60@0(2) 67@2(1) 67@4(2) 60@6(1) 60@8(2) 67@10(1)
	// 6
}

// The same on a grid of steps, for a note.Voice
func ExampleSteps() {
	s := isorhythm.New([]float64{2, 1, -1}, []float64{60, 67})
	v := note.Voice{Pitch: s.Pitch(), Length: s.Length()}
	for _, e := range v.Render(8) {
		fmt.Printf("%g@%g(%g) ", e.Pitch, e.Start, e.Duration)
	}
	fmt.Println()
	// Output: 60@0(2) 67@2(1) 67@4(2) 60@6(1)
}
