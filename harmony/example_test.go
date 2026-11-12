package harmony_test

import (
	"fmt"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/harmony"
)

// A phrase of four chords in G major that ends on a half cadence, then one that ends on an
// authentic cadence
func ExampleGenerator() {
	genny.Seed(3)
	g := harmony.New(harmony.Config{Key: harmony.Key{Tonic: 7}, PhraseLength: 4, SeventhProb: 0.5})
	for range 8 {
		c := g.Generate()
		fmt.Print(c.Numeral, " ")
	}
	fmt.Println()
	// Output: I ii7 IV V I IV V7 I
}

// Modulate gives a pivot chord and the dominant seventh of the new key
func ExampleModulate() {
	for _, c := range harmony.Modulate(harmony.Key{Tonic: 7}, harmony.Key{Tonic: 2}) {
		fmt.Println(c.Numeral, c.Name(), c.Key)
	}
	// Output:
	// vi=ii Em D major
	// V7 A7 D major
}

// Check lists what is wrong with a voicing: the classic parallel fifths of I - V
func ExampleCheck() {
	key := harmony.Key{}
	I, V := harmony.Diatonic(key, 0, false), harmony.Diatonic(key, 4, false)
	bad := []harmony.Voicing{
		{Chord: I, Notes: [4]int{72, 64, 55, 48}},
		{Chord: V, Notes: [4]int{71, 67, 62, 55}}, // tenor and bass both a fifth up
	}
	for _, v := range harmony.Check(bad) {
		fmt.Println(v)
	}
	// Output:
	// chord 2: parallel fifths [tenor bass]: 55,48 to 62,55
}
