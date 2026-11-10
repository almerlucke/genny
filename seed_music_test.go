package genny_test

import (
	"slices"
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/automaton"
	"github.com/almerlucke/genny/choose"
	"github.com/almerlucke/genny/constant"
	"github.com/almerlucke/genny/float/dist"
	"github.com/almerlucke/genny/float/lfo"
	"github.com/almerlucke/genny/float/noise"
	"github.com/almerlucke/genny/groove"
	"github.com/almerlucke/genny/mutate"
	"github.com/almerlucke/genny/ngram"
	"github.com/almerlucke/genny/progression"
	"github.com/almerlucke/genny/ratchet"
	"github.com/almerlucke/genny/rest"
	"github.com/almerlucke/genny/row"
	"github.com/almerlucke/genny/scale"
	"github.com/almerlucke/genny/sequence"
)

// every generator that uses randomness: a function that builds it (after the seed, as noise and
// automaton draw at creation) and takes values from it
var randomGenerators = map[string]func() []float64{
	"rest.Probability": func() (out []float64) {
		g := rest.NewProbability[float64](sequence.NewLoop(1.0, 2, 3), 0.5)
		for range 40 {
			o := g.Generate()
			out = append(out, o.Value, b2f(o.Rest))
		}
		return
	},
	"groove.Swing humanize": func() []float64 { return take(groove.NewSwing(0.2).Humanize(0.05), 40) },
	"groove.Humanize":       func() []float64 { return take(groove.NewHumanize(0.1), 40) },
	"groove.Jitter":         func() []float64 { return take(groove.Jitter(constant.New(0.5), 0.1), 40) },
	"scale.Walk": func() []float64 {
		return take(scale.NewWalk(scale.New(60, scale.Major), 7, 0, 14, 1, 0.2, 4), 40)
	},
	"progression": func() (out []float64) {
		g := progression.New(progression.Jazz, 0)
		for range 40 {
			out = append(out, float64(g.Generate()))
		}
		return
	},
	"automaton.NewRandom": func() (out []float64) {
		for _, c := range automaton.NewRandom(30, 40).Generate() {
			out = append(out, b2f(c))
		}
		return
	},
	"mutate": func() []float64 {
		return take(mutate.New([]float64{1, 2, 3, 4, 5, 6}, mutate.Config[float64]{
			SwapProb: 1, RotateProb: 1, ReplaceProb: 1,
		}), 60)
	},
	"choose": func() []float64 {
		return take(choose.New(
			choose.Option[float64]{Gen: constant.New(1.0), Weight: 1},
			choose.Option[float64]{Gen: constant.New(2.0), Weight: 1},
		), 40)
	},
	"lfo.SampleHold":   func() []float64 { return take(lfo.New(lfo.SampleHold, 1, 4), 40) },
	"lfo.SmoothRandom": func() []float64 { return take(lfo.New(lfo.SmoothRandom, 1, 4), 40) },
	"noise":            func() []float64 { return take(noise.New(0.1, 3), 40) },
	"dist.Gaussian":    func() []float64 { return take(dist.NewGaussian(0, 1), 40) },
	"dist.Exponential": func() []float64 { return take(dist.NewExponential(1), 40) },
	"ratchet.Random": func() (out []float64) {
		g := ratchet.NewRandom(0.5, 4)
		for range 40 {
			out = append(out, g.Generate()...)
		}
		return
	},
	"row.Random": func() []float64 { return row.Random(12) },
	"ngram": func() (out []float64) {
		g := ngram.Learn([]float64{1, 2, 1, 3, 1, 2, 4, 1, 3}, 1)
		for range 40 {
			out = append(out, g.Generate())
		}
		return
	},
}

func b2f(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

func take(g genny.Generator[float64], n int) (out []float64) {
	for range n {
		out = append(out, g.Generate())
	}
	return
}

func TestMusicGeneratorsAreSeeded(t *testing.T) {
	for name, f := range randomGenerators {
		genny.Seed(21)
		a := f()
		genny.Seed(21)
		b := f()
		genny.Seed(22)
		c := f()
		if !slices.Equal(a, b) {
			t.Errorf("%s: the same seed gave other values", name)
		}
		if slices.Equal(a, c) {
			t.Errorf("%s: another seed gave the same values", name)
		}
	}
}
