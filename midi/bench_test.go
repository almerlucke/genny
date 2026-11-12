package midi

import (
	"bytes"
	"testing"

	"github.com/almerlucke/genny/note"
)

func benchFile(n int) File {
	ev := make([]note.Event, n)
	for i := range ev {
		ev[i] = note.Event{Pitch: float64(48 + i%24), Velocity: 0.8, Start: float64(i), Duration: 0.9}
	}
	return File{Tempo: 120, Tracks: []Track{{Events: ev}, {Channel: 1, Events: ev}}}
}

func BenchmarkWrite1000(b *testing.B) {
	f := benchFile(1000)
	var buf bytes.Buffer
	b.ReportAllocs()
	for range b.N {
		buf.Reset()
		if err := Write(&buf, f); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRead1000(b *testing.B) {
	var buf bytes.Buffer
	Write(&buf, benchFile(1000))
	data := buf.Bytes()
	b.ReportAllocs()
	for range b.N {
		if _, err := Read(bytes.NewReader(data)); err != nil {
			b.Fatal(err)
		}
	}
}
