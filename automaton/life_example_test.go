package automaton_test

import (
	"fmt"
	"strings"

	"github.com/almerlucke/genny/automaton"
	"github.com/almerlucke/genny/scale"
)

// A glider walks over the grid: every 4 generations it has moved one cell down and to the right
func ExampleLife() {
	l := automaton.NewLife(6, 6, automaton.Conway)
	l.Place(automaton.Glider, 0, 0)
	fmt.Println(l)
	for range 4 {
		l.Step()
	}
	fmt.Println()
	fmt.Println(l)
	// Output:
	// .#....
	// ..#...
	// ###...
	// ......
	// ......
	// ......
	//
	// ......
	// ..#...
	// ...#..
	// .###..
	// ......
	// ......
}

// A row of the grid as a rhythm: the blinker flips the pattern every cycle
func ExampleLifeGate() {
	l := automaton.NewLife(8, 8, automaton.Conway)
	l.Place(automaton.Blinker, 2, 3)
	g := automaton.NewRowGate(l, 3)
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
	// ..xxx...
	// ...x....
	// ..xxx...
	// ...x....
}

// The live cells of a row play scale degrees, the dead ones rest
func ExampleLifePitches() {
	key := scale.New(60, scale.Minor)
	l := automaton.NewLife(8, 4, automaton.Conway)
	l.Place(automaton.Glider, 0, 0)
	p := automaton.NewLifePitches(l, 2, key)
	for range 3 {
		var cells []string
		for range 8 {
			if o := p.Generate(); o.Played() {
				cells = append(cells, fmt.Sprintf("%.0f", o.Value))
			} else {
				cells = append(cells, "-")
			}
		}
		fmt.Println(strings.Join(cells, " "))
	}
	// Output:
	// 60 62 63 - - - - -
	// - 62 63 - - - - -
	// 60 - 63 65 - - - -
}

// The population of a random grid over some generations, as a density signal
func ExampleLifeDensity() {
	l := automaton.NewLife(5, 5, automaton.Conway)
	l.Place(automaton.RPentomino, 1, 1)
	d := automaton.NewLifeDensity(l)
	var out []string
	for range 5 {
		out = append(out, fmt.Sprintf("%.2f", d.Generate()))
	}
	fmt.Println(strings.Join(out, " "))
	// Output:
	// 0.20 0.24 0.28 0.44 0.44
}
