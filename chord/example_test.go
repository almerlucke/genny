package chord_test

import (
	"fmt"

	"github.com/almerlucke/genny/chord"
	"github.com/almerlucke/genny/scale"
	"github.com/almerlucke/genny/sequence"
	"github.com/almerlucke/genny/transform"
)

func ExampleChord() {
	c := chord.New(60, chord.Dominant7)
	fmt.Println(c)
	fmt.Println(c.Invert(1))
	fmt.Println(chord.New(60, chord.Major).Open())
	// Output:
	// [60 64 67 70]
	// [64 67 70 72]
	// [52 60 67]
}

// Triads built from the notes of a scale
func ExampleDiatonic() {
	cMajor := scale.New(60, scale.Major)
	for _, degree := range []int{0, 3, 4} { // I IV V
		fmt.Println(chord.Diatonic(cMajor, degree, 3))
	}
	// Output:
	// [60 64 67]
	// [65 69 72]
	// [67 71 74]
}

// Without voice leading every chord is in root position and the voices jump around. The voice
// leader picks the inversion closest to the previous chord
func ExampleVoiceLeader() {
	cMajor := scale.New(60, scale.Major)
	var chords []chord.Chord
	for _, degree := range []int{0, 3, 4, 0} { // I IV V I
		chords = append(chords, chord.Diatonic(cMajor, degree, 3))
	}
	progression := sequence.New[[]float64](chords[0], chords[1], chords[2], chords[3])
	smooth := transform.NewWithTransformer(progression, chord.NewVoiceLeader(55, 79))
	for range 4 {
		fmt.Println(smooth.Generate())
	}
	// Output:
	// [60 64 67]
	// [60 65 69]
	// [59 62 67]
	// [60 64 67]
}
