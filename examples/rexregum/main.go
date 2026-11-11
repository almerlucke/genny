// rexregum is a piece in praise of Jesus Christ, King of kings, risen from the dead. It is built on numbers
// that mean something in the Bible:
//
//   - 3: the Trinity, and the three days in the tomb. The meter is 3 groups of 4, 12 steps (the twelve apostles
//     and the twelve tribes) to a bar; the rhythms are Euclidean patterns of 3 and 7 pulses in 12; the canon follows after 3 bars
//   - 7: completeness. Seven sections of seven bars, 49 bars in all (7 times 7); the theme is the Fibonacci series
//     on the 7 degrees of the scale
//   - 12: the pad chords are stacks of the overtones 4, 5, 6, 8, 10 and 12 of the series God put in every sounding string
//   - 153: the fish that were caught after the resurrection (John 21). Exactly 153 chimes ring from bar 15 on
//   - 40: the days between the resurrection and the ascension; the tempo is 40 + 26
//
// It goes from darkness to light: the first two sections are minor, on the undertones (the harmonic series upside
// down, a minor chord that sounds as if it is falling), three bars of nothing but the choir follow (the three days), then the stone
// is rolled away and everything turns to D major on the overtones. The piece ends with an Amen (IV to I).
//
//	go run ./examples/rexregum -seed 7 -mid rexregum.mid
package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"sort"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/chord"
	"github.com/almerlucke/genny/constant"
	"github.com/almerlucke/genny/constrain"
	"github.com/almerlucke/genny/euclid"
	"github.com/almerlucke/genny/float/dist"
	"github.com/almerlucke/genny/float/tendency"
	"github.com/almerlucke/genny/function"
	"github.com/almerlucke/genny/midi"
	"github.com/almerlucke/genny/note"
	"github.com/almerlucke/genny/numseq"
	"github.com/almerlucke/genny/pulse"
	"github.com/almerlucke/genny/rest"
	"github.com/almerlucke/genny/scale"
	"github.com/almerlucke/genny/sequence"
	"github.com/almerlucke/genny/transform"
)

