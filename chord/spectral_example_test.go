package chord_test

import (
	"fmt"

	"github.com/almerlucke/genny/chord"
)

// Partials 1 to 6 of A1 (55 Hz): microtonal, the 5th, 7th partials are off the piano
func ExampleHarmonic() {
	c := chord.Harmonic(55, []int{1, 2, 3, 4, 5, 6}, 1, false)
	fmt.Printf("%.2f\n", c)
	// stretched partials, as in a bell or a piano string
	stretched := chord.Harmonic(55, []int{1, 2, 3, 4}, 1.05, false)
	fmt.Printf("%.2f\n", stretched)
	// subharmonics go down from the fundamental
	fmt.Printf("%.2f\n", chord.Harmonic(440, []int{1, 2, 3, 4}, 1, true))
	// Output:
	// [33.00 45.00 52.02 57.00 60.86 64.02]
	// [33.00 45.60 52.97 58.20]
	// [45.00 49.98 57.00 69.00]
}

// Chord multiplication: a major triad built on every note of a fifth
func ExampleMultiply() {
	fmt.Println(chord.Multiply(chord.Chord{60, 67}, chord.Chord{0, 4, 7}))
	// Output:
	// [60 64 67 71 74]
}

// Frequency modulation and ring modulation of two tones
func ExampleFM() {
	fmt.Printf("%.2f\n", chord.FM(440, 110, 2))
	fmt.Printf("%.2f\n", chord.RingMod(440, 660))
	// Output:
	// [57.00 64.02 69.00 72.86 76.02]
	// [57.00 84.86]
}
