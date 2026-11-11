// suite is a piece of 32 bars that uses most of the newer tools together, in A minor:
//
//   - grammar: derives how many pulses the lead plays in every bar
//
//   - tendency masks: the bells are drawn from a band that moves, and kept to the scale by an interval filter
//
//   - sieve: the pitch set of the closing run, and the snare gate
//
//   - pulse: the metric hierarchy thins the kick and hi-hat so they fade in and out
//
//   - split, listtools, choose.Index: the run is a recursively split bar, with rests and jumps through the sieve scale,
//     picking pitches with a Gaussian distribution on the index
//
//   - constrain: the lead is a walk over the scale, the bells never repeat a note
//
//   - row: an all-interval row (prime, retrograde, inversion, retrograde inversion) as counter melody
//
//   - chord: a spectral pad of stretched partials with pitch bend, every partial a track
//
//   - chaos: Duffing, pendulum and standard map modulate velocity and timing
//
//   - dist, choose.ChangingWeights, euclid: beta, Weibull, Cauchy and Lehmer for velocity, length and gates
//     in the bass, which changes its weights every bar
//
//     go run ./examples/suite -seed 3 -mid suite.mid
package main

import (
	"flag"
	"fmt"
	"log"
	"math"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/choose"
	"github.com/almerlucke/genny/chord"
	"github.com/almerlucke/genny/clock"
	"github.com/almerlucke/genny/constant"
	"github.com/almerlucke/genny/constrain"
	"github.com/almerlucke/genny/euclid"
	"github.com/almerlucke/genny/float/dist"
	"github.com/almerlucke/genny/float/iter"
	"github.com/almerlucke/genny/float/iter/updaters/chaos"
	"github.com/almerlucke/genny/float/tendency"
	"github.com/almerlucke/genny/function"
	"github.com/almerlucke/genny/grammar"
	"github.com/almerlucke/genny/listtools"
	"github.com/almerlucke/genny/midi"
	"github.com/almerlucke/genny/note"
	"github.com/almerlucke/genny/pulse"
	"github.com/almerlucke/genny/rest"
	"github.com/almerlucke/genny/row"
	"github.com/almerlucke/genny/scale"
	"github.com/almerlucke/genny/sequence"
	"github.com/almerlucke/genny/sieve"
	"github.com/almerlucke/genny/split"
	"github.com/almerlucke/genny/transform"
)

const (
	bars        = 32
	stepsPerBar = 16
	total       = bars * stepsPerBar
)

var names = []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}

func name(p float64) string {
	n := int(math.Round(p))
	s := fmt.Sprintf("%s%d", names[((n%12)+12)%12], n/12-1)
	if cents := int(math.Round((p - float64(n)) * 100)); cents != 0 {
		s += fmt.Sprintf("%+dc", cents)
	}
	return s
}

// window keeps the events that start in bars [from, to)
func window(events []note.Event, from, to int) []note.Event {
	var out []note.Event
	for _, e := range events {
		if e.Start >= float64(from*stepsPerBar) && e.Start < float64(to*stepsPerBar) {
			out = append(out, e)
		}
	}
	return out
}

// level is the thinning level of the drums per bar: sparse at the edges, dense in the middle
func level(bar int, high, low float64) float64 {
	t := 1 - math.Abs(float64(bar-16)/16) // 0 at the edges, 1 in the middle
	return math.Round(high - (high-low)*math.Min(1, t*1.4))
}

func levelGen(high, low float64) genny.Generator[float64] {
	step := 0
	return function.New(func() float64 {
		l := level(min(step/stepsPerBar, bars-1), high, low)
		step++
		return l
	})
}

// pulsesPerBar derives the lead density per bar from a grammar
func pulsesPerBar() []int {
	ref := grammar.Ref[int]
	g := grammar.New[int]().
		Rule("piece", grammar.Choice(1, ref("phrase"), ref("phrase"), ref("phrase"), ref("close"))).
		Rule("phrase",
			grammar.Choice(2, ref("statement"), ref("answer")),
			grammar.Choice(1, ref("statement"), ref("climb"))).
		Rule("statement", grammar.Seq(5, 7, 5, 5), grammar.Seq(5, 5, 7, 5)).
		Rule("answer", grammar.Seq(7, 9, 7, 5), grammar.Seq(7, 7, 9, 5)).
		Rule("climb", grammar.Seq(5, 7, 9, 11)).
		Rule("close", grammar.Seq(7, 5, 3, 1))
	var out []int
	for len(out) < bars {
		out = append(out, g.Expand("piece", 5)...)
	}
	return out[:bars]
}

