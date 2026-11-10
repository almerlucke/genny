// music combines the music generators into a small piece: four bars of sixteenth notes with chords
// from a progression (voice led), a bass on a Euclidean rhythm and a lead that is a mutating phrase in
// the scale, thinned out, swung and ratcheted. It prints the notes and can write a MIDI file.
//
//	go run ./examples/music                       a text score
//	go run ./examples/music -seed 7 -mid out.mid  another piece, also as a MIDI file
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/chord"
	"github.com/almerlucke/genny/clock"
	"github.com/almerlucke/genny/constant"
	"github.com/almerlucke/genny/euclid"
	"github.com/almerlucke/genny/groove"
	"github.com/almerlucke/genny/midi"
	"github.com/almerlucke/genny/mutate"
	"github.com/almerlucke/genny/note"
	"github.com/almerlucke/genny/progression"
	"github.com/almerlucke/genny/ratchet"
	"github.com/almerlucke/genny/rest"
	"github.com/almerlucke/genny/scale"
	"github.com/almerlucke/genny/sequence"
)

const (
	bars        = 4
	stepsPerBar = 16
)

var names = []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}

func name(p float64) string {
	n := int(p + 0.5)
	return fmt.Sprintf("%s%d", names[n%12], n/12-1)
}

func main() {
	seed := flag.Uint64("seed", 1, "seed of the random stream: the same seed gives the same piece")
	midiFile := flag.String("mid", "", "write the piece to this MIDI file")
	flag.Parse()
	genny.Seed(*seed)

	key := scale.New(57, scale.Minor) // A minor

	// one chord per bar from a progression, with the voices moving as little as possible
	degrees := progression.New(progression.Pop, 0)
	leader := chord.NewVoiceLeader(55, 76)
	var barChords [][]float64
	var roots []float64
	for range bars {
		degree := degrees.Generate()
		barChords = append(barChords, leader.Transform(chord.Diatonic(key, degree, 3)))
		roots = append(roots, key.Degree(degree)-12)
	}

	// the chords sound at the start of each bar, and last the whole bar
	pad := note.Voice{
		Chord: rest.NewGated[[]float64](
			sequence.NewLoop(barChords...),
			clock.NewDivider(stepsPerBar, 0),
		).Consume(false),
		Velocity: constant.New(0.55),
		Length:   constant.New(float64(stepsPerBar)),
	}

	// the bass plays the root of the bar on a Euclidean rhythm
	var bassPerStep []float64
	for _, r := range roots {
		for range stepsPerBar {
			bassPerStep = append(bassPerStep, r)
		}
	}
	bass := note.Voice{
		Pitch:    rest.NewGated[float64](sequence.NewLoop(bassPerStep...), euclid.New(stepsPerBar, 5, 0)),
		Velocity: groove.NewAccent(3, 3, 3, 3, 4),
		Length:   constant.New(1.5),
	}

	// the lead: a phrase of scale degrees that mutates, with rests, a little swing and the odd roll
	phrase := mutate.New([]int{7, 9, 11, 9, 8, 7, 4, 7}, mutate.Config[int]{
		SwapProb:    0.6,
		RotateProb:  0.2,
		ReplaceProb: 0.3,
		Replace:     func() int { return genny.Intn(12) + 2 },
	})
	lead := note.Voice{
		Pitch:    rest.NewProbability[float64](scale.NewDegrees(key, phrase), 0.65),
		Velocity: groove.Jitter(groove.NewAccent(4, 4, 4, 4), 0.05),
		Length:   constant.New(1.0),
		Offset:   groove.NewSwing(0.15).Humanize(0.01),
		Ratchet:  ratchet.NewRandom(0.12, 3),
	}

	total := bars * stepsPerBar
	tracks := []midi.Track{
		{Name: "pad", Channel: 0, Program: 89, Events: pad.Render(total)},
		{Name: "bass", Channel: 1, Program: 33, Events: bass.Render(total)},
		{Name: "lead", Channel: 2, Program: 81, Events: lead.Render(total)},
	}

	for bar := range bars {
		fmt.Printf("bar %d\n", bar+1)
		for _, t := range tracks {
			fmt.Printf("  %-5s", t.Name)
			for _, e := range t.Events {
				if int(e.Start)/stepsPerBar == bar {
					fmt.Printf(" %s@%.2f", name(e.Pitch), e.Start-float64(bar*stepsPerBar))
				}
			}
			fmt.Println()
		}
	}

	if *midiFile != "" {
		if err := midi.WriteFile(*midiFile, midi.File{Tempo: 110, StepsPerBeat: 4, Tracks: tracks}); err != nil {
			log.Fatal(err)
		}
		fmt.Println("written to", *midiFile)
	}
}
