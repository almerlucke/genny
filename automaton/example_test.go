package automaton_test

import (
	"fmt"
	"strings"

	"github.com/almerlucke/genny/automaton"
)

func show(row []bool) string {
	var b strings.Builder
	for _, v := range row {
		if v {
			b.WriteByte('#')
		} else {
			b.WriteByte('.')
		}
	}
	return b.String()
}

// Rule 30 grows chaotic patterns from a single cell
func ExampleAutomaton() {
	a := automaton.New(30, 15)
	for range 6 {
		fmt.Println(show(a.Generate()))
	}
	// Output:
	// .......#.......
	// ......###......
	// .....##..#.....
	// ....##.####....
	// ...##..#...#...
	// ..##.####.###..
}

// Gate plays the cells one by one: a rhythm that slowly changes
func ExampleGate() {
	g := automaton.NewGate(automaton.New(30, 8))
	for range 4 {
		for range 8 {
			if g.Generate() {
				fmt.Print("x")
			} else {
				fmt.Print(".")
			}
		}
		fmt.Println()
	}
	// Output:
	// ....x...
	// ...xxx..
	// ..xx..x.
	// .xx.xxxx
}
