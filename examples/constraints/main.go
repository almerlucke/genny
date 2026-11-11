// constraints makes a short piece (16 bars) from the constraint, grammar, row and spectral tools:
//
//   - a grammar derives the phrase structure: how busy each bar of the lead is
//
//   - the lead is a melodic walk in A minor (constrain.ScaleIntervals), with Duffing
//     oscillator for its velocity and a chaotic pendulum for its timing
//
//   - the counter melody is a tone row from row.AllIntervals, in all four forms
//
//   - the pad is a spectral chord (chord.Harmonic), one partial per MIDI channel with a pitch bend
//     for the microtones
//
//     go run ./examples/constraints                       a text score
//     go run ./examples/constraints -seed 4 -mid out.mid  another piece, also as a MIDI file
package main

import (
	"flag"
	"fmt"
	"log"
	"math"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/chord"
	"github.com/almerlucke/genny/clock"
	"github.com/almerlucke/genny/constant"
	"github.com/almerlucke/genny/constrain"
	"github.com/almerlucke/genny/euclid"
	"github.com/almerlucke/genny/float/iter"
	"github.com/almerlucke/genny/float/iter/updaters/chaos"
	"github.com/almerlucke/genny/function"
	"github.com/almerlucke/genny/grammar"
	"github.com/almerlucke/genny/midi"
	"github.com/almerlucke/genny/note"
	"github.com/almerlucke/genny/rest"
	"github.com/almerlucke/genny/row"
	"github.com/almerlucke/genny/scale"
	"github.com/almerlucke/genny/sequence"
)

