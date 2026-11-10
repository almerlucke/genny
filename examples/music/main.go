// music combines the music generators into a small piece and prints it as text: four bars of
// sixteenth notes, with chords from a progression, a bass on a Euclidean rhythm, and a lead that
// is a mutating phrase in the scale, thinned out and swung.
//
//	go run ./examples/music        (add a number to hear another piece: go run ./examples/music 7)
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/chord"
	"github.com/almerlucke/genny/euclid"
	"github.com/almerlucke/genny/groove"
	"github.com/almerlucke/genny/mutate"
	"github.com/almerlucke/genny/progression"
	"github.com/almerlucke/genny/rest"
	"github.com/almerlucke/genny/scale"
)

var names = []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}

func note(p float64) string {
	n := int(p + 0.5)
	return fmt.Sprintf("%s%d", names[n%12], n/12-1)
}

func main() {
	seed := uint64(1)
	if len(os.Args) > 1 {
		if v, err := strconv.ParseUint(os.Args[1], 10, 64); err == nil {
			seed = v
		}
	}
	genny.Seed(seed)

	key := scale.New(57, scale.Minor) // A minor

	// one chord per bar: a progression, as triads of the scale, with smooth voice leading
	degrees := progression.New(progression.Pop, 0)
	leader := chord.NewVoiceLeader(55, 76)

	// bass: root of the chord on a Euclidean rhythm
	bassRhythm := euclid.New(16, 5, 0)

	// lead: a phrase of scale degrees that mutates, with some rests and swing
	phrase := mutate.New([]int{7, 9, 11, 9, 8, 7, 4, 7}, mutate.Config[int]{
		SwapProb:    0.6,
		RotateProb:  0.2,
		ReplaceProb: 0.3,
		Replace:     func() int { return genny.Intn(12) + 2 },
	})
	lead := rest.NewProbability[int](phrase, 0.65)
	swing := groove.NewSwing(0.15).Humanize(0.01)

	for bar := range 4 {
		degree := degrees.Generate()
		triad := leader.Transform(chord.Diatonic(key, degree, 3))
		fmt.Printf("bar %d  degree %d  chord %v\n", bar+1, degree+1, notes(triad))

		bass, melody := "", ""
		for step := range 16 {
			if bassRhythm.Generate() {
				bass += "x"
			} else {
				bass += "."
			}
			off := swing.Generate()
			if n := lead.Generate(); n.Played() {
				melody += fmt.Sprintf(" %s(%+.2f)", note(key.Degree(n.Value)), off)
			} else if step%4 == 0 {
				melody += " ----"
			}
		}
		fmt.Printf("  bass %s  (%s)\n  lead%s\n", bass, note(key.Degree(degree)-12), melody)
	}
}

func notes(c []float64) []string {
	out := make([]string, len(c))
	for i, p := range c {
		out[i] = note(p)
	}
	return out
}
