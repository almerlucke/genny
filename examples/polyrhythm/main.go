// polyrhythm makes a piece of about 75 seconds for pitched percussion and a drone. Three sections
// play 3:2, 5:4 and 7:4 and are joined by metric modulations, so the tempo changes are audible: the
// triplet pulse of the 3:2 section becomes the beat of the 5:4 section (90 -> 135 bpm), and the
// eighth note of the 5:4 section becomes a quintuplet eighth in the 7:4 section (135 -> 108 bpm).
// Every time is an exact fraction of a beat (package polyrhythm), the MIDI file gets the tempo
// changes and the time signatures. It prints the plan.
//
//	go run ./examples/polyrhythm                       the plan
//	go run ./examples/polyrhythm -seed 7 -mid out.mid  also as a MIDI file
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/midi"
	"github.com/almerlucke/genny/note"
	"github.com/almerlucke/genny/polyrhythm"
)

const stepsPerBeat = 4

// roles of the tracks
const (
	low = iota
	ratioA
	ratioB
	cross
	nested
	drone
	numRoles
)

var trackInfo = [numRoles]struct {
	name          string
	channel, prog int
}{
	{"low", 0, 12},     // marimba
	{"layer a", 1, 11}, // vibraphone
	{"layer b", 2, 13}, // xylophone
	{"cross", 3, 14},   // tubular bells
	{"nested", 4, 9},   // glockenspiel
	{"drone", 5, 89},   // warm pad
}

type part struct {
	section polyrhythm.Section
	ratio   string
	layers  func() map[int]polyrhythm.Layer
	meter   *polyrhythm.Meter // cross rhythm as an additive meter
	drone   float64
}

func voice(l polyrhythm.Layer, pitch, lo, hi, gate float64) polyrhythm.Layer {
	l.Voice = polyrhythm.Voice{Pitch: pitch, Lo: lo, Hi: hi, Gate: gate}
	return l
}

func parts() []part {
	// the span of every layer is one bar
	bar := func(s polyrhythm.Section) polyrhythm.Frac { return s.BarBeats() }
	a := polyrhythm.Section{Name: "A", Bars: 20, Num: 2, Den: 4}
	b := polyrhythm.Section{Name: "B", Bars: 14, Num: 4, Den: 4,
		Into: polyrhythm.Equate(polyrhythm.F(2, 3), polyrhythm.Int(1))} // triplet pulse = new beat
	c := polyrhythm.Section{Name: "C", Bars: 14, Num: 4, Den: 4,
		Into: polyrhythm.Equate(polyrhythm.F(1, 2), polyrhythm.F(2, 5))} // eighth = quintuplet eighth
	m33 := polyrhythm.Additive(polyrhythm.F(1, 2), 3, 3, 2)
	m223 := polyrhythm.Additive(polyrhythm.F(1, 2), 2, 2, 3)
	return []part{
		{section: a, ratio: "3:2", drone: 36, layers: func() map[int]polyrhythm.Layer {
			return map[int]polyrhythm.Layer{
				low:    voice(polyrhythm.NewLayer(2, bar(a)), 43, 70, 105, 0.6),
				ratioA: voice(polyrhythm.NewLayer(3, bar(a)).WithEuclidAccent(3, 2, 0, 0.3), 67, 55, 100, 0.4),
			}
		}},
		{section: b, ratio: "5:4 + 3+3+2", meter: &m33, drone: 38, layers: func() map[int]polyrhythm.Layer {
			return map[int]polyrhythm.Layer{
				low:    voice(polyrhythm.NewLayer(4, bar(b)), 43, 65, 105, 0.5),
				ratioA: voice(polyrhythm.NewLayer(5, bar(b)).WithBarlow(0.2, 0, 5), 69, 55, 105, 0.4),
				ratioB: voice(polyrhythm.NewLayer(4, bar(b)).WithEuclid(4, 3, 1), 76, 50, 90, 0.3),
			}
		}},
		{section: c, ratio: "7:4 + 2+2+3 + 3:(2:5)", meter: &m223, drone: 34, layers: func() map[int]polyrhythm.Layer {
			return map[int]polyrhythm.Layer{
				low:    voice(polyrhythm.NewLayer(4, bar(c)), 41, 65, 105, 0.5),
				ratioA: voice(polyrhythm.NewLayer(7, bar(c)).WithBarlow(0.15, 0, 2, 2, 2), 65, 50, 105, 0.35),
				ratioB: voice(polyrhythm.NewLayer(7, bar(c)).WithEuclid(7, 5, 0), 77, 45, 85, 0.25),
				nested: voice(polyrhythm.NewLayer(2, polyrhythm.Int(5)).Nest(3), 89, 40, 70, 0.2),
			}
		}},
	}
}

