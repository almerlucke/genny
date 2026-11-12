// evolve makes a piece of 24 bars (16 steps each) from chance, life and evolution:
//
//   - bars 1-24: a bass line and a hi-hat driven by Conway's Game of Life. A row of the grid is a bar,
//     a live cell is a hit (the bass takes its pitch from the column), and the grid steps one generation
//     every bar, so the rhythm drifts
//
//   - bars 1-24: a Poisson cloud on Messiaen's second mode that thickens to bar 12 and thins out again,
//     its pitch band widens and narrows (a tendency mask)
//
//   - bars 1-16: a melody that is evolved while it plays: it starts as a random phrase and every bar a
//     few generations towards an arch contour in A minor are made. The fitness is printed per bar
//
//   - bars 17-24: the same idea in the Bohlen-Pierce scale (13 steps in a tritave, 3/1, from a Scala
//     file); the bass moves over to that tuning too. These lines are on tracks with pitch bend
//
//     go run ./examples/evolve                       a text score
//     go run ./examples/evolve -seed 3 -mid out.mid  another piece, also as a MIDI file
package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/automaton"
	"github.com/almerlucke/genny/clock"
	"github.com/almerlucke/genny/constant"
	"github.com/almerlucke/genny/evolve"
	"github.com/almerlucke/genny/float/dist"
	"github.com/almerlucke/genny/float/tendency"
	"github.com/almerlucke/genny/midi"
	"github.com/almerlucke/genny/note"
	"github.com/almerlucke/genny/rest"
	"github.com/almerlucke/genny/scale"
	"github.com/almerlucke/genny/sequence"
	"github.com/almerlucke/genny/stochastic"
)

const (
	bars        = 24
	stepsPerBar = 16
	total       = bars * stepsPerBar
	microFrom   = 16 // the Bohlen-Pierce section starts at this bar (0 based)
)

var names = []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}

// name writes a pitch as a note name, with cents when it is microtonal: A4 or E5+2c
func name(p float64) string {
	n := int(math.Round(p))
	s := fmt.Sprintf("%s%d", names[((n%12)+12)%12], n/12-1)
	if cents := int(math.Round((p - float64(n)) * 100)); cents != 0 {
		s += fmt.Sprintf("%+dc", cents)
	}
	return s
}

// window keeps the events that start in [from, to), moved to start at 0
func window(events []note.Event, from, to int) []note.Event {
	var out []note.Event
	for _, e := range events {
		if int(e.Start) >= from && int(e.Start) < to {
			e.Start += float64(-from)
			out = append(out, e)
		}
	}
	return out
}

// at moves events later by a number of steps
func at(events []note.Event, steps int) []note.Event {
	out := make([]note.Event, len(events))
	for i, e := range events {
		e.Start += float64(steps)
		out[i] = e
	}
	return out
}

// before keeps the events that start before step
func before(events []note.Event, step int) []note.Event {
	var out []note.Event
	for _, e := range events {
		if int(e.Start) < step {
			out = append(out, e)
		}
	}
	return out
}

// evolved plays a melody that evolves while it plays: one phrase of 8 notes (2 steps each) per bar, and per bar
// gens generations towards an arch. It prints the generation and fitness of every phrase
func evolved(label string, s *scale.Scale, lo, hi, nbars, gens int) []note.Event {
	ga := evolve.New(evolve.Config{
		Population: 12, Length: 8, Min: lo, Max: hi,
		Fitness: evolve.Weighted(
			evolve.Weight(evolve.Contour(evolve.Arch(8)), 3),
			evolve.Weight(evolve.Smoothness(), 1),
			evolve.Weight(evolve.IntervalVariety(), 0.5)),
		MutationRate: 0.12, Elitism: 1,
	})
	fmt.Printf("%s: evolving towards an arch, %d generations a bar\n", label, gens)
	player := evolve.NewPlayer(ga, gens)
	var degrees []int
	for bar := range nbars {
		phrase := player.Phrase()
		fmt.Printf("  bar %2d: generation %2d, fitness %.3f, degrees %v\n", bar+1, ga.Generation(), ga.Best().Fitness, phrase)
		for range len(phrase) {
			degrees = append(degrees, player.Generate())
		}
	}
	v := note.Voice{
		Pitch:    rest.NewGated[float64](scale.NewDegrees(s, sequence.New(degrees...)), clock.NewDivider(2, 0)).Consume(false),
		Velocity: constant.New(0.75),
		Length:   constant.New(1.8),
	}
	return v.Render(nbars * stepsPerBar)
}

