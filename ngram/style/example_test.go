package style_test

import (
	"fmt"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/ngram/style"
	"github.com/almerlucke/genny/note"
)

// Learn the style of a tune as intervals and duration ratios, then play new music in that style a
// fifth higher. Render it to events to write it to a MIDI file
func ExampleLearn() {
	genny.Seed(3)
	var tune []note.Event
	t := 0.0
	for i, p := range []float64{60, 62, 64, 60, 60, 62, 64, 60, 64, 65, 67, 64, 65, 67} {
		d := 4.0
		if i%4 == 3 {
			d = 8
		}
		tune = append(tune, note.Event{Pitch: p, Velocity: 0.8, Start: t, Duration: d})
		t += d
	}

	s := style.Learn(tune, 2, style.Intervals(), style.Ratios()).Transpose(7)
	phrase := s.Fill(24)
	for _, n := range phrase {
		fmt.Printf("%.0f:%.0f,", n.Pitch, n.Duration)
	}
	fmt.Println()
	events := style.Events(phrase, 0, s.Velocity)
	fmt.Println(len(events), "events, the last one starts at", events[len(events)-1].Start)
	// Output:
	// 71:4,67:8,67:4,69:4,71:4,
	// 5 events, the last one starts at 20
}
