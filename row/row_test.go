package row

import (
	"slices"
	"testing"

	"github.com/almerlucke/genny"
)

func TestOperations(t *testing.T) {
	r := []float64{0, 2, 7, 3}
	for name, c := range map[string]struct{ got, want []float64 }{
		"retrograde":  {Retrograde(r), []float64{3, 7, 2, 0}},
		"inversion":   {Inversion(r), []float64{0, -2, -7, -3}},
		"retro-inv":   {RetrogradeInversion(r), []float64{-3, -7, -2, 0}},
		"transpose":   {Transpose(r, 5), []float64{5, 7, 12, 8}},
		"apply prime": {Apply(r, Form{Prime, 0}), []float64{0, 2, 7, 3}},
		"apply I+2":   {Apply(r, Form{Invert, 2}), []float64{2, 0, -5, -1}},
		"apply R+1":   {Apply(r, Form{Retro, 1}), []float64{4, 8, 3, 1}},
	} {
		if !slices.Equal(c.got, c.want) {
			t.Errorf("%s: %v, want %v", name, c.got, c.want)
		}
	}
	if !slices.Equal(r, []float64{0, 2, 7, 3}) {
		t.Error("the row was changed")
	}
}

func TestRandomIsAPermutation(t *testing.T) {
	genny.Seed(1)
	r := Random(12)
	s := slices.Clone(r)
	slices.Sort(s)
	for i, v := range s {
		if v != float64(i) {
			t.Fatalf("%v", r)
		}
	}
}

func TestSerial(t *testing.T) {
	s := NewSerial([]float64{0, 4, 7}, Form{Prime, 0}, Form{Retro, 12})
	var got []float64
	for range 8 {
		got = append(got, s.Generate())
	}
	if !slices.Equal(got, []float64{0, 4, 7, 19, 16, 12, 0, 4}) {
		t.Errorf("%v", got)
	}
}

// the row of Schoenberg's Op. 25
func TestMatrix(t *testing.T) {
	r := []int{4, 5, 7, 1, 6, 3, 8, 2, 11, 0, 9, 10}
	m := Matrix(r)
	if !slices.Equal(m[0], r) {
		t.Errorf("first row %v", m[0])
	}
	for i := range r {
		if want := ((2*r[0]-r[i])%12 + 12) % 12; m[i][0] != want {
			t.Errorf("first column %d: %d, want %d", i, m[i][0], want)
		}
		row := slices.Clone(m[i])
		col := make([]int, len(r))
		for j := range r {
			col[j] = m[j][i]
		}
		slices.Sort(row)
		slices.Sort(col)
		for j := range 12 {
			if row[j] != j || col[j] != j {
				t.Fatalf("row or column %d is not a permutation", i)
			}
		}
	}
	// the second row is the row transposed to start on 3 (the inversion of 5 around 4)
	if !slices.Equal(m[1], []int{3, 4, 6, 0, 5, 2, 7, 1, 10, 11, 8, 9}) {
		t.Errorf("second row %v", m[1])
	}
}