func main() {
	seed := flag.Uint64("seed", 1, "seed of the random stream: the same seed gives the same piece")
	midiFile := flag.String("mid", "", "write the piece to this MIDI file")
	flag.Parse()
	genny.Seed(*seed)

	minor := scale.New(57, scale.Minor) // A minor
	tuning, err := scale.ParseScl(strings.NewReader(scale.SclBohlenPierce))
	if err != nil {
		log.Fatal(err)
	}
	bp := tuning.Scale(57) // Bohlen-Pierce from A

	// the rhythm: a grid of Life, one row is a bar; the bass plays the live cells, the pitch is
	// the degree that belongs to the column
	life := automaton.NewLifeRandom(stepsPerBar, 16, automaton.Conway, 0.4)
	bassKey := scale.New(33, scale.Minor)
	bassPitch := automaton.NewLifePitches(life, 4, bassKey, 0, 4, 2, 4, 0, 5, 2, 4, 0, 4, 3, 4, 0, 6, 2, 4)
	bass := note.Voice{Pitch: bassPitch, Velocity: constant.New(0.8), Length: constant.New(0.9)}.Render(total)

	// the hi-hat: HighLife on another grid, read as a gate
	hatLife := automaton.NewLifeRandom(stepsPerBar, 16, automaton.HighLife, 0.4)
	hats := note.Voice{
		Pitch:    rest.NewGated[float64](constant.New(42.0), automaton.NewRowGate(hatLife, 2)),
		Velocity: constant.New(0.45),
		Length:   constant.New(0.5),
	}.Render(total)

	// the cloud: Poisson events in Messiaen's second mode. The density rises and falls (a sine), the band
	// of pitches opens in the first half and closes again in the second
	mode2 := scale.New(48, scale.Mode2)
	band := tendency.NewMask(false,
		tendency.Segment{N: 150, A1: 66, A2: 72, Z1: 48, Z2: 96},
		tendency.Segment{N: 150, A1: 48, A2: 96, Z1: 72, Z2: 78},
	)
	density := func(step int) float64 { return 0.1 + 1.1*math.Pow(math.Sin(math.Pi*float64(step)/total), 2) }
	cloud := stochastic.NewCloud(stochastic.Config{
		Density:  density,
		Pitch:    tendency.NewBetaValue(band, 2, 2),
		Velocity: dist.NewGaussian(0.4, 0.1).Clamp(0.15, 0.75),
		Duration: dist.NewExponential(3).Clamp(1, 10),
		Scale:    mode2,
	}).Render(total)

	// the melodies: A minor (degrees 0..14) for 16 bars, Bohlen-Pierce (0..16) for the last 8
	melody := evolved("melody in A minor", minor, 0, 14, microFrom, 3)
	microMelody := at(evolved("melody in Bohlen-Pierce", bp, 0, 16, bars-microFrom, 2), microFrom*stepsPerBar)

	// the bass moves to the Bohlen-Pierce tuning in the last section
	from := microFrom * stepsPerBar
	bassBP := tuning.Scale(33)
	var microBass []note.Event
	for _, e := range window(bass, from, total) {
		e.Pitch = bassBP.Quantize(e.Pitch)
		microBass = append(microBass, e)
	}
	microBass = at(microBass, from)

	tracks := []midi.Track{
		{Name: "melody", Channel: 0, Program: 11, Events: melody},
		{Name: "bass", Channel: 1, Program: 33, Events: before(bass, from)},
		{Name: "cloud", Channel: 2, Program: 98, Events: cloud},
		{Name: "hats", Channel: 9, Events: hats},
		{Name: "bp melody", Channel: 3, Program: 11, Events: microMelody, PitchBend: true},
		{Name: "bp bass", Channel: 4, Program: 33, Events: microBass, PitchBend: true},
	}

	fmt.Println("\nthe piece, per bar: the bass rhythm (from Life), the number of cloud notes, the melody")
	for bar := range bars {
		lo, hi := bar*stepsPerBar, (bar+1)*stepsPerBar
		pat := []byte(strings.Repeat(".", stepsPerBar))
		bassNotes := append(before(bass, from), microBass...)
		for _, e := range bassNotes {
			if int(e.Start) >= lo && int(e.Start) < hi {
				pat[int(e.Start)-lo] = 'x'
			}
		}
		n := len(window(cloud, lo, hi))
		var line []string
		for _, e := range append(append([]note.Event{}, melody...), microMelody...) {
			if int(e.Start) >= lo && int(e.Start) < hi {
				line = append(line, name(e.Pitch))
			}
		}
		fmt.Printf("bar %2d %s cloud %2d %s\n", bar+1, pat, n, strings.Join(line, " "))
	}

	for _, t := range tracks {
		fmt.Printf("%-10s %4d notes\n", t.Name, len(t.Events))
	}

	if *midiFile != "" {
		if err := midi.WriteFile(*midiFile, midi.File{Tempo: 108, StepsPerBeat: 4, Tracks: tracks}); err != nil {
			log.Fatal(err)
		}
		fmt.Println("written to", *midiFile)
	}
}
