package midi_test

import (
	"bytes"
	"fmt"

	"github.com/almerlucke/genny/midi"
	"github.com/almerlucke/genny/note"
)

// Read what Write made: a track with notes, a tempo change and a control change halfway
func ExampleRead() {
	var buf bytes.Buffer
	midi.Write(&buf, midi.File{
		Tempo:        100,
		TempoChanges: []midi.TempoChange{{Step: 8, BPM: 60}},
		Tracks: []midi.Track{{
			Name: "lead", Program: 4,
			Controllers: []midi.Controller{{Step: 4, Number: 7, Value: 90}},
			Events: []note.Event{
				{Pitch: 60, Velocity: 1, Start: 0, Duration: 2},
				{Pitch: 64, Velocity: 1, Start: 2.5, Duration: 1.5},
			},
		}},
	})

	f, err := midi.Read(&buf)
	fmt.Println(err, f.Tempo, f.TempoChanges, len(f.Tracks))
	t := f.Tracks[0]
	fmt.Println(t.Name, t.Program, t.Controllers)
	for _, e := range t.Events {
		fmt.Printf("%.1f at %.1f for %.1f\n", e.Pitch, e.Start, e.Duration)
	}
	// Output:
	// <nil> 100 [{8 60}] 1
	// lead 4 [{4 7 90}]
	// 60.0 at 0.0 for 2.0
	// 64.0 at 2.5 for 1.5
}
