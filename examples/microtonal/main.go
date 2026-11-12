// microtonal plays chords and a melody in tunings that are not 12-tone equal, through ordinary MIDI
// (package microtonal): every note is a MIDI note with a 14 bit pitch bend, and chords use MPE, a channel
// per note, so the bends do not disturb each other. About 80 seconds in three parts:
//
//   - 31-EDO: seventh chords with the pure-ish thirds and the harmonic seventh of 31 equal steps to the
//     octave, and a melody that walks along them
//
//   - just intonation: a harmonic series on C, the partials entering one by one (up to the 11th), then the
//     same on G, with a melody on the partials 8..15
//
//   - Bohlen-Pierce: chords 3:5:7:9 in a scale that repeats at the tritave (3/1) instead of the octave
//
//     go run ./examples/microtonal                       a text score
//     go run ./examples/microtonal -seed 4 -mid out.mid  another piece, also as a MIDI file
//
// The file needs a synth that follows pitch bend per channel (an MPE synth, or any multitimbral synth with
// its bend range at 2 semitones), see the notes in package microtonal
package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/microtonal"
	"github.com/almerlucke/genny/midi"
	"github.com/almerlucke/genny/note"
	"github.com/almerlucke/genny/scale"
)

const (
	stepsPerBar = 16
	bars        = 24
	melodyChan  = 15
)

// chordSpec is a chord as scale degrees above a root degree
type chordSpec struct {
	root    int
	degrees []int
}

func (c chordSpec) at(i int) int { return c.root + c.degrees[i] }

// tones are the degrees of the chord modulo the length of the tuning
func (c chordSpec) tones(n int) []int {
	var out []int
	for _, d := range c.degrees {
		out = append(out, ((c.root+d)%n+n)%n)
	}
	return out
}

// melody walks along a tuning, one chord per bar, in monophonic notes that favour the tones of the chord.
// Degrees are counted from the root of the tuning
func melody(t microtonal.Tuning, chords []chordSpec, firstBar, lo, hi, start int, density float64) []note.Event {
	var events []note.Event
	cur := start
	for bar, c := range chords {
		tones := c.tones(t.Len())
		isTone := func(d int) bool {
			for _, x := range tones {
				if ((d%t.Len())+t.Len())%t.Len() == x {
					return true
				}
			}
			return false
		}
		for slot := 0; slot < stepsPerBar/2; {
			if genny.Float64() > density && slot > 0 {
				slot++
				continue
			}
			if slot == 0 || genny.Float64() < 0.35 {
				// jump to a chord tone near the current degree
				var near []int
				for d := max(cur-6, lo); d <= min(cur+6, hi); d++ {
					if isTone(d) {
						near = append(near, d)
					}
				}
				if len(near) > 0 {
					cur = near[genny.Intn(len(near))]
				}
			} else {
				cur = min(max(cur+(genny.Intn(3)+1)*(2*genny.Intn(2)-1), lo), hi)
			}
			slots := 1 + genny.Intn(3)
			if slot+slots > stepsPerBar/2 {
				slots = stepsPerBar/2 - slot
			}
			events = append(events, note.Event{
				Pitch:    t.Pitch(cur),
				Velocity: 0.6 + 0.25*genny.Float64(),
				Start:    float64((firstBar+bar)*stepsPerBar + slot*2),
				Duration: float64(slots*2) - 0.2,
			})
			slot += slots
		}
	}
	return events
}

// pad plays chords as a slight strum
func pad(t microtonal.Tuning, chords []chordSpec, firstBar, barsPerChord int, extra []int) []note.Event {
	var events []note.Event
	for i, c := range chords {
		start := float64((firstBar + i*barsPerChord) * stepsPerBar)
		for j := range c.degrees {
			events = append(events, note.Event{
				Pitch: t.Pitch(c.at(j)), Velocity: 0.45 + 0.1*genny.Float64(),
				Start: start + float64(j)*0.75, Duration: float64(barsPerChord*stepsPerBar) + 1,
			})
		}
		for _, d := range extra { // a bass note
			events = append(events, note.Event{Pitch: t.Pitch(c.root + d), Velocity: 0.55, Start: start, Duration: float64(barsPerChord*stepsPerBar) + 1})
		}
	}
	return events
}

// perBar repeats every chord for a number of bars
func perBar(chords []chordSpec, barsEach int) []chordSpec {
	var out []chordSpec
	for _, c := range chords {
		for range barsEach {
			out = append(out, c)
		}
	}
	return out
}

func describe(t microtonal.Tuning, c chordSpec) string {
	var parts []string
	for i := range c.degrees {
		b, _ := microtonal.Locate(t.Pitch(c.at(i)), 2)
		parts = append(parts, fmt.Sprintf("degree %d = %.2f cents (note %d %+.2f cents, bend %d)", c.at(i), t.Cent(c.at(i)), b.Note, b.Cents(2), b.Value))
	}
	return "    " + strings.Join(parts, "\n    ")
}

