package row_test

import (
	"fmt"

	"github.com/almerlucke/genny/row"
)

func ExampleInversion() {
	r := []float64{0, 2, 7, 3}
	fmt.Println("prime        ", r)
	fmt.Println("retrograde   ", row.Retrograde(r))
	fmt.Println("inversion    ", row.Inversion(r))
	fmt.Println("retro-invert ", row.RetrogradeInversion(r))
	fmt.Println("up a fifth   ", row.Transpose(r, 7))
	// Output:
	// prime         [0 2 7 3]
	// retrograde    [3 7 2 0]
	// inversion     [0 -2 -7 -3]
	// retro-invert  [-3 -7 -2 0]
	// up a fifth    [7 9 14 10]
}

// A melody of a row: the prime form, then its retrograde a fifth higher, and around again
func ExampleNewSerial() {
	melody := row.NewSerial([]float64{60, 62, 67, 63},
		row.Form{Op: row.Prime},
		row.Form{Op: row.Retro, Transpose: 7},
	)
	for range 10 {
		fmt.Print(melody.Generate(), " ")
	}
	fmt.Println()
	// Output: 60 62 67 63 70 74 69 67 60 62
}

// The twelve tone matrix of a row of pitch classes: every row and column holds all twelve notes
func ExampleMatrix() {
	for _, line := range row.Matrix([]int{4, 5, 7, 1, 6, 3, 8, 2, 11, 0, 9, 10}) {
		fmt.Printf("%2d\n", line)
	}
	// Output:
	// [ 4  5  7  1  6  3  8  2 11  0  9 10]
	// [ 3  4  6  0  5  2  7  1 10 11  8  9]
	// [ 1  2  4 10  3  0  5 11  8  9  6  7]
	// [ 7  8 10  4  9  6 11  5  2  3  0  1]
	// [ 2  3  5 11  4  1  6  0  9 10  7  8]
	// [ 5  6  8  2  7  4  9  3  0  1 10 11]
	// [ 0  1  3  9  2 11  4 10  7  8  5  6]
	// [ 6  7  9  3  8  5 10  4  1  2 11  0]
	// [ 9 10  0  6 11  8  1  7  4  5  2  3]
	// [ 8  9 11  5 10  7  0  6  3  4  1  2]
	// [11  0  2  8  1 10  3  9  6  7  4  5]
	// [10 11  1  7  0  9  2  8  5  6  3  4]
}
