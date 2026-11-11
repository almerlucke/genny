package row_test

import (
	"fmt"

	"github.com/almerlucke/genny/row"
)

// Arrangements of three notes, in a fixed order
func ExamplePermutations() {
	for _, p := range row.Permutations([]int{0, 4, 7}, 4, 0) {
		fmt.Println(p)
	}
	// Output:
	// [0 4 7]
	// [0 7 4]
	// [4 0 7]
	// [4 7 0]
}

// Every interval from 1 to 11 occurs once in an all-interval row
func ExampleAllIntervals() {
	rows := row.AllIntervals(12, 3)
	for _, r := range rows {
		fmt.Println(r)
	}
	// Output:
	// [0 1 3 2 7 10 8 4 11 5 9 6]
	// [0 1 3 2 9 5 10 4 7 11 8 6]
	// [0 1 3 2 9 7 4 8 11 5 10 6]
}

// A reader hands out the rows one by one
func ExampleReader() {
	r := row.NewReader(row.Permutations([]int{1, 2, 3}, 0, 2), row.ReadInOrder)
	for !r.Done() {
		fmt.Print(r.Generate(), " ")
	}
	fmt.Println()
	// Output:
	// [1 2] [1 3] [2 1] [2 3] [3 1] [3 2]
}