func main() {
	seed := flag.Uint64("seed", 1, "seed of the random stream: the same seed gives the same piece")
	midiFile := flag.String("mid", "", "write the piece to this MIDI file")
	flag.Parse()
	genny.Seed(*seed)

	// part 1: 31-EDO, bars 0..7
	e31 := microtonal.EDO(60, 31)
	maj, min7 := []int{0, 10, 18, 25}, []int{0, 8, 18, 26}
	prog31 := []chordSpec{{0, maj}, {13, maj}, {8, min7}, {18, maj}}
	e31lead := microtonal.EDO(72, 31)
	events31 := pad(e31, prog31, 0, 2, []int{-31})
	lead31 := melody(e31lead, perBar(prog31, 2), 0, -4, 20, 0, 0.75)

	// part 2: just intonation, bars 8..15: harmonic series on C (partials enter one by one) and then on G
	var just []note.Event
	partial := func(fund float64, n float64) float64 { return fund + 12*math.Log2(n) }
	for i, n := range []float64{1, 2, 3, 4, 5, 6, 7, 9, 11} {
		start := float64(8*stepsPerBar) + float64(i)*6
		just = append(just, note.Event{Pitch: partial(36, n), Velocity: 0.5, Start: start, Duration: float64(12*stepsPerBar) - start + 1})
	}
	gFund := 36 + 12*math.Log2(1.5)
	for i, n := range []float64{1, 2, 3, 4, 5, 6, 7, 9} {
		start := float64(12*stepsPerBar) + float64(i)*6
		just = append(just, note.Event{Pitch: partial(gFund, n), Velocity: 0.5, Start: start, Duration: float64(16*stepsPerBar) - start + 1})
	}
	harm := []float64{1, 9.0 / 8, 5.0 / 4, 11.0 / 8, 3.0 / 2, 13.0 / 8, 7.0 / 4, 15.0 / 8}
	jc := chordSpec{0, []int{0, 2, 4, 6}} // 1, 5/4, 3/2, 7/4
	jLeadC := microtonal.Just(72, 2, harm...)
	jLeadG := microtonal.Just(72+12*math.Log2(1.5), 2, harm...)
	leadJust := melody(jLeadC, perBar([]chordSpec{jc}, 4), 8, 0, 11, 4, 0.7)
	leadJust = append(leadJust, melody(jLeadG, perBar([]chordSpec{jc}, 4), 12, 0, 11, 4, 0.6)...)

	// part 3: Bohlen-Pierce, bars 16..23: 3:5:7:9 chords (degrees 0, 6, 10, 13 of the just scale)
	scl, err := scale.ParseScl(strings.NewReader(scale.SclBohlenPierce))
	if err != nil {
		log.Fatal(err)
	}
	bp := microtonal.FromScl(60, scl)
	bpLead := microtonal.FromScl(72, scl)
	bpChord := []int{0, 6, 10, 13}
	progBP := []chordSpec{{0, bpChord}, {4, bpChord}, {7, bpChord}, {0, bpChord}}
	eventsBP := pad(bp, progBP, 16, 2, []int{-13})
	leadBP := melody(bpLead, perBar(progBP, 2), 16, -3, 16, 3, 0.7)

	fmt.Printf("31-EDO: %d steps of %.2f cents; Bohlen-Pierce: %d steps to the tritave (%.2f cents)\n",
		e31.Len(), e31.Period/float64(e31.Len()), bp.Len(), bp.Period)
	fmt.Println("first chord (31-EDO seventh chord on C, degrees 0 10 18 25):")
	fmt.Println(describe(e31, prog31[0]))

	var padEvents, leadEvents []note.Event
	padEvents = append(append(append(padEvents, events31...), just...), eventsBP...)
	leadEvents = append(append(append(leadEvents, lead31...), leadJust...), leadBP...)

	// the pad uses MPE (channels 2..15), the lead a single channel 16
	pads, skipped := microtonal.Plan(padEvents, nil, microtonal.Strategy{MPE: true, Members: 14, Program: 89, Name: "pad"})
	lead := microtonal.Render(leadEvents, nil, microtonal.Strategy{Channel: melodyChan, Program: 80, Name: "lead"})
	tracks := append(pads, lead...)

	last := 0.0
	for _, e := range append(padEvents, leadEvents...) {
		last = max(last, e.Start+e.Duration)
	}
	fmt.Printf("%d pad notes on %d tracks, %d lead notes, %.0f seconds at 72 bpm (%d notes left out)\n",
		len(padEvents), len(pads), len(leadEvents), last/4*60/72, skipped)

	if *midiFile != "" {
		if err := midi.WriteFile(*midiFile, midi.File{Tempo: 72, StepsPerBeat: 4, Tracks: tracks}); err != nil {
			log.Fatal(err)
		}
		fmt.Println("written to", *midiFile)
	}
}
