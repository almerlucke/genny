package samplehold_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/samplehold"
	"github.com/almerlucke/genny/sequence"
)

// Sample every third call
func ExampleSampleHold() {
	trig := sequence.NewLoop(true, false, false)
	s := samplehold.New(trig)
	for i := 1; i <= 7; i++ {
		fmt.Print(s.Shape(float64(i)), " ")
	}
	fmt.Println()
	// Output:
	// 1 1 1 4 4 4 7
}

// Track for two calls, hold for two
func ExampleTrackHold() {
	trig := sequence.NewLoop(true, true, false, false)
	s := samplehold.NewTrack(trig)
	for i := 1; i <= 8; i++ {
		fmt.Print(s.Shape(float64(i)), " ")
	}
	fmt.Println()
	// Output:
	// 1 2 2 2 5 6 6 6
}
