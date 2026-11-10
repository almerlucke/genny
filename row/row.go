// Package row has tone rows and the serial operations on them: retrograde, inversion, transposition and
// the twelve tone matrix. A row is a list of pitches in semitones (or pitch classes 0..11)
package row

import (
	"slices"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/sequence"
)

// Random returns a random permutation of 0..n-1 (n = 12 for a twelve tone row)
func Random(n int) []float64 {
	r := make([]float64, n)
	for i := range r {
		r[i] = float64(i)
	}
	genny.Shuffle(n, func(i, j int) { r[i], r[j] = r[j], r[i] })
	return r
}

// Retrograde plays the row backwards
func Retrograde(r []float64) []float64 {
	out := slices.Clone(r)
	slices.Reverse(out)
	return out
}

// Inversion mirrors every interval around the first note: up becomes down
func Inversion(r []float64) []float64 {
	out := make([]float64, len(r))
	for i, p := range r {
		out[i] = 2*r[0] - p
	}
	return out
}

// RetrogradeInversion is the inversion played backwards
func RetrogradeInversion(r []float64) []float64 { return Retrograde(Inversion(r)) }

// Transpose moves all notes by semitones
func Transpose(r []float64, semitones float64) []float64 {
	out := make([]float64, len(r))
	for i, p := range r {
		out[i] = p + semitones
	}
	return out
}

// Op is a serial operation
type Op int

const (
	Prime Op = iota
	Retro
	Invert
	RetroInvert
)

// Form is an operation followed by a transposition
type Form struct {
	Op        Op
	Transpose float64
}

// Apply returns the form of the row
func Apply(r []float64, f Form) []float64 {
	var out []float64
	switch f.Op {
	case Retro:
		out = Retrograde(r)
	case Invert:
		out = Inversion(r)
	case RetroInvert:
		out = RetrogradeInversion(r)
	default:
		out = slices.Clone(r)
	}
	return Transpose(out, f.Transpose)
}

// NewSerial plays the forms of a row one after the other, and loops, for instance
// NewSerial(r, Form{Prime, 0}, Form{Retro, 0}, Form{Invert, 5})
func NewSerial(r []float64, forms ...Form) *sequence.Sequence[float64] {
	var all []float64
	for _, f := range forms {
		all = append(all, Apply(r, f)...)
	}
	return sequence.NewLoop(all...)
}

// Matrix is the twelve tone matrix of a row of pitch classes (0..11): the first row is the row itself,
// the first column its inversion, and every other row is the row transposed to start on the
// first column's note
func Matrix(r []int) [][]int {
	n := len(r)
	m := make([][]int, n)
	for i := range m {
		m[i] = make([]int, n)
		start := ((2*r[0]-r[i])%12 + 12) % 12
		for j := range n {
			m[i][j] = ((r[j]-r[0]+start)%12 + 12) % 12
		}
	}
	return m
}
