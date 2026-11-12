package melody_test

import (
	"fmt"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/chord"
	"github.com/almerlucke/genny/melody"
	"github.com/almerlucke/genny/scale"
)

// A melody over I and V in C major, 4 steps to a bar: chord tones on the beat, a scale tone between
func ExampleChordTonesList() {
	genny.Seed(1)
	key := scale.New(60, scale.Major)
	chords := []chord.Chord{chord.Diatonic(key, 0, 3), chord.Diatonic(key, 4, 3)}
	events := melody.ChordTonesList(chords, 2, melody.Config{Scale: key, BarSteps: 4, BeatSteps: 2})
	for _, e := range events {
		fmt.Printf("%g@%g ", e.Pitch, e.Start)
	}
	fmt.Println()
	// Output:
	// 72@0 71@1 67@2 72@3 71@4 72@5 71@6 71@7
}