const (
	bars        = 16
	stepsPerBar = 16
	total       = bars * stepsPerBar
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

// barPulses derives the pulses per bar for the lead from a grammar of phrases
func barPulses() []int {
	g := grammar.New[int]().
		Rule("piece", grammar.Choice(1, grammar.Ref[int]("phrase"), grammar.Ref[int]("phrase"), grammar.Ref[int]("phrase"), grammar.Ref[int]("close"))).
		Rule("phrase",
			grammar.Choice(2, grammar.Ref[int]("statement"), grammar.Ref[int]("answer")),
			grammar.Choice(1, grammar.Ref[int]("statement"), grammar.Ref[int]("climb"))).
		Rule("statement", grammar.Seq(5, 7, 5, 5), grammar.Seq(5, 5, 7, 5)).
		Rule("answer", grammar.Seq(7, 9, 7, 5), grammar.Seq(7, 7, 9, 5)).
		Rule("climb", grammar.Seq(5, 7, 9, 11)).
		Rule("close", grammar.Seq(7, 5, 3, 1))
	phrase := g.Expand("piece", 5)
	var out []int
	for len(out) < bars {
		out = append(out, phrase...)
	}
	return out[:bars]
}

func main() {
	seed := flag.Uint64("seed", 1, "seed of the random stream: the same seed gives the same piece")
	midiFile := flag.String("mid", "", "write the piece to this MIDI file")
	flag.Parse()
	genny.Seed(*seed)

	key := scale.New(57, scale.Minor) // A minor

	// the structure: pulses per bar, the lead plays them as a Euclidean rhythm
	pulses := barPulses()
	var gate []bool
	for _, p := range pulses {
		gate = append(gate, euclid.Pattern(stepsPerBar, p, 0)...)
	}

	// the lead: a walk over the notes of A minor, in the range 57..84, with random intervals that
	// repeat in a cycle
	var allowed []int
	for p := 57; p <= 84; p++ {
		if key.Quantize(float64(p)) == float64(p) {
			allowed = append(allowed, p)
		}
	}
	var intervals []int
	for _, sign := range []int{1, -1, 1, 1, -1, -1, 1, -1} { // up and down, about evenly
		intervals = append(intervals, sign*(genny.Intn(4)+1))
	}
	walk := constrain.ScaleIntervals(69, intervals, allowed)
	pitch := function.New(func() float64 { return float64(walk.Generate()) })

	// modulation: a Duffing oscillator for the velocity, a pendulum angle for the timing
	duffing := iter.New([]float64{1, 0}, chaos.NewDuffing())
	pendulum := iter.New([]float64{0.2, 0}, chaos.NewPendulum())
	velocity := function.New(func() float64 {
		x := duffing.Generate()[0]
		for range 3 { // a slower oscillator: one step in four
			duffing.Generate()
		}
		return 0.65 + 0.2*x/1.5
	})
	offset := function.New(func() float64 { return 0.12 * pendulum.Generate()[0] / math.Pi })

	lead := note.Voice{
		Pitch:    rest.NewGated[float64](pitch, sequence.NewLoop(gate...)).Consume(false),
		Velocity: velocity,
		Length:   constant.New(1.5),
		Offset:   offset,
	}

	// the counter melody: a row in which every interval occurs once, prime, retrograde, inversion and
	// retrograde inversion, one note every four steps
	rows := row.AllIntervals(12, 50)
	pick := rows[genny.Intn(len(rows))]
	base := make([]float64, len(pick))
	for i, p := range pick {
		base[i] = float64(p)
	}
	var line []float64
	for _, op := range []row.Op{row.Prime, row.Retro, row.Invert, row.RetroInvert, row.Prime} {
		for _, p := range row.Apply(base, row.Form{Op: op}) {
			line = append(line, 52+math.Mod(math.Mod(p, 12)+12, 12))
		}
	}
	counter := note.Voice{
		Pitch:    rest.NewGated[float64](sequence.NewLoop(line...), clock.NewDivider(4, 0)).Consume(false),
		Velocity: constant.New(0.5),
		Length:   constant.New(5.0),
	}

	tracks := []midi.Track{
		{Name: "lead", Channel: 0, Program: 11, Events: lead.Render(total)},
		{Name: "row", Channel: 1, Program: 42, Events: counter.Render(total)},
	}

	// the pad: spectral chords on A, F, D and E (stretched partials), changing every four bars. Every partial
	// is a track of its own so that its pitch bend does not move the other partials
	fundamentals := []float64{55, 43.65, 36.71, 41.2}
	harmonics := []int{4, 5, 6, 7, 9, 11}
	var pad []chord.Chord
	for i, f := range fundamentals {
		pad = append(pad, chord.Harmonic(f, harmonics, 1+0.01*float64(i), i == 3))
	}
	partials := make([][]note.Event, len(harmonics))
	for i, c := range pad {
		for j, p := range c {
			partials[j] = append(partials[j], note.Event{
				Pitch: p, Velocity: 0.4 + 0.03*float64(j), Start: float64(i * 4 * stepsPerBar), Duration: 4 * stepsPerBar,
			})
		}
	}
	for j, events := range partials {
		tracks = append(tracks, midi.Track{
			Name: fmt.Sprintf("pad %d", j+1), Channel: 2 + j, Program: 89, Events: events, PitchBend: true,
		})
	}

	for bar := range bars {
		fmt.Printf("bar %2d (%2d pulses)\n", bar+1, pulses[bar])
		for _, t := range tracks {
			if len(t.Name) > 3 && t.Name[:3] == "pad" && t.Name != "pad 1" {
				continue // the pad is shown by its first partial and the chord
			}
			var s string
			for _, e := range t.Events {
				if int(e.Start)/stepsPerBar == bar {
					s += fmt.Sprintf(" %s@%.1f", name(e.Pitch), e.Start-float64(bar*stepsPerBar))
				}
			}
			if t.Name == "pad 1" && s != "" {
				s = ""
				for _, j := range partials {
					for _, e := range j {
						if int(e.Start)/stepsPerBar == bar {
							s += " " + name(e.Pitch)
						}
					}
				}
			}
			if s != "" {
				fmt.Printf("  %-6s%s\n", t.Name, s)
			}
		}
	}

	if *midiFile != "" {
		if err := midi.WriteFile(*midiFile, midi.File{Tempo: 96, StepsPerBeat: 4, Tracks: tracks}); err != nil {
			log.Fatal(err)
		}
		fmt.Println("written to", *midiFile)
	}
}
