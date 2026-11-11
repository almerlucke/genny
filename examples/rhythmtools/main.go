// rhythmtools makes a short piece of 20 bars with the rhythm tools: a union and intersection of sieves
// gives the pitch set, the metric indispensability of the pulses thins a kick, a hi-hat and a bass so
// the rhythm fades in and out, and recursively split durations make the lead. It prints the notes and
// can write a MIDI file.
//
//	go run ./examples/rhythmtools                       a text score
//	go run ./examples/rhythmtools -seed 7 -mid out.mid  another piece, also as a MIDI file
package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/constant"
	"github.com/almerlucke/genny/function"
	"github.com/almerlucke/genny/listtools"
	"github.com/almerlucke/genny/midi"
	"github.com/almerlucke/genny/note"
	"github.com/almerlucke/genny/pulse"
	"github.com/almerlucke/genny/rest"
	"github.com/almerlucke/genny/sequence"
	"github.com/almerlucke/genny/sieve"
	"github.com/almerlucke/genny/split"
)

const (
	bars        = 20
	stepsPerBar = 16
)

var names = []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}

func name(p float64) string {
	n := int(p + 0.5)
	return fmt.Sprintf("%s%d", names[n%12], n/12-1)
}

// fade is the thinning level for a bar: sparse at first, dense in the middle, sparse at the end.
// high is the level of the empty bars (16 or more means no pulse at all), low the densest one
func fade(bar int, high, low float64) float64 {
	pos := float64(bar) / float64(bars-1)         // 0..1
	tri := 1 - math.Abs(2*pos-1)                  // 0, up to 1 in the middle, down to 0
	return math.Round(high - (high-low)*tri*1.25) // a little beyond the edge, so the middle stays full
}

// levels is a generator of the thinning level per pulse, it changes every bar
func levels(high, low float64) *function.Function[float64] {
	step := 0
	return function.New(func() float64 {
		level := fade(min(step/stepsPerBar, bars-1), high, low)
		step++
		return level
	})
}

func main() {
	seed := flag.Uint64("seed", 1, "seed of the random stream: the same seed gives the same piece")
	midiFile := flag.String("mid", "", "write the piece to this MIDI file")
	flag.Parse()
	genny.Seed(*seed)

	// pitch classes from sieves: thirds of five from 2, fours, and the odd multiples of three
	fives := sieve.New(5, 2, 0, 11)
	fours := sieve.New(4, 0, 0, 11)
	threes := sieve.New(3, 0, 0, 11).Intersection(sieve.New(2, 1, 0, 11))
	pcs := fives.Union(fours).Union(threes)
	fmt.Println("pitch classes", pcs.Values())

	// three octaves of it above A2, as one sieve of residues mod 12
	var pitchSet sieve.Sieve
	for _, pc := range pcs.Values() {
		pitchSet = pitchSet.Union(sieve.New(12, pc, 0, 35))
	}
	pitches := sieve.Scale(pitchSet, 45)

	// the meter: sixteenth notes in 4/4 and how much each pulse matters
	ind := pulse.Hierarchy(2, 2, 2, 2)
	accent := pulse.Velocity(ind, 0.45, 1)

	// the kick and hi-hat fade in and out with the level, the bass keeps to the strong pulses
	kick := note.Voice{
		Pitch:    rest.NewGated[float64](constant.New(36.0), pulse.NewGate(ind, levels(16, 4))),
		Velocity: sequence.NewLoop(accent...),
		Length:   constant.New(1.0),
	}
	hat := note.Voice{
		Pitch:    rest.NewGated[float64](constant.New(42.0), pulse.NewGate(ind, levels(18, 0))),
		Velocity: sequence.NewLoop(accent...),
		Length:   constant.New(0.5),
	}
	// one bass note per bar, from the pitch set one octave down
	roots := listtools.NewJump(pitches[:7], 1, 2, 3)
	bass := note.Voice{
		Pitch: rest.NewGated[float64](
			transpose(listtools.NewGroupN[float64](roots, stepsPerBar), -12),
			pulse.NewConstGate(ind, 12)),
		Velocity: sequence.NewLoop(accent...),
		Length:   constant.New(2.0),
	}

	// the lead: every bar is split into durations, some of them rests, the pitches jump through the set
	melody := listtools.NewJump(pitches[6:], 3, 1, 2)
	var lead []note.Event
	for bar := 4; bar < bars; bar++ {
		parts := split.Split(stepsPerBar, 1, 0.75+0.012*float64(bar-4), 2)
		rests := listtools.NewPlusMin(sequence.NewLoop(parts...), 0.25)
		pos := float64(bar * stepsPerBar)
		for range parts {
			d := rests.Generate()
			if d > 0 {
				step := int(pos) % stepsPerBar
				lead = append(lead, note.Event{Pitch: melody.Generate(), Velocity: accent[step], Start: pos, Duration: d * 0.9})
			}
			pos += math.Abs(d)
		}
	}

	total := bars * stepsPerBar
	tracks := []midi.Track{
		{Name: "kick", Channel: 9, Events: kick.Render(total)},
		{Name: "hat", Channel: 9, Events: hat.Render(total)},
		{Name: "bass", Channel: 1, Program: 38, Events: bass.Render(total)},
		{Name: "lead", Channel: 2, Program: 81, Events: lead},
	}

	for bar := range bars {
		fmt.Printf("bar %2d  level %2.0f  kick %s  hat %s", bar+1, fade(bar, 16, 4),
			steps(tracks[0].Events, bar), steps(tracks[1].Events, bar))
		fmt.Printf("  bass %-4s lead", barNotes(tracks[2].Events, bar, 1))
		fmt.Println(barNotes(tracks[3].Events, bar, 8))
	}

	if *midiFile != "" {
		if err := midi.WriteFile(*midiFile, midi.File{Tempo: 112, StepsPerBeat: 4, Tracks: tracks}); err != nil {
			log.Fatal(err)
		}
		fmt.Println("written to", *midiFile)
	}
}

// transpose shifts a generator of pitches
func transpose(g genny.Generator[float64], semitones float64) genny.Generator[float64] {
	return &shifted{g, semitones}
}

type shifted struct {
	gen   genny.Generator[float64]
	shift float64
}

func (s *shifted) Generate() float64 { return s.gen.Generate() + s.shift }
func (s *shifted) Continuous() bool  { return s.gen.Continuous() }
func (s *shifted) Done() bool        { return s.gen.Done() }
func (s *shifted) Reset()            { s.gen.Reset() }

// steps draws the events of a bar as x and .
func steps(events []note.Event, bar int) string {
	row := []byte(strings.Repeat(".", stepsPerBar))
	for _, e := range events {
		if int(e.Start)/stepsPerBar == bar {
			row[int(e.Start)%stepsPerBar] = 'x'
		}
	}
	return string(row)
}

// barNotes lists the first notes of a bar as name@step
func barNotes(events []note.Event, bar, limit int) string {
	var parts []string
	for _, e := range events {
		if int(e.Start)/stepsPerBar == bar && len(parts) < limit {
			parts = append(parts, fmt.Sprintf(" %s@%.0f", name(e.Pitch), e.Start-float64(bar*stepsPerBar)))
		}
	}
	return strings.Join(parts, "")
}
