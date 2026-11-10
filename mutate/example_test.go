package mutate_test

import (
	"fmt"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/mutate"
)

// A four note phrase that is played again and again, but changes a little every time
func ExampleSequence() {
	genny.Seed(2)
	phrase := mutate.New([]int{60, 62, 64, 67}, mutate.Config[int]{
		SwapProb:   0.7,
		RotateProb: 0.3,
	})
	for range 6 {
		var bar []int
		for range 4 {
			bar = append(bar, phrase.Generate())
		}
		fmt.Println(bar)
	}
	// Output:
	// [60 62 64 67]
	// [60 62 64 67]
	// [60 62 64 67]
	// [67 64 62 60]
	// [62 64 67 60]
	// [67 64 60 62]
}

// New notes can come from your own function, for instance a random note of a scale
func ExampleConfig() {
	genny.Seed(2)
	pool := []int{60, 62, 64, 67, 69}
	phrase := mutate.New([]int{60, 60, 60, 60}, mutate.Config[int]{
		ReplaceProb: 1,
		Replace:     func() int { return pool[genny.Intn(len(pool))] },
	})
	for range 5 {
		for range 4 {
			phrase.Generate()
		}
		fmt.Println(phrase.Values())
	}
	// Output:
	// [60 62 60 60]
	// [60 62 60 60]
	// [60 69 60 60]
	// [60 69 60 67]
	// [60 69 60 69]
}
