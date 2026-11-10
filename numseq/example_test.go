package numseq_test

import (
	"fmt"

	"github.com/almerlucke/genny/numseq"
)

// Fibonacci mod 8 as scale degrees gives a melody that repeats after 12 notes (the Pisano period)
func ExampleFibonacci() {
	f := numseq.NewFibonacci(8)
	for range 14 {
		fmt.Print(f.Generate(), " ")
	}
	fmt.Println()
	// Output: 0 1 1 2 3 5 0 5 5 2 7 1 0 1
}

// Hailstone numbers: up and down in a jagged path that always ends at 1
func ExampleNewCollatz() {
	c := numseq.NewCollatz(7)
	for !c.Done() {
		fmt.Print(c.Generate(), " ")
	}
	fmt.Println()
	// Output: 7 22 11 34 17 52 26 13 40 20 10 5 16 8 4 2 1
}

// Thue-Morse as a gate: no long runs, never a repeating pattern
func ExampleThueMorse() {
	tm := numseq.NewThueMorse()
	for range 24 {
		if tm.Generate() == 1 {
			fmt.Print("x")
		} else {
			fmt.Print(".")
		}
	}
	fmt.Println()
	// Output: .xx.x..xx..x.xx.x..x.xx.
}
