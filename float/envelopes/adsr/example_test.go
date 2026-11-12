package adsr_test

import (
	"fmt"
	"strings"

	"github.com/almerlucke/genny/float/envelopes/adsr"
)

// An envelope at a sample rate of 1 kHz: 4 ms attack to 1, 4 ms decay to 0.5,
// a held sustain and a 4 ms release. Durations are in milliseconds.
func ExampleNew() {
	s := adsr.NewSetting(1, 4, 0.5, 4, 0, 4)
	e := adsr.New(s, adsr.NoteOff, 1000)
	e.Trigger(1)
	var out []string
	for i := 0; i < 12; i++ {
		out = append(out, fmt.Sprintf("%.2f", e.Generate()))
	}
	fmt.Println(strings.Join(out, " "))
	e.Release()
	out = out[:0]
	for !e.Done() {
		out = append(out, fmt.Sprintf("%.2f", e.Generate()))
	}
	fmt.Println(strings.Join(out, " "))
	// Output:
	// 0.00 0.25 0.50 0.75 1.00 0.88 0.75 0.62 0.50 0.50 0.50 0.50
	// 0.50 0.38 0.25 0.12
}

// An attack-release envelope that ends by itself after the sustain duration.
func ExampleNewARSetting() {
	e := adsr.New(adsr.NewARSetting(1, 2, 2, 2), adsr.Automatic, 1000)
	e.Trigger(1)
	n := 0
	for !e.Done() && n < 100 {
		fmt.Printf("%.2f ", e.Generate())
		n++
	}
	fmt.Println()
	// Output:
	// 0.00 0.50 1.00 1.00 1.00 0.50
}
