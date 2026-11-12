package main

import (
	"log"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/chord"
	"github.com/almerlucke/genny/counterpoint"
	"github.com/almerlucke/genny/float/export/plot"
	"github.com/almerlucke/genny/isorhythm"
	"github.com/almerlucke/genny/line"
	"github.com/almerlucke/genny/melody"
	"github.com/almerlucke/genny/note"
	"github.com/almerlucke/genny/phase"
	"github.com/almerlucke/genny/scale"
	"gonum.org/v1/plot/vg"
)

// lineRoll draws tracks as a piano roll to a PNG file
func lineRoll(dir, name string, tracks ...plot.Track) {
	if err := plot.Roll(tracks, 24*vg.Centimeter, 10*vg.Centimeter, path(dir, name)); err != nil {
		log.Fatalf("%s: %v", name, err)
	}
}

// plotsLine draws piano rolls of the line tools: counterpoint, a canon, isorhythm and phasing
func plotsLine(dir string) {
	genny.Seed(11)

	// a cantus firmus with a first species counterpoint above and below
	cantus := []int{62, 65, 64, 67, 65, 69, 67, 65, 64, 62, 60, 62}
	above, err := counterpoint.Generate(cantus, counterpoint.Config{})
	if err != nil {
		log.Print("counterpoint above: ", err)
	}
	below, err := counterpoint.Generate(cantus, counterpoint.Config{Side: counterpoint.Below, Low: 36, High: 72})
	if err != nil {
		log.Print("counterpoint below: ", err)
	}
	lineRoll(dir, "counterpoint.png",
		plot.Track{Name: "above", Events: counterpoint.Line(above, 4)},
		plot.Track{Name: "cantus firmus", Events: counterpoint.Line(cantus, 4)},
		plot.Track{Name: "below", Events: counterpoint.Line(below, 4)})

	// a melody on chord tones, as a canon at the fifth with entries after 2 bars
	key := scale.New(62, scale.Dorian)
	chords := []chord.Chord{
		chord.Diatonic(key, 0, 3), chord.Diatonic(key, 3, 3), chord.Diatonic(key, 4, 3), chord.Diatonic(key, 0, 3),
	}
	tune := melody.ChordTonesList(chords, 4, melody.Config{Scale: key, Low: 55, High: 70})
	var canon []plot.Track
	for i, v := range line.Canon(tune, 32, 7, 3) {
		canon = append(canon, plot.Track{Name: "voice " + string(rune('1'+i)), Events: v})
	}
	lineRoll(dir, "canon.png", canon...)

	// isorhythm: a talea of 5 against a color of 7 (the pitches only come back in the same order after 35
	// notes), and below it the rhythm alone, on one pitch
	talea := []float64{3, 1, 2, -1, 1}
	color := []float64{60, 63, 62, 67, 65, 69, 64}
	iso := isorhythm.Events(talea, color, 8*float64(len(color))*5/2)
	lineRoll(dir, "isorhythm.png",
		plot.Track{Name: "talea 5, color 7", Events: iso},
		plot.Track{Name: "talea only (color 60)", Events: line.Transpose(isorhythm.Events(talea, []float64{60}, 8*float64(len(color))*5/2), -14)})

	// phasing: the same pattern twice, the second voice gains a step per cycle (shown a quarter tone higher to keep the bars apart)
	var pattern []note.Event
	for i, p := range []float64{64, 66, 71, 73, 74, 66, 64, 73, 71, 66, 74, 73} {
		pattern = append(pattern, note.Event{Pitch: p, Velocity: 0.7, Start: float64(i), Duration: 1})
	}
	n := phase.Cycles(pattern, 1)
	a, b := phase.Events(pattern, n, 1)
	lineRoll(dir, "phasing.png",
		plot.Track{Name: "voice 1", Events: a},
		plot.Track{Name: "voice 2", Events: line.Transpose(b, 0.25)})
}
