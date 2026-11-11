package genny_test

import (
	"slices"
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/constant"
	"github.com/almerlucke/genny/listtools"
	"github.com/almerlucke/genny/sieve"
	"github.com/almerlucke/genny/split"
)

// the random generators of the rhythm packages, see randomGenerators
var randomRhythmGenerators = map[string]func() []float64{
	"sieve.Funnel": func() []float64 {
		in := make([]float64, 60)
		for i := range in {
			in[i] = float64(i) / 2 // many ties between whole numbers
		}
		return sieve.Funnel(in, []float64{0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30})
	},
	"split.Split": func() (out []float64) {
		for range 5 {
			out = append(out, split.Split(8, 0.125, 0.6, 2)...)
		}
		return
	},
	"split.Gen":         func() []float64 { return take(split.NewGen(4, 0.125, 0.6, 2), 60) },
	"listtools.PlusMin": func() []float64 { return take(listtools.NewPlusMin(constant.New(1.0), 0.5), 60) },
}

func TestRhythmGeneratorsAreSeeded(t *testing.T) {
	for name, f := range randomRhythmGenerators {
		genny.Seed(31)
		a := f()
		genny.Seed(31)
		b := f()
		genny.Seed(32)
		c := f()
		if !slices.Equal(a, b) {
			t.Errorf("%s: the same seed gave other values", name)
		}
		if slices.Equal(a, c) {
			t.Errorf("%s: another seed gave the same values", name)
		}
	}
}
