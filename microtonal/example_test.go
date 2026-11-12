package microtonal_test

import (
	"fmt"

	"github.com/almerlucke/genny/microtonal"
	"github.com/almerlucke/genny/note"
)

// A 31-EDO triad (root, third, fifth) needs three channels: every note gets its own
// pitch bend.
func ExampleRender() {
	e31 := microtonal.EDO(60, 31)
	var chord []note.Event
	for _, degree := range []int{0, 10, 18} {
		chord = append(chord, note.Event{Pitch: e31.Pitch(degree), Velocity: 0.8, Duration: 4})
	}
	for _, tr := range microtonal.Render(chord, &e31, microtonal.Strategy{MPE: true}) {
		for _, e := range tr.Events {
			b, _ := microtonal.Locate(e.Pitch, 2)
			fmt.Printf("channel %d: note %d bend %d (%+.2f cents)\n", tr.Channel+1, b.Note, b.Value, b.Cents(2))
		}
	}
	// Output:
	// channel 2: note 60 bend 8192 (+0.00 cents)
	// channel 3: note 64 bend 7663 (-12.92 cents)
	// channel 4: note 67 bend 8060 (-3.22 cents)
}
