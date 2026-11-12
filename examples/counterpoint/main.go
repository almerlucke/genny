// counterpoint is a short piece of 24 bars built on whole lines of notes, in D Dorian:
//
//   - a cantus firmus of 12 notes, with a first species counterpoint above it and one below (package counterpoint)
//
//   - an isorhythmic tenor, a talea of 5 durations against a color of 7 pitches, which only
//     repeat together after 35 notes (package isorhythm)
//
//   - two ostinato voices that phase against each other, Reich style: one gains a step on every cycle (package phase)
//
//   - a melody on chord tones over a progression, as a canon at the fifth with entries after 2 bars
//     (packages melody, progression and line)
//
//     go run ./examples/counterpoint -seed 3 -mid line.mid
package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/chord"
	"github.com/almerlucke/genny/counterpoint"
	"github.com/almerlucke/genny/isorhythm"
	"github.com/almerlucke/genny/line"
	"github.com/almerlucke/genny/melody"
	"github.com/almerlucke/genny/midi"
	"github.com/almerlucke/genny/note"
	"github.com/almerlucke/genny/phase"
	"github.com/almerlucke/genny/progression"
	"github.com/almerlucke/genny/scale"
)

const (
	stepsPerBar = 16
	noteSteps   = 8 // one note of the cantus firmus: half a bar
)

var names = []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}

func name(p float64) string {
	n := int(math.Round(p))
	return fmt.Sprintf("%s%d", names[((n%12)+12)%12], n/12-1)
}

func intNames(ps []int) string {
	var s []string
	for _, p := range ps {
		s = append(s, name(float64(p)))
	}
	return strings.Join(s, " ")
}

func eventNames(ev []note.Event, max int) string {
	var s []string
	for i, e := range ev {
		if i == max {
			s = append(s, "...")
			break
		}
		s = append(s, name(e.Pitch))
	}
	return strings.Join(s, " ")
}

func bars(steps float64) float64 { return steps / stepsPerBar }

func main() {
	seed := flag.Int("seed", 3, "seed of the random stream: the same seed gives the same piece")
	midiFile := flag.String("mid", "", "write the piece to this MIDI file")
	flag.Parse()
	genny.Seed(uint64(*seed))

	var tracks []midi.Track
	add := func(title string, channel, program int, events []note.Event) {
		tracks = append(tracks, midi.Track{Name: title, Channel: channel, Program: program, Events: events})
		end := 0.0
		for _, e := range events {
			end = max(end, e.Start+e.Duration)
		}
		fmt.Printf("%-14s %3d notes, bars %5.2f to %5.2f: %s\n", title, len(events),
			bars(firstStart(events)), bars(end), eventNames(events, 12))
	}

	// ---- bars 0-6: a cantus firmus in D Dorian with a counterpoint above and one below
	cantus := []int{62, 65, 64, 67, 65, 69, 67, 65, 64, 62, 60, 62}
	fmt.Println("cantus firmus:", intNames(cantus))
	above, err := counterpoint.Generate(cantus, counterpoint.Config{Side: counterpoint.Above})
	report("above", cantus, above, err)
	below, err := counterpoint.Generate(cantus, counterpoint.Config{Side: counterpoint.Below, Low: 36, High: 72})
	report("below", cantus, below, err)
	fmt.Println()

	add("cantus", 0, 19, counterpoint.Line(cantus, noteSteps))
	add("cp above", 1, 73, counterpoint.Line(above, noteSteps))
	add("cp below", 2, 42, counterpoint.Line(below, noteSteps))

	// ---- an isorhythmic tenor from bar 3: a talea of 5 against a color of 7 (35 notes before it repeats)
	talea := []float64{6, 2, 4, -2, 2} // one bar of 16 steps, with a rest
	color := []float64{50, 53, 52, 55, 53, 57, 55}
	tenorLen := 21 * stepsPerBar
	tenor := line.Shift(isorhythm.Events(talea, color, float64(tenorLen)), 3*stepsPerBar)
	fmt.Printf("isorhythm: talea %d, color %d, repeats after %d notes\n", len(talea), len(color), isorhythm.Period(talea, color))
	add("tenor", 3, 52, tenor)

	// ---- bars 6-15: two ostinato voices; the second gains one step on each cycle of 12 steps
	ostinato := ostinatoPattern()
	cycles := phase.Cycles(ostinato, 1)
	p1, p2 := phase.Events(ostinato, cycles, 1)
	fmt.Printf("phasing: %d cycles of %g steps, back in unison after %d cycles\n", cycles, phase.Length(ostinato), cycles)
	add("phase 1", 4, 12, line.Shift(p1, 6*stepsPerBar))
	add("phase 2", 5, 12, line.Shift(p2, 6*stepsPerBar))

	// ---- bars 12-24: a melody on a progression, as a canon at the fifth, 2 bars apart
	key := scale.New(62, scale.Dorian)
	degrees := progression.New(progression.Classical, 0)
	var chords []chord.Chord
	var roman []string
	for range 8 {
		d := degrees.Generate()
		roman = append(roman, []string{"i", "ii", "III", "IV", "v", "vi", "VII"}[d%7])
		chords = append(chords, chord.Diatonic(key, d, 3))
	}
	fmt.Println("progression:", strings.Join(roman, " "))
	tune := melody.ChordTonesList(chords, 8, melody.Config{Scale: key, Low: 55, High: 70})
	voices := line.Canon(tune, 2*stepsPerBar, 7, 3)
	for i, v := range voices {
		add(fmt.Sprintf("canon %d", i+1), 6+i, []int{71, 68, 73}[i], line.Shift(v, 12*stepsPerBar))
	}

	var all []note.Event
	for _, t := range tracks {
		all = line.Merge(all, t.Events)
	}
	end := 0.0
	for _, e := range all {
		end = max(end, e.Start+e.Duration)
	}
	fmt.Printf("\n%d notes in %.0f bars\n", len(all), math.Ceil(bars(end)))

	if *midiFile != "" {
		if err := midi.WriteFile(*midiFile, midi.File{Tempo: 76, StepsPerBeat: 4, Tracks: tracks}); err != nil {
			log.Fatal(err)
		}
		fmt.Println("MIDI written to", *midiFile)
	}
}

// ostinatoPattern is a pattern of 12 steps over the notes of D Dorian
func ostinatoPattern() []note.Event {
	pitches := []float64{74, 76, 81, 76, 77, 76, 81, 76, 74, 76, 81, 77}
	pattern := make([]note.Event, len(pitches))
	for i, p := range pitches {
		pattern[i] = note.Event{Pitch: p, Velocity: 0.6, Start: float64(i), Duration: 1}
	}
	return pattern
}

func report(side string, cf, cp []int, err error) {
	fmt.Println("counterpoint", side+":", intNames(cp))
	if err != nil {
		fmt.Println("  no perfect solution:", err)
	}
	for _, v := range counterpoint.Check(cf, cp) {
		fmt.Println("  ", v)
	}
}

func firstStart(ev []note.Event) float64 {
	if len(ev) == 0 {
		return 0
	}
	return ev[0].Start
}
