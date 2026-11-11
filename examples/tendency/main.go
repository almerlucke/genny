// tendency builds a piece of 20 bars with tendency masks and distributions: a lead whose pitches
// are drawn from a band that widens and narrows, quantized to A minor; a bass that picks its notes by index
// with weights that change every bar; bells chosen from a pentatonic scale by a mask in percentages and
// played when a Lehmer generator says so. Velocities, lengths and timing come from beta, Weibull and
// pink noise. It prints the notes and can write a MIDI file.
//
//	go run ./examples/tendency                       a text score
//	go run ./examples/tendency -seed 7 -mid out.mid  another piece, also as a MIDI file
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/choose"
	"github.com/almerlucke/genny/euclid"
	"github.com/almerlucke/genny/float/dist"
	"github.com/almerlucke/genny/float/tendency"
	"github.com/almerlucke/genny/function"
	"github.com/almerlucke/genny/midi"
	"github.com/almerlucke/genny/note"
	"github.com/almerlucke/genny/rest"
	"github.com/almerlucke/genny/scale"
	"github.com/almerlucke/genny/transform"
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

// perBar makes a segment that lasts a number of bars, with the band at the start and at the end
func perBar(nbars int, a1, a2, z1, z2 float64) tendency.Segment {
	return tendency.Segment{N: nbars * stepsPerBar, A1: a1, A2: a2, Z1: z1, Z2: z2}
}

func main() {
	seed := flag.Uint64("seed", 1, "seed of the random stream: the same seed gives the same piece")
	midiFile := flag.String("mid", "", "write the piece to this MIDI file")
	flag.Parse()
	genny.Seed(*seed)

	key := scale.New(57, scale.Minor) // A minor

	// the lead: a narrow band around A that opens to four octaves, closes on middle C, opens a
	// little less, and comes home to A. Values in the middle of the band are more likely (beta 2, 2)
	pitchMask := tendency.NewMask(false,
		perBar(4, 69, 69, 67, 72),
		perBar(6, 67, 72, 40, 96),
		perBar(4, 40, 96, 60, 62),
		perBar(4, 60, 62, 52, 84),
		perBar(2, 52, 84, 69, 69),
	)
	lead := note.Voice{
		Pitch: rest.NewProbability[float64](
			transform.New[float64](tendency.NewBetaValue(pitchMask, 2, 2), key.Quantize),
			0.7),
		Velocity: dist.NewBeta(2, 3).Scale(0.4, 1),
		Length:   dist.NewWeibull(0.5, 1.5, 1.2).Clamp(0.5, 6),
		Offset:   transform.New[float64](dist.NewPink(6), func(p float64) float64 { return p * 0.2 }),
	}

	// the bass: on a Euclidean rhythm, picks a note of the scale low in the register. The weights
	// change every 16 picks (one bar): the root, then the fifth and fourth, then the third
	bassNotes := []float64{33, 40, 38, 36, 43}
	bass := note.Voice{
		Pitch: rest.NewGated[float64](
			choose.NewChangingWeights(bassNotes, stepsPerBar,
				[]float64{8, 1, 1, 0, 0},
				[]float64{3, 4, 2, 0, 1},
				[]float64{1, 0, 2, 6, 1},
				[]float64{5, 3, 0, 0, 2}),
			euclid.New(stepsPerBar, 4, 0)),
		Velocity: dist.NewBeta(4, 2).Scale(0.5, 0.9),
		Length:   dist.NewWeibull(1, 2, 2).Clamp(1, 5),
	}

	// the bells: a pentatonic scale, the mask picks from the lower part of the list at first and
	// moves to the top; played when a Lehmer generator (repeatable by its own seed) is below 0.25
	pent := scale.New(69, scale.MinorPentatonic)
	var bellNotes []float64
	for i := range 15 {
		bellNotes = append(bellNotes, pent.Degree(i))
	}
	bellMask := tendency.NewMask(false,
		perBar(10, 0, 30, 70, 100),
		perBar(10, 70, 100, 0, 30),
	)
	lehmer := dist.NewLehmer(*seed)
	bells := note.Voice{
		Pitch: rest.NewGated[float64](
			tendency.NewChoice(bellMask, bellNotes),
			function.New(func() bool { return lehmer.Generate() < 0.25 })),
		Velocity: dist.NewBeta(2, 5).Scale(0.3, 0.8),
		Length:   dist.NewCauchy(3, 1).Retry(1, 12, 10),
	}

	total := bars * stepsPerBar
	tracks := []midi.Track{
		{Name: "bass", Channel: 0, Program: 33, Events: bass.Render(total)},
		{Name: "lead", Channel: 1, Program: 81, Events: lead.Render(total)},
		{Name: "bells", Channel: 2, Program: 10, Events: bells.Render(total)},
	}

	for bar := range bars {
		fmt.Printf("bar %2d\n", bar+1)
		for _, t := range tracks {
			fmt.Printf("  %-5s", t.Name)
			for _, e := range t.Events {
				if int(e.Start)/stepsPerBar == bar {
					fmt.Printf(" %s@%.1f", name(e.Pitch), e.Start-float64(bar*stepsPerBar))
				}
			}
			fmt.Println()
		}
	}

	if *midiFile != "" {
		if err := midi.WriteFile(*midiFile, midi.File{Tempo: 100, StepsPerBeat: 4, Tracks: tracks}); err != nil {
			log.Fatal(err)
		}
		fmt.Println("written to", *midiFile)
	}
}
