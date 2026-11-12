package plot_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/almerlucke/genny/float/export/plot"
	"github.com/almerlucke/genny/note"
	"gonum.org/v1/plot/vg"
)

func TestRollWritesAPNG(t *testing.T) {
	line := func(pitches ...float64) []note.Event {
		var out []note.Event
		for i, p := range pitches {
			out = append(out, note.Event{Pitch: p, Velocity: 0.8, Start: float64(i * 2), Duration: 2})
		}
		return out
	}
	tests := []struct {
		name   string
		tracks []plot.Track
	}{
		{"two tracks", []plot.Track{{"cantus", line(62, 65, 64, 62)}, {"counterpoint", line(69, 72, 71, 74)}}},
		{"empty track", []plot.Track{{"none", nil}, {"one", line(60)}}},
		{"no tracks", nil},
		{"zero duration", []plot.Track{{"x", []note.Event{{Pitch: 60}}}}},
	}
	for _, tt := range tests {
		file := filepath.Join(t.TempDir(), "roll.png")
		if err := plot.Roll(tt.tracks, 12*vg.Centimeter, 6*vg.Centimeter, file); err != nil {
			t.Fatalf("%s: %v", tt.name, err)
		}
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if len(data) < 8 || string(data[1:4]) != "PNG" {
			t.Errorf("%s: not a png (%d bytes)", tt.name, len(data))
		}
	}
	if err := plot.Roll(nil, vg.Inch, vg.Inch, ""); err == nil {
		t.Error("an empty path should be an error")
	}
}
