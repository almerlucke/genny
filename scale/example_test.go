package scale_test

import (
	"fmt"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/float/rwalk"
	"github.com/almerlucke/genny/float/shape"
	"github.com/almerlucke/genny/scale"
)

// Pitches are MIDI notes: 60 is middle C
func ExampleScale_Degree() {
	cMajor := scale.New(60, scale.Major)
	for _, d := range []int{0, 2, 4, 7, 9, -1} {
		fmt.Print(cMajor.Degree(d), " ")
	}
	fmt.Println()
	// Output: 60 64 67 72 76 59
}

// Quantize snaps any pitch to the scale
func ExampleScale_Quantize() {
	pentatonic := scale.New(60, scale.MinorPentatonic)
	for _, p := range []float64{60.4, 61.6, 64.6, 66.2, 71} {
		fmt.Print(pentatonic.Quantize(p), " ")
	}
	fmt.Println()
	// Output: 60 63 65 67 70
}

// A scale is a shaper: it makes a random walk musical
func ExampleScale_Shape() {
	genny.Seed(5)
	walk := rwalk.New(60, 55, 72, 1.5, rwalk.Bounce)
	melody := shape.NewSingle(walk, scale.New(57, scale.Minor))
	for range 8 {
		fmt.Print(melody.Generate(), " ")
	}
	fmt.Println()
	// Output: 60 60 62 60 60 60 59 60
}

// A walk over scale degrees: mostly stepwise, now and then a leap
func ExampleWalk() {
	genny.Seed(7)
	w := scale.NewWalk(scale.New(60, scale.Major), 7, 0, 14, 1, 0.2, 4)
	for range 10 {
		fmt.Print(w.Generate(), " ")
	}
	fmt.Println()
	// Output: 71 69 62 64 62 64 62 60 62 60
}

// Just intonation: the major third is the pure 5/4, a bit lower than the equal tempered 4 semitones
func ExampleFromRatios() {
	justMajorChord := scale.FromRatios(60, 1, 5.0/4, 3.0/2)
	for _, d := range []int{0, 1, 2} {
		fmt.Printf("%.3f ", justMajorChord.Degree(d))
	}
	fmt.Println()
	// Output: 60.000 63.863 67.020
}

// Quarter tones: the pitches are in fractional semitones
func ExampleEqual() {
	q := scale.Equal(60, 24)
	for _, p := range []float64{60.2, 60.4, 60.9, 61.3} {
		fmt.Print(q.Quantize(p), " ")
	}
	fmt.Println()
	// Output: 60 60.5 61 61.5
}
