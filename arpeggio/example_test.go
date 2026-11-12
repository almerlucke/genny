package arpeggio_test

import (
	"fmt"

	"github.com/almerlucke/genny/arpeggio"
)

func ExampleNew() {
	chord := []int{60, 64, 67, 72}
	for _, mode := range []arpeggio.Mode{arpeggio.Up, arpeggio.Converge, arpeggio.Alternate} {
		a := arpeggio.New(chord, mode, arpeggio.None, false)
		var out []int
		for !a.Done() {
			out = append(out, a.Generate())
		}
		fmt.Println(out)
	}
	// Output:
	// [60 64 67 72]
	// [60 72 64 67]
	// [60 64 60 67 60 72]
}

// Mirroring turns an upward run into an up-and-down figure.
func ExampleNew_mirror() {
	a := arpeggio.New([]int{60, 64, 67, 72}, arpeggio.Up, arpeggio.Exclusive, false)
	for !a.Done() {
		fmt.Print(a.Generate(), " ")
	}
	fmt.Println()
	// Output: 60 64 67 72 67 64
}

func ExampleNewLoop() {
	a := arpeggio.NewLoop([]int{60, 64, 67}, arpeggio.Up, arpeggio.None, true)
	for i := 0; i < 6; i++ {
		fmt.Print(a.Generate(), " ")
	}
	fmt.Println()
	// Output: 67 64 60 67 64 60
}