const (
	stepsPerBar = 12
	sections    = 7
	barsPerSec  = 7
	bars        = sections * barsPerSec // 49
	sec         = barsPerSec * stepsPerBar
	total       = bars * stepsPerBar
	chimes      = 153
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

// window keeps the events that start in steps [from, to)
func window(events []note.Event, from, to int) []note.Event {
	var out []note.Event
	for _, e := range events {
		if e.Start >= float64(from) && e.Start < float64(to) {
			out = append(out, e)
		}
	}
	return out
}

// shift moves events later and up
func shift(events []note.Event, steps, semitones float64) []note.Event {
	out := make([]note.Event, len(events))
	for i, e := range events {
		e.Start += steps
		e.Pitch += semitones
		out[i] = e
	}
	return out
}

func main() {
	seed := flag.Uint64("seed", 7, "seed of the random stream: the same seed gives the same piece")
	midiFile := flag.String("mid", "", "write the piece to this MIDI file")
	flag.Parse()
	genny.Seed(*seed)

	dMajor := scale.New(62, scale.Major)
	dMinor := scale.New(62, scale.Minor)

	// ---- the pad: harmonic series chords, one partial per track (pitch bend makes the pure intervals).
	// Sections 1 and 2 use the undertones of D5 and A5 (dark, minor), after that the overtones of D, G and A
	type step struct {
		from, to    int
		fundamental float64
		under       bool
	}
	const (
		d2, g2, a2 = 73.416, 97.999, 110.0
		d5, a5     = 587.33, 880.0
	)
	chords := []step{
		{0, 1 * sec, d5, true},
		{1 * sec, 2 * sec, a5, true},
		{2 * sec, 3 * sec, d2, false},
		{3 * sec, 3*sec + sec/2, d2, false},
		{3*sec + sec/2, 4 * sec, a2, false},
		{4 * sec, 4*sec + sec/2, g2, false},
		{4*sec + sec/2, 5 * sec, d2, false},
		{5 * sec, 5*sec + sec/2, a2, false},
		{5*sec + sec/2, 6 * sec, d2, false},
		{6 * sec, 6*sec + sec/2, g2, false}, // IV
		{6*sec + sec/2, total, d2, false},   // I, Amen
	}
	overtones := []int{4, 5, 6, 8, 10, 12}
	undertones := []int{1, 2, 3, 4, 5, 6}
	padEvents := make([][]note.Event, 6)
	var padNames []string
	for _, c := range chords {
		harmonics := overtones
		if c.under {
			harmonics = undertones
		}
		ch := chord.Harmonic(c.fundamental, harmonics, 1, c.under)
		s := fmt.Sprint(c.from/stepsPerBar+1, ":")
		for j, p := range ch {
			padEvents[j] = append(padEvents[j], note.Event{
				Pitch: p, Velocity: 0.35 + 0.03*float64(j), Start: float64(c.from), Duration: float64(c.to - c.from),
			})
			s += " " + name(p)
		}
		padNames = append(padNames, s)
	}

	// ---- the bass pedal on the root of each chord, an organ
	var bass []note.Event
	for _, c := range chords {
		root := 38.0 // D2
		switch c.fundamental {
		case g2:
			root = 31 // G1
		case a2, a5:
			root = 33 // A1
		}
		if c.under {
			root += 12
		}
		bass = append(bass, note.Event{Pitch: root, Velocity: 0.55, Start: float64(c.from), Duration: float64(c.to - c.from)})
	}

	// ---- section 1, darkness: celesta notes at the Fibonacci steps (1 2 3 5 8 13 21 34 55), low in D minor
	var celesta []note.Event
	fib := numseq.NewFibonacci(1 << 30)
	fib.Generate() // 0
	for {
		f := fib.Generate()
		if f >= sec {
			break
		}
		celesta = append(celesta, note.Event{
			Pitch: dMinor.Degree(genny.Intn(5) - 3), Velocity: 0.35 + 0.25*genny.Float64(), Start: float64(f), Duration: 8,
		})
	}

	// ---- section 2, the passion: a slow lament, three notes a bar (Euclid 3 in 12) in D minor, as a walk
	// that only takes small steps; it is silent for the last three bars: three days
	var allowed []int
	for p := 50; p <= 74; p++ {
		if dMinor.Quantize(float64(p)) == float64(p) {
			allowed = append(allowed, p)
		}
	}
	walk := constrain.ScaleIntervals(69, []int{-1, -2, 1, -1, -2, -1, 2, -3, 1, -1, 3, -2}, allowed)
	lament := note.Voice{
		Pitch:    rest.NewGated[float64](function.New(func() float64 { return float64(walk.Generate()) }), euclid.New(stepsPerBar, 3, 0)).Consume(false),
		Velocity: dist.NewBeta(3, 3).Scale(0.35, 0.65),
		Length:   constant.New(5.0),
	}
	lamentEvents := window(shift(lament.Render(4*stepsPerBar), float64(sec), 0), sec, sec+4*stepsPerBar)

	// ---- sections 3 to 7, the resurrection: timpani on the Trinity rhythm (3 in 12) that swells over the third section
	// and keeps its force, D and A
	swell := tendency.NewMask(false,
		tendency.Segment{N: sec, A1: 0.25, A2: 0.4, Z1: 0.8, Z2: 1},
		tendency.Segment{N: 4 * sec, A1: 0.8, A2: 1, Z1: 0.8, Z2: 1},
	)
	timpaniPitch := sequence.NewLoop(38.0, 45.0, 38.0)
	timpani := note.Voice{
		Pitch:    rest.NewGated[float64](timpaniPitch, euclid.New(stepsPerBar, 3, 0)).Consume(false),
		Velocity: tendency.NewValue(swell),
		Length:   constant.New(3.0),
	}
	timpaniEvents := window(timpani.Render(total), 2*sec, 6*sec+6*stepsPerBar)
	// the last blow
	timpaniEvents = append(timpaniEvents, note.Event{Pitch: 38, Velocity: 1, Start: float64(total - stepsPerBar), Duration: stepsPerBar})

	// ---- section 4 and on: the theme, the Fibonacci series on the seven degrees of D major, played on seven
	// of the twelve steps. It becomes a canon after three bars. Accents from the metric hierarchy of 3 x 2 x 2
	ind := pulse.Hierarchy(3, 2, 2)
	accent := pulse.Velocity(ind, 0.55, 1)
	degrees := numseq.NewFibonacci(7)
	theme := note.Voice{
		Pitch:    rest.NewGated[float64](scale.NewDegrees(dMajor, degrees), euclid.New(stepsPerBar, 7, 0)).Consume(false),
		Velocity: sequence.NewLoop(accent...),
		Length:   constant.New(1.7),
	}
	themeEvents := window(theme.Render(total), 3*sec, total-2*stepsPerBar)
	canonEvents := window(shift(themeEvents, 3*stepsPerBar, 12), 4*sec, total-stepsPerBar)

	// ---- the 153 chimes: from bar 15 (the stone is rolled away) to the end, the pitches climb in a tendency band
	// quantized to D major, no pitch twice in a row; exactly 153 of the candidates stay
	band := tendency.NewMask(false,
		tendency.Segment{N: total - 2*sec, A1: 74, A2: 80, Z1: 86, Z2: 98},
	)
	quantized := transform.New[float64](tendency.NewBetaValue(band, 2, 2), dMajor.Quantize)
	pitches := constrain.RemoveRepetitions[float64](function.New(func() float64 { return quantized.Generate() }))
	var candidates []note.Event
	for t := 2 * sec; t < total-stepsPerBar; t++ {
		candidates = append(candidates, note.Event{Pitch: pitches.Generate(), Velocity: 0.3 + 0.35*genny.Float64(), Start: float64(t), Duration: 6})
	}
	idx := make([]int, len(candidates))
	for i := range idx {
		idx[i] = i
	}
	genny.Shuffle(len(idx), func(i, j int) { idx[i], idx[j] = idx[j], idx[i] })
	idx = idx[:chimes]
	sort.Ints(idx)
	var chimeEvents []note.Event
	for _, i := range idx {
		chimeEvents = append(chimeEvents, candidates[i])
	}

	tracks := []midi.Track{
		{Name: "organ pedal", Channel: 1, Program: 19, Events: bass},
		{Name: "celesta", Channel: 13, Program: 8, Events: celesta},
		{Name: "lament", Channel: 2, Program: 42, Events: lamentEvents},
		{Name: "timpani", Channel: 10, Program: 47, Events: timpaniEvents},
		{Name: "theme", Channel: 0, Program: 61, Events: themeEvents},
		{Name: "canon", Channel: 12, Program: 56, Events: canonEvents},
		{Name: "chimes", Channel: 11, Program: 14, Events: chimeEvents},
	}
	for j, events := range padEvents {
		tracks = append(tracks, midi.Track{Name: fmt.Sprintf("choir %d", j+1), Channel: 3 + j, Program: 52, Events: events, PitchBend: true})
	}

	fmt.Println("Rex Regum: 7 sections of 7 bars, 12 steps a bar")
	for _, s := range padNames {
		fmt.Println("  bar", s)
	}
	for _, t := range tracks {
		fmt.Printf("%-12s %4d notes\n", t.Name, len(t.Events))
	}
	fmt.Println("chimes:", len(chimeEvents), "(the fish of John 21)")

	if *midiFile != "" {
		if err := midi.WriteFile(*midiFile, midi.File{Tempo: 66, StepsPerBeat: 4, Tracks: tracks}); err != nil {
			log.Fatal(err)
		}
		fmt.Println("written to", *midiFile)
	}
}
