package midi

import (
	"bytes"
	"math"
	"testing"

	"github.com/almerlucke/genny/note"
)

func fuzzSeedFiles() [][]byte {
	files := []File{
		{},
		{Tracks: []Track{{Events: []note.Event{{Pitch: 60, Velocity: 1, Start: 0, Duration: 1}}}}},
		{
			Tempo: 90, StepsPerBeat: 2,
			Tracks: []Track{
				{Name: "a", Channel: 3, Program: 12, Events: []note.Event{{Pitch: 64, Velocity: 0.5, Start: 1, Duration: 2}, {Pitch: 67, Velocity: 0.2, Start: 1, Duration: 4}}},
				{Channel: 9, Events: []note.Event{{Pitch: 36, Velocity: 0.9, Start: 0, Duration: 0.5}}},
			},
			TempoChanges:   []TempoChange{{Step: 8, BPM: 140}},
			TimeSignatures: []TimeSignature{{Step: 0, Num: 6, Den: 8}},
		},
		{Tracks: []Track{{PitchBend: true, Events: []note.Event{{Pitch: 60.5, Velocity: 1, Start: 0, Duration: 1}}}}},
		{Tracks: []Track{{
			Events:         []note.Event{{Pitch: 60, Velocity: 1, Start: 0, Duration: 4}},
			Controllers:    []Controller{{Step: 1, Number: 7, Value: 100}},
			ProgramChanges: []ProgramChange{{Step: 2, Program: 5}},
		}}},
	}
	var out [][]byte
	for _, f := range files {
		var buf bytes.Buffer
		if Write(&buf, f) == nil {
			out = append(out, buf.Bytes())
		}
	}
	out = append(out, nil, []byte("MThd"), []byte("MThd\x00\x00\x00\x06\x00\x01\x00\x01\x01\xe0MTrk\x00\x00\x00\x00"), []byte("garbage that is no midi file"))
	return out
}

// FuzzReadRoundTrip: arbitrary bytes never panic Read, and a file that parses can be written and read again
// to the same notes
func FuzzReadRoundTrip(f *testing.F) {
	for _, s := range fuzzSeedFiles() {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		got, err := Read(bytes.NewReader(data))
		if err != nil {
			return
		}
		var buf bytes.Buffer
		if err := Write(&buf, got); err != nil {
			return // a file may hold things Write refuses (for instance steps out of range)
		}
		again, err := Read(&buf)
		if err != nil {
			t.Fatalf("a written file does not read: %v", err)
		}
		// durations may shrink (a note retriggered on the same pitch cuts the one before it), so only
		// pitch, start and velocity are compared; compare the notes of all tracks together, ordered by track: empty tracks may vanish
		flat := func(f File) (out []note.Event) {
			for _, tr := range f.Tracks {
				out = append(out, tr.Events...)
			}
			return
		}
		a, b := flat(got), flat(again)
		if len(a) != len(b) {
			t.Fatalf("%d notes became %d", len(a), len(b))
		}
		for i := range a {
			if math.Abs(a[i].Start-b[i].Start) > 0.05 || !(b[i].Duration > 0) ||
				math.Abs(a[i].Pitch-b[i].Pitch) > 0.51 || math.Abs(a[i].Velocity-b[i].Velocity) > 0.02 {
				t.Fatalf("note %d: %+v became %+v", i, a[i], b[i])
			}
		}
	})
}
