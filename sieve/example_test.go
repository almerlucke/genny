package sieve_test

import (
	"fmt"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/sequence"
	"github.com/almerlucke/genny/sieve"
)

// A sieve is a residue class: every third number from 1
func ExampleNew() {
	fmt.Println(sieve.New(3, 1, 0, 12).Values())
	// Output:
	// [1 4 7 10]
}

// Sieves combine into new sets. Here a union and intersection of thirds and fours
func ExampleSieve_Union() {
	thirds := sieve.New(3, 0, 0, 24)
	fours := sieve.New(4, 0, 0, 24)
	fmt.Println("union       ", thirds.Union(fours).Values())
	fmt.Println("intersection", thirds.Intersection(fours).Values())
	fmt.Println("difference  ", thirds.Difference(fours).Values())
	fmt.Println("complement  ", sieve.New(4, 0, 0, 12).Complement().Values())
	// Output:
	// union        [0 3 4 6 8 9 12 15 16 18 20 21 24]
	// intersection [0 12 24]
	// difference   [3 6 9 15 18 21]
	// complement   [1 2 3 5 6 7 9 10 11]
}

// A sieve as a pitch set: the semitones above a base note
func ExampleScale() {
	s := sieve.New(5, 0, 0, 12).Union(sieve.New(7, 0, 0, 12))
	fmt.Println(sieve.Scale(s, 60))
	// Output:
	// [60 65 67 70]
}

// The sieve as a gate: a rhythm of every third step in 8
func ExampleNewGate() {
	g := sieve.NewGate(sieve.New(3, 0, 0, 100), 8)
	for range 8 {
		if g.Generate() {
			fmt.Print("x")
		} else {
			fmt.Print(".")
		}
	}
	fmt.Println()
	// Output:
	// x..x..x.
}

// Filter keeps the notes of a melody at the sieve positions, the rest are rests
func ExampleNewFilter() {
	f := sieve.NewFilter[int](sequence.NewLoop(60, 62, 64, 65, 67, 69), sieve.New(2, 0, 0, 5), 6)
	for range 6 {
		o := f.Generate()
		if o.Played() {
			fmt.Print(o.Value, " ")
		} else {
			fmt.Print("- ")
		}
	}
	fmt.Println()
	// Output:
	// 60 - 64 - 67 -
}

// Funnel snaps values to the nearest allowed value
func ExampleFunnel() {
	genny.Seed(1)
	major := sieve.Scale(sieve.Of(0, 12, 0, 2, 4, 5, 7, 9, 11, 12), 0)
	fmt.Println(sieve.Funnel([]float64{0.2, 1.4, 3.1, 6.2, 10.8}, major))
	// Output:
	// [0 2 4 7 11]
}

// Interpret applies a sieve to a rhythm: kind 1 makes the other values rests, kind 2 lets each
// indexed value absorb the ones after it
func ExampleInterpret() {
	pattern := []float64{1, 1, 1, 1, 1, 1, 1, 1}
	s := sieve.New(3, 0, 0, 7)
	fmt.Println(sieve.Interpret(pattern, s, 1))
	fmt.Println(sieve.Interpret(pattern, s, 2))
	fmt.Println(sieve.Interpret(pattern, s, 3))
	fmt.Println(sieve.Interpret(pattern, s, 4))
	// Output:
	// [1 -1 -1 1 -1 -1 1 -1]
	// [3 3 2]
	// [-1 1 1 -1 1 1 -1 1]
	// [-1 1 2 1 2 1]
}

// Gen plays the members of a sieve
func ExampleNewGen() {
	g := sieve.NewGen(sieve.New(4, 1, 0, 12), true)
	for range 6 {
		fmt.Print(g.Generate(), " ")
	}
	fmt.Println()
	// Output:
	// 1 5 9 1 5 9
}