func main() {
	seed := flag.Uint64("seed", 1, "seed of the random stream: the same seed gives the same piece")
	midiFile := flag.String("mid", "", "write the piece to this MIDI file")
	flag.Parse()
	genny.Seed(*seed)

	key := scale.New(57, scale.Minor) // A minor
	pulses := pulsesPerBar()

	// ---- modulation: Duffing (slow) for velocity, a pendulum and the standard map for timing
	duffing := iter.New([]float64{1, 0}, chaos.NewDuffing())
	pendulum := iter.New([]float64{0.2, 0}, chaos.NewPendulum())
	stdmap := iter.New([]float64{1, 0.5}, chaos.NewStandardMap(1.1))
	duffVel := function.New(func() float64 {
		x := duffing.Generate()[0]
		for range 3 {
			duffing.Generate()
		}
		return 0.65 + 0.2*x/1.5
	})
	pendOff := function.New(func() float64 { return 0.1 * pendulum.Generate()[0] / math.Pi })
	mapOff := function.New(func() float64 { return 0.08 * (stdmap.Generate()[0]/(2*math.Pi) - 0.5) })

	// ---- pad: a spectral chord on a new fundamental every four bars, every partial a track with pitch bend
	fundamentals := []float64{55, 43.65, 36.71, 41.2, 55, 49, 36.71, 55}
	harmonics := []int{4, 5, 6, 7, 9, 11}
	partials := make([][]note.Event, len(harmonics))
	var padChords []chord.Chord
	for i, f := range fundamentals {
		c := chord.Harmonic(f, harmonics, 1+0.008*float64(i), i == 3)
		padChords = append(padChords, c)
		for j, p := range c {
			partials[j] = append(partials[j], note.Event{
				Pitch: p, Velocity: 0.34 + 0.03*float64(j), Start: float64(i * 4 * stepsPerBar), Duration: 4 * stepsPerBar,
			})
		}
	}

	// ---- bass: Euclidean 16/5, notes picked with weights that change every bar
	bassNotes := []float64{33, 40, 38, 36, 43}
	bass := note.Voice{
		Pitch: rest.NewGated[float64](
			choose.NewChangingWeights(bassNotes, stepsPerBar,
				[]float64{8, 1, 1, 0, 0}, []float64{3, 4, 2, 0, 1}, []float64{1, 0, 2, 6, 1}, []float64{5, 3, 0, 0, 2}),
			euclid.New(stepsPerBar, 5, 0)),
		Velocity: dist.NewBeta(4, 2).Scale(0.5, 0.9),
		Length:   dist.NewWeibull(1, 2, 2).Clamp(1, 4),
	}

	// ---- bells: a tendency mask (beta) opening up over the piece, to the scale, no repeated note,
	// played when a Lehmer generator is below 0.3; the timing wanders with the standard map
	bellMask := tendency.NewMask(false,
		tendency.Segment{N: 16 * stepsPerBar, A1: 72, A2: 72, Z1: 64, Z2: 84},
		tendency.Segment{N: 16 * stepsPerBar, A1: 64, A2: 84, Z1: 76, Z2: 76},
	)
	lehmer := dist.NewLehmer(*seed)
	quantized := transform.New[float64](tendency.NewBetaValue(bellMask, 2, 2), key.Quantize)
	bellPitch := function.New(func() float64 { return quantized.Generate() })
	bells := note.Voice{
		Pitch: rest.NewGated[float64](
			constrain.RemoveRepetitions[float64](bellPitch),
			function.New(func() bool { return lehmer.Generate() < 0.3 })),
		Velocity: dist.NewBeta(2, 5).Scale(0.3, 0.75),
		Length:   dist.NewCauchy(3, 1).Retry(1, 10, 10),
		Offset:   mapOff,
	}

	// ---- lead: a walk over A minor in 57..84 (cycling random intervals), as many pulses per bar as the
	// grammar says, played from bar 9 to bar 25
	var allowed []int
	for p := 57; p <= 84; p++ {
		if key.Quantize(float64(p)) == float64(p) {
			allowed = append(allowed, p)
		}
	}
	var intervals []int
	for _, sign := range []int{1, -1, 1, 1, -1, -1, 1, -1} {
		intervals = append(intervals, sign*(genny.Intn(4)+1))
	}
	walk := constrain.ScaleIntervals(69, intervals, allowed)
	var gate []bool
	for _, p := range pulses {
		gate = append(gate, euclid.Pattern(stepsPerBar, p, 0)...)
	}
	lead := note.Voice{
		Pitch:    rest.NewGated[float64](function.New(func() float64 { return float64(walk.Generate()) }), sequence.NewLoop(gate...)).Consume(false),
		Velocity: duffVel,
		Length:   constant.New(1.5),
		Offset:   pendOff,
	}

	// ---- counter melody: an all-interval row in four forms, one note per four steps, from bar 13
	rows := row.AllIntervals(12, 50)
	pick := rows[genny.Intn(len(rows))]
	base := make([]float64, len(pick))
	for i, p := range pick {
		base[i] = float64(p)
	}
	var line []float64
	for _, op := range []row.Op{row.Prime, row.Retro, row.Invert, row.RetroInvert, row.Prime, row.Retro} {
		for _, p := range row.Apply(base, row.Form{Op: op}) {
			line = append(line, 52+math.Mod(math.Mod(p, 12)+12, 12))
		}
	}
	counter := note.Voice{
		Pitch:    rest.NewGated[float64](sequence.NewLoop(line...), clock.NewDivider(4, 0)).Consume(false),
		Velocity: constant.New(0.5),
		Length:   constant.New(5.0),
	}

	// ---- drums: kick and hat thinned by the meter, snare on a sieve gate
	ind := pulse.Hierarchy(2, 2, 2, 2)
	accent := pulse.Velocity(ind, 0.45, 1)
	kick := note.Voice{
		Pitch:    rest.NewGated[float64](constant.New(36.0), pulse.NewGate(ind, levelGen(16, 4))),
		Velocity: sequence.NewLoop(accent...), Length: constant.New(1.0),
	}
	hat := note.Voice{
		Pitch:    rest.NewGated[float64](constant.New(42.0), pulse.NewGate(ind, levelGen(18, 0))),
		Velocity: sequence.NewLoop(accent...), Length: constant.New(0.5),
	}
	snare := note.Voice{
		Pitch:    rest.NewGated[float64](constant.New(38.0), sieve.NewGate(sieve.New(8, 4, 0, 15), stepsPerBar)),
		Velocity: dist.NewBeta(5, 2).Scale(0.5, 0.9), Length: constant.New(1.0),
	}

	// ---- the run (bars 25-32): the pitch set is a sieve (the C major pitch classes are the complement of
	// the black keys), each bar is split recursively, some durations become rests, the pitches are picked with a
	// Gaussian on the index (the middle of the set is most likely) and the jump tool steps through the rhythm
	black := sieve.Of(0, 11, 1, 3, 6, 8, 10)
	pcs := black.Complement()
	var pitchSet sieve.Sieve
	for _, pc := range pcs.Values() {
		pitchSet = pitchSet.Union(sieve.New(12, pc, 0, 35))
	}
	scalePitches := sieve.Scale(pitchSet, 57)
	picker := choose.NewIndex(scalePitches, choose.Gaussian(float64(len(scalePitches))/2, float64(len(scalePitches))/6))
	var run []note.Event
	for bar := 24; bar < bars; bar++ {
		parts := split.Split(stepsPerBar, 1, 0.6+0.05*float64(bar-24), 2)
		durations := listtools.NewPlusMin(sequence.NewLoop(parts...), 0.2)
		stutter := listtools.NewStutter[float64](picker, 3, 1) // every third pitch twice
		pos := float64(bar * stepsPerBar)
		for range parts {
			d := durations.Generate()
			if d > 0 {
				run = append(run, note.Event{Pitch: stutter.Generate(), Velocity: accent[int(pos)%stepsPerBar], Start: pos, Duration: d * 0.9})
			}
			pos += math.Abs(d)
		}
	}

	// ---- assemble, with entries and exits
	tracks := []midi.Track{
		{Name: "bells", Channel: 2, Program: 10, Events: bells.Render(total)},
		{Name: "bass", Channel: 1, Program: 38, Events: window(bass.Render(total), 4, 29)},
		{Name: "lead", Channel: 0, Program: 81, Events: window(lead.Render(total), 8, 24)},
		{Name: "row", Channel: 10, Program: 42, Events: window(counter.Render(total), 12, 28)},
		{Name: "run", Channel: 11, Program: 73, Events: run},
		{Name: "kick", Channel: 9, Events: window(kick.Render(total), 8, 32)},
		{Name: "hat", Channel: 9, Events: window(hat.Render(total), 8, 32)},
		{Name: "snare", Channel: 9, Events: window(snare.Render(total), 16, 28)},
	}
	for j, events := range partials {
		tracks = append(tracks, midi.Track{Name: fmt.Sprintf("pad %d", j+1), Channel: 3 + j, Program: 89, Events: events, PitchBend: true})
	}

	for bar := range bars {
		fmt.Printf("bar %2d  lead pulses %2d  drum level %2.0f", bar+1, pulses[bar], level(bar, 16, 4))
		if bar%4 == 0 {
			fmt.Print("  pad")
			for _, p := range padChords[bar/4] {
				fmt.Print(" ", name(p))
			}
		}
		fmt.Println()
	}
	for _, t := range tracks {
		fmt.Printf("%-7s %4d notes\n", t.Name, len(t.Events))
	}

	if *midiFile != "" {
		if err := midi.WriteFile(*midiFile, midi.File{Tempo: 104, StepsPerBeat: 4, Tracks: tracks}); err != nil {
			log.Fatal(err)
		}
		fmt.Println("written to", *midiFile)
	}
}
