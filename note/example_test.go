package note_test

import (
	"fmt"

	"github.com/almerlucke/genny/euclid"
	"github.com/almerlucke/genny/groove"
	"github.com/almerlucke/genny/note"
	"github.com/almerlucke/genny/rest"
	"github.com/almerlucke/genny/scale"
	"github.com/almerlucke/genny/sequence"
)

// A voice combines the generators: pitches with rests from a Euclidean rhythm, accents for the velocity,
// and swing for the timing
func ExampleVoice() {
	minor := scale.New(57, scale.Minor)
	degrees := sequence.NewLoop(0, 2, 4, 7)
	pitch := rest.NewGated[float64](scale.NewDegrees(minor, degrees), euclid.New(8, 5, 0))

	voice := note.Voice{
		Pitch:    pitch,
		Velocity: groove.NewAccent(2, 2, 2, 2),
		Offset:   groove.NewSwing(0.25),
	}
	for _, e := range voice.Render(8) {
		fmt.Printf("step %.2f: note %.0f velocity %.2f\n", e.Start, e.Pitch, e.Velocity)
	}
	// Output:
	// step 0.00: note 57 velocity 1.00
	// step 2.00: note 64 velocity 0.80
	// step 3.25: note 69 velocity 0.55
	// step 5.25: note 60 velocity 0.55
	// step 6.00: note 64 velocity 0.80
}