func main() {
	seed := flag.Uint64("seed", 1, "seed of the random stream (the piece is exact and does not use chance)")
	midiFile := flag.String("mid", "", "write the piece to this MIDI file")
	flag.Parse()
	genny.Seed(*seed)

	ps := parts()
	secs := make([]polyrhythm.Section, len(ps))
	for i, p := range ps {
		secs[i] = p.section
	}
	plan := polyrhythm.NewPlan(polyrhythm.Int(90), secs...)

	events := make([][]note.Event, numRoles)
	fmt.Printf("%-3s %-22s %-6s %-9s %-4s %-10s %-9s %s\n", "sec", "ratio", "meter", "tempo", "bar", "start s", "length s", "modulation")
	for i, p := range ps {
		pl := plan[i]
		layers := p.layers()
		var all []polyrhythm.Layer
		for role := range numRoles {
			l, ok := layers[role]
			if !ok {
				continue
			}
			all = append(all, l)
			for _, e := range l.Events(polyrhythm.Int(0), pl.Length, stepsPerBeat) {
				e.Start += pl.Start.Float() * stepsPerBeat
				events[role] = append(events[role], e)
			}
		}
		if p.meter != nil {
			l := p.meter.Layer()
			l.Voice = polyrhythm.Voice{Pitch: 72 + float64(i)*5, Lo: 50, Hi: 100, Gate: 0.35}
			all = append(all, l)
			// the meter has its own bar; it plays across the section from its start
			for _, e := range l.Events(polyrhythm.Int(0), pl.Length, stepsPerBeat) {
				e.Start += pl.Start.Float() * stepsPerBeat
				events[cross] = append(events[cross], e)
			}
		}
		events[drone] = append(events[drone], note.Event{Pitch: p.drone, Velocity: 55,
			Start: pl.Start.Float() * stepsPerBeat, Duration: pl.Length.Float() * stepsPerBeat})
		events[drone] = append(events[drone], note.Event{Pitch: p.drone + 7, Velocity: 40,
			Start: pl.Start.Float() * stepsPerBeat, Duration: pl.Length.Float() * stepsPerBeat})

		mod := "-"
		if i > 0 {
			mod = fmt.Sprintf("%s beats = %s beats (x%s)", p.section.Into.Old, p.section.Into.New, p.section.Into.Factor())
		}
		secLen := pl.Length.MulInt(60).Div(pl.Tempo)
		fmt.Printf("%-3s %-22s %d/%-4d %-9s %-4d %-10.2f %-9.2f %s\n", pl.Name, p.ratio, pl.Num, pl.Den,
			pl.Tempo.String()+" bpm", pl.FirstBar+1, pl.Time.Float(), secLen.Float(), mod)
		fmt.Printf("    layers realign every %s beats (%s s)\n", polyrhythm.Cycle(all...), polyrhythm.Cycle(all...).MulInt(60).Div(pl.Tempo).String())
	}
	fmt.Printf("total: %s beats, %.2f seconds\n", plan.Beats(), plan.Seconds().Float())

	if *midiFile == "" {
		return
	}
	f := midi.File{
		Tempo:          plan.Start(),
		StepsPerBeat:   stepsPerBeat,
		TempoChanges:   plan.TempoChanges(stepsPerBeat),
		TimeSignatures: plan.TimeSignatures(stepsPerBeat),
	}
	for role, info := range trackInfo {
		f.Tracks = append(f.Tracks, midi.Track{Name: info.name, Channel: info.channel, Program: info.prog, Events: events[role]})
	}
	if err := midi.WriteFile(*midiFile, f); err != nil {
		log.Fatal(err)
	}
	fmt.Println("wrote", *midiFile)
}
