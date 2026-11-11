package pulse_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/line"
	"github.com/almerlucke/genny/pulse"
)

func show(p []bool) string {
	s := ""
	for _, v := range p {
		if v {
			s += "x"
		} else {
			s += "."
		}
	}
	return s
}

// The importance of the pulses of a bar: the first is the most important, then the half bar
func ExampleHierarchy() {
	fmt.Println(pulse.Hierarchy(2, 2))
	fmt.Println(pulse.Hierarchy(3, 2, 2))
	fmt.Println(pulse.Hierarchy(2, 2, 2, 2))
	// Output:
	// [3 0 2 1]
	// [11 0 6 5 9 2 8 3 10 1 7 4]
	// [15 0 8 7 12 3 11 4 14 1 9 6 13 2 10 5]
}

// Thinning keeps the pulses that matter most. A higher level is sparser
func ExampleThin() {
	ind := pulse.Hierarchy(2, 2, 2, 2)
	for _, level := range []int{0, 8, 12, 14, 15, 16} {
		fmt.Printf("%2d %s\n", level, show(pulse.Thin(ind, level)))
	}
	// Output:
	//  0 xxxxxxxxxxxxxxxx
	//  8 x.x.x.x.x.x.x.x.
	// 12 x...x...x...x...
	// 14 x.......x.......
	// 15 x...............
	// 16 ................
}

// The thinning level follows a generator: here a line, so the rhythm fades in over four bars
func ExampleNewGate() {
	ind := pulse.Hierarchy(2, 2, 2, 2)
	level := line.New(16)
	level.To(0, 64)
	g := pulse.NewGate(ind, level)
	for bar := 0; bar < 4; bar++ {
		var steps []bool
		for range 16 {
			steps = append(steps, g.Generate())
		}
		fmt.Println(show(steps))
	}
	// Output:
	// ........x...x...
	// x...x.x.x...x.x.
	// x.xxx.x.x.xxx.xx
	// x.xxxxxxx.xxxxxx
}

// Importance as velocity
func ExampleVelocity() {
	fmt.Printf("%.2f\n", pulse.Velocity(pulse.Hierarchy(2, 2), 0.3, 1))
	// Output:
	// [1.00 0.30 0.77 0.53]
}

// Interpolate morphs one hierarchy into another
func ExampleInterpolate() {
	a := pulse.Hierarchy(2, 2, 2)
	b := pulse.Hierarchy(2, 2, 2)
	for i := range b {
		b[i] = len(b) - 1 - b[i]
	}
	fmt.Println(pulse.Interpolate(a, b, 0))
	fmt.Println(pulse.Interpolate(a, b, 0.5))
	fmt.Println(pulse.Interpolate(a, b, 1))
	// Output:
	// [7 0 4 3 6 1 5 2]
	// [4 4 4 4 4 4 4 4]
	// [0 7 3 4 1 6 2 5]
}
