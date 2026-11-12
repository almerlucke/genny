package polyrhythm_test

import (
	"fmt"

	"github.com/almerlucke/genny/polyrhythm"
)

// Three against two in a bar of two beats on the least common multiple grid, and when they realign
func ExampleNewGrid() {
	three := polyrhythm.NewLayer(3, polyrhythm.Int(2))
	two := polyrhythm.NewLayer(2, polyrhythm.Int(2))
	g := polyrhythm.NewGrid(three, two)
	fmt.Println(g)
	fmt.Println("unit", g.Unit, "beat; cycle", polyrhythm.Cycle(three, two), "beats")
	// Output:
	// x.x.x.
	// x..x..
	// unit 1/3 beat; cycle 2 beats
}

// 7 pulses in 4 beats against a cross rhythm of 7/2 beats realign only after 28 beats
func ExampleCycle() {
	seven := polyrhythm.NewLayer(7, polyrhythm.Int(4))
	cross := polyrhythm.Every(polyrhythm.F(7, 2))
	fmt.Println(polyrhythm.Cycle(seven, cross), polyrhythm.Realign(seven, cross))
	// Output: 28 [0]
}

// Quarter note triplet = quarter note: 90 bpm becomes 135. Dotted quarter = quarter slows down
func ExampleEquate() {
	up := polyrhythm.Equate(polyrhythm.F(2, 3), polyrhythm.Int(1))
	down := polyrhythm.Equate(polyrhythm.F(3, 2), polyrhythm.Int(1))
	fmt.Println(up.Tempo(polyrhythm.Int(90)), down.Tempo(polyrhythm.Int(120)), polyrhythm.Ratio(3, 2).Tempo(polyrhythm.Int(120)))
	// Output: 135 80 180
}

func ExampleNewPlan() {
	plan := polyrhythm.NewPlan(polyrhythm.Int(120),
		polyrhythm.Section{Name: "A", Bars: 2},
		polyrhythm.Section{Name: "B", Bars: 2, Into: polyrhythm.Ratio(3, 2)},
	)
	for _, s := range plan {
		fmt.Printf("%s bar %d at %s bpm, %s s\n", s.Name, s.FirstBar, s.Tempo, s.Time)
	}
	fmt.Println(plan.TempoChanges(4))
	// Output:
	// A bar 0 at 120 bpm, 0 s
	// B bar 2 at 180 bpm, 4 s
	// [{32 180}]
}
