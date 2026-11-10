package noise_test

import (
	"fmt"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/float/noise"
	"github.com/almerlucke/genny/float/shape"
	"github.com/almerlucke/genny/scale"
	"github.com/almerlucke/genny/transform"
)

// Smooth random modulation: no jumps like white noise, and no drift away like a random walk.
// Here it plays a slowly wandering melody in D dorian: 62 + 12 * noise, quantized to the scale
func Example() {
	genny.Seed(10)
	wander := transform.New(noise.New(0.2, 2), func(v float64) float64 { return 62 + v*12 })
	melody := shape.NewSingle(wander, scale.New(62, scale.Dorian))
	for range 12 {
		fmt.Print(melody.Generate(), " ")
	}
	fmt.Println()
	// Output: 64 60 62 64 64 62 60 57 57 60 62 64
}
