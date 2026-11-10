package lsystem_test

import (
	"fmt"

	"github.com/almerlucke/genny/lsystem"
)

// Lindenmayer's algae: A becomes AB and B becomes A, the lengths follow the Fibonacci numbers
func ExampleExpand() {
	rules := map[rune]string{'A': "AB", 'B': "A"}
	for n := range 6 {
		fmt.Println(lsystem.Expand("A", rules, n))
	}
	// Output:
	// A
	// AB
	// ABA
	// ABAAB
	// ABAABABA
	// ABAABABAABAAB
}

// Play an L-system as a melody: every symbol is a note, symbols without a mapping are skipped.
// The brackets are structure only. The result is self-similar: every generation starts with the previous one
func ExampleNewLoop() {
	rules := map[rune]string{'A': "A[B]A", 'B': "C", 'C': "B"}
	notes := map[rune]int{'A': 60, 'B': 64, 'C': 67}

	melody := lsystem.NewLoop("A", rules, 3, notes)
	for range 12 {
		fmt.Print(melody.Generate(), " ")
	}
	fmt.Println()
	// Output: 60 64 60 67 60 64 60 64 60 64 60 67
}
