package scale_test

import (
	"fmt"
	"strings"

	"github.com/almerlucke/genny/scale"
)

// Messiaen's modes of limited transposition: how many versions each one has
func ExampleTranspositions() {
	for i, mode := range [][]float64{scale.Mode1, scale.Mode2, scale.Mode3, scale.Mode4, scale.Major} {
		fmt.Printf("mode %d: %d notes, %d transpositions\n", i+1, len(mode), scale.Transpositions(mode))
	}
	// Output:
	// mode 1: 6 notes, 2 transpositions
	// mode 2: 8 notes, 3 transpositions
	// mode 3: 9 notes, 4 transpositions
	// mode 4: 8 notes, 6 transpositions
	// mode 5: 7 notes, 12 transpositions
}

// A Scala file gives a tuning: here the Bohlen-Pierce scale, which repeats every tritave (3/1)
func ExampleParseScl() {
	tuning, err := scale.ParseScl(strings.NewReader(scale.SclBohlenPierce))
	if err != nil {
		fmt.Println(err)
		return
	}
	s := tuning.Scale(60)
	fmt.Println(tuning.Description)
	fmt.Printf("%d degrees, period %.3f semitones\n", s.Len(), tuning.Period()/100)
	fmt.Printf("%.2f %.2f %.2f\n", s.Degree(0), s.Degree(1), s.Degree(13))
	fmt.Printf("%.2f\n", s.Quantize(63.3))
	// Output:
	// Bohlen-Pierce scale, just version
	// 13 degrees, period 19.020 semitones
	// 60.00 61.33 79.02
	// 63.02
}

// Quarter-comma meantone has a pure major third, 5/4
func ExampleScl_Scale() {
	tuning, _ := scale.ParseScl(strings.NewReader(scale.SclMeantone))
	s := tuning.Scale(0)
	fmt.Printf("third %.3f fifth %.3f (equal tempered: 4 and 7)\n", s.Degree(4), s.Degree(7))
	// Output:
	// third 3.863 fifth 6.966 (equal tempered: 4 and 7)
}
