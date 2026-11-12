package genny_test

import (
	"slices"
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/chord"
	"github.com/almerlucke/genny/counterpoint"
	"github.com/almerlucke/genny/function"
	"github.com/almerlucke/genny/line"
	"github.com/almerlucke/genny/melody"
	"github.com/almerlucke/genny/note"
	"github.com/almerlucke/genny/scale"
)

func lineEventValues(events []note.Event) (out []float64) {
	for _, e := range events {
		out = append(out, e.Pitch, e.Velocity, e.Start, e.Duration)
	}
	return
}

func lineIntValues(ints []int) (out []float64) {
	for _, i := range ints {
		out = append(out, float64(i))
	}
	return
}

var lineCantus = []int{62, 65, 64, 67, 65, 69, 67, 65, 64, 62, 60, 62}

// every random generator of the line packages
var lineRandomGenerators = map[string]func() []float64{
	"counterpoint above": func() []float64 {
		cp, _ := counterpoint.Generate(lineCantus, counterpoint.Config{})
		return lineIntValues(cp)
	},
	"counterpoint below": func() []float64 {
		cp, _ := counterpoint.Generate(lineCantus, counterpoint.Config{Side: counterpoint.Below})
		return lineIntValues(cp)
	},
	"counterpoint second species": func() []float64 {
		cp, _ := counterpoint.GenerateSecond(lineCantus, counterpoint.Config{})
		return lineIntValues(cp)
	},
	"melody.ChordTones": func() []float64 {
		key := scale.New(60, scale.Major)
		chords := []chord.Chord{chord.Diatonic(key, 0, 3), chord.Diatonic(key, 3, 3), chord.Diatonic(key, 4, 3)}
		return lineEventValues(melody.ChordTonesList(chords, 12, melody.Config{Scale: key}))
	},
	"line.Thin": func() []float64 {
		var events []note.Event
		for i := range 60 {
			events = append(events, note.Event{Pitch: float64(i), Start: float64(i), Duration: 1})
		}
		gate := function.New(func() bool { return genny.Float64() < 0.5 })
		return lineEventValues(line.Thin(events, gate))
	},
}

func TestLineGeneratorsAreSeeded(t *testing.T) {
	for name, f := range lineRandomGenerators {
		genny.Seed(31)
		a := f()
		genny.Seed(31)
		b := f()
		if len(a) == 0 {
			t.Errorf("%s: no values", name)
		}
		if !slices.Equal(a, b) {
			t.Errorf("%s: the same seed gave other values", name)
		}
		// another seed should differ for at least one of a few seeds
		differs := false
		for seed := uint64(32); seed < 40 && !differs; seed++ {
			genny.Seed(seed)
			differs = !slices.Equal(a, f())
		}
		if !differs {
			t.Errorf("%s: other seeds gave the same values", name)
		}
	}
}
