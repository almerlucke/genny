package genny_test

import (
	"bytes"
	"slices"
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/midi"
	"github.com/almerlucke/genny/ngram/style"
	"github.com/almerlucke/genny/note"
)

// a melody in the form of events
func midiTune() []note.Event {
	var out []note.Event
	t := 0.0
	for i, p := range []float64{64, 64, 65, 67, 67, 65, 64, 62, 60, 60, 62, 64, 64, 62, 62} {
		d := 4.0
		if i == 12 {
			d = 6
		}
		out = append(out, note.Event{Pitch: p, Velocity: 0.8, Start: t, Duration: d})
		t += d
	}
	return out
}

// a learned style, read back from a MIDI file, plays the same notes after the same seed and other ones after
// another seed
func TestLearnedStyleIsSeeded(t *testing.T) {
	var buf bytes.Buffer
	if err := midi.Write(&buf, midi.File{Tracks: []midi.Track{{Events: midiTune()}}}); err != nil {
		t.Fatal(err)
	}
	f, err := midi.Read(&buf)
	if err != nil || len(f.Tracks) != 1 {
		t.Fatal(err)
	}

	play := func(seed uint64) []style.Note {
		genny.Seed(seed)
		return style.Learn(f.Tracks[0].Events, 2, style.Intervals(), style.Ratios()).Phrase(40)
	}
	a, b, c := play(31), play(31), play(32)
	if !slices.Equal(a, b) {
		t.Error("the same seed gave other notes")
	}
	if slices.Equal(a, c) {
		t.Error("another seed gave the same notes")
	}
}
