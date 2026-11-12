package scale_test

import (
	"slices"
	"testing"

	"github.com/almerlucke/genny/scale"
)

func TestMessiaenModes(t *testing.T) {
	for _, c := range []struct {
		name  string
		mode  []float64
		notes int
		trans int
	}{
		{"mode 1", scale.Mode1, 6, 2},
		{"mode 2", scale.Mode2, 8, 3},
		{"mode 3", scale.Mode3, 9, 4},
		{"mode 4", scale.Mode4, 8, 6},
		{"mode 5", scale.Mode5, 6, 6},
		{"mode 6", scale.Mode6, 8, 6},
		{"mode 7", scale.Mode7, 10, 6},
	} {
		if len(c.mode) != c.notes {
			t.Errorf("%s has %d notes, want %d", c.name, len(c.mode), c.notes)
		}
		if got := scale.Transpositions(c.mode); got != c.trans {
			t.Errorf("%s: %d transpositions, want %d", c.name, got, c.trans)
		}
		scale.New(60, c.mode) // valid scale steps: does not panic
	}
}

func TestTranspositions(t *testing.T) {
	for _, c := range []struct {
		name  string
		steps []float64
		want  int
	}{
		{"major", scale.Major, 12},
		{"chromatic", scale.Chromatic, 1},
		{"diminished", scale.Diminished, 3},
		{"whole tone", scale.WholeTone, 2},
		{"empty", nil, 1},
		{"single", []float64{0}, 12},
		{"augmented triad", []float64{0, 4, 8}, 4},
		{"shifted whole tone", []float64{1, 3, 5, 7, 9, 11}, 2},
		{"beyond the octave", []float64{0, 14, 28, 18, 20, 22}, 2},
	} {
		if got := scale.Transpositions(c.steps); got != c.want {
			t.Errorf("%s: %d, want %d", c.name, got, c.want)
		}
	}
}

// the modes are themselves transposed versions of the same sets: mode 2 starting on its second note is
// one of its three transpositions
func TestModeTranspositionIsSameSet(t *testing.T) {
	m := scale.Mode2
	shifted := make([]float64, 0, len(m))
	for _, s := range m {
		shifted = append(shifted, float64((int(s)+3)%12))
	}
	slices.Sort(shifted)
	if !slices.Equal(shifted, m) {
		t.Errorf("mode 2 +3 = %v", shifted)
	}
}
