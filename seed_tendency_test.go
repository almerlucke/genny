package genny_test

import (
	"slices"
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/choose"
	"github.com/almerlucke/genny/float/dist"
	"github.com/almerlucke/genny/float/tendency"
)

func takeAny[T any](g genny.Generator[T], n int) (out []T) {
	for range n {
		out = append(out, g.Generate())
	}
	return
}

func toFloats(ints []int) (out []float64) {
	for _, i := range ints {
		out = append(out, float64(i))
	}
	return
}

var items10 = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

// the generators of the tendency, dist and choose additions that draw from the genny stream
var randomGenerators2 = map[string]func() []float64{
	"tendency.Value": func() []float64 {
		return takeAny[float64](tendency.NewValue(tendency.NewMask(true, tendency.Segment{N: 5, A1: 0, A2: 1, Z1: 5, Z2: 9})), 40)
	},
	"tendency.BetaValue": func() []float64 {
		return takeAny[float64](tendency.NewBetaValue(tendency.NewMask(true, tendency.Segment{N: 1, A1: 0, A2: 9}), 2, 3), 40)
	},
	"tendency.Choice": func() []float64 {
		m := tendency.NewMask(true, tendency.Segment{N: 4, A1: 0, A2: 100, Z1: 0, Z2: 50})
		return toFloats(takeAny[int](tendency.NewChoice(m, items10), 40))
	},
	"tendency.BetaChoice": func() []float64 {
		m := tendency.NewMask(true, tendency.Segment{N: 1, A1: 0, A2: 100})
		return toFloats(takeAny[int](tendency.NewBetaChoice(m, items10, 0.5, 0.5), 40))
	},
	"tendency.TwoMasks": func() []float64 {
		outer := tendency.NewMask(true, tendency.Segment{N: 1, A1: 0, A2: 10})
		inner := tendency.NewMask(true, tendency.Segment{N: 1, A1: 4, A2: 6})
		return takeAny[float64](tendency.NewTwoMasks(outer, inner, 10), 40)
	},
	"tendency.Spray": func() []float64 { return tendency.Spray([]float64{1, 2, 3, 4, 5, 6, 7, 8}, 2) },
	"dist.Beta":      func() []float64 { return takeAny[float64](dist.NewBeta(2, 3), 40) },
	"dist.Cauchy":    func() []float64 { return takeAny[float64](dist.NewCauchy(0, 1).Retry(-5, 5, 3), 40) },
	"dist.Weibull":   func() []float64 { return takeAny[float64](dist.NewWeibull(0, 1, 1.5), 40) },
	"dist.Pink":      func() []float64 { return takeAny[float64](dist.NewPink(6), 40) },
	"choose.Index": func() []float64 {
		var out []float64
		for _, d := range []choose.Distribution{
			choose.Uniform(), choose.Linear(), choose.Exponential(2), choose.Gaussian(4, 2), choose.Beta(2, 2),
			choose.Cauchy(0.5, 0.2), choose.Weibull(0.5, 1.5), choose.Triangle(0.3), choose.Ratio(3, 2, 1),
		} {
			out = append(out, toFloats(takeAny[int](choose.NewIndex(items10, d), 20))...)
		}
		return out
	},
	"choose.ChangingWeights": func() []float64 {
		g := choose.NewChangingWeights(items10, 5, []float64{1, 1, 1, 1, 1, 1}, []float64{0, 0, 0, 0, 1, 1, 1, 1, 1, 1})
		return toFloats(takeAny[int](g, 40))
	},
}

func TestTendencyGeneratorsAreSeeded(t *testing.T) {
	for name, f := range randomGenerators2 {
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

// the Lehmer generator is repeatable by its own seed, and ignores genny.Seed
func TestLehmerHasItsOwnSeed(t *testing.T) {
	genny.Seed(1)
	a := takeAny[float64](dist.NewLehmer(5), 20)
	genny.Seed(2)
	b := takeAny[float64](dist.NewLehmer(5), 20)
	c := takeAny[float64](dist.NewLehmer(6), 20)
	if !slices.Equal(a, b) || slices.Equal(a, c) {
		t.Error("Lehmer should depend on its own seed only")
	}
}
