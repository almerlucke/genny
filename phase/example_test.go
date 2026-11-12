package phase_test

import (
	"fmt"

	"github.com/almerlucke/genny/note"
	"github.com/almerlucke/genny/phase"
)

// Two voices play the same pattern of 4 steps; the second gains 1 step per cycle, so after 4 cycles it
// has caught up a whole cycle
func ExampleEvents() {
	var pattern []note.Event
	for i, p := range []float64{64, 66, 71, 73} {
		pattern = append(pattern, note.Event{Pitch: p, Velocity: 0.8, Start: float64(i), Duration: 1})
	}
	n := phase.Cycles(pattern, 1)
	first, second := phase.Events(pattern, n, 1)
	fmt.Println(n, len(first), len(second))
	for _, e := range second[4:6] {
		fmt.Printf("%g@%g ", e.Pitch, e.Start)
	}
	fmt.Println()
	// Output:
	// 4 16 21
	// 64@3 66@4
}
