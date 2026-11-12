package genny_test

import (
	"slices"
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/automaton"
	"github.com/almerlucke/genny/constant"
	"github.com/almerlucke/genny/evolve"
	"github.com/almerlucke/genny/scale"
	"github.com/almerlucke/genny/stochastic"
)

// the generators of stochastic, life and evolve that use randomness, built after the seed
var stochasticGenerators = map[string]func() []float64{
	"stochastic.Poisson": func() []float64 { return stochastic.Poisson(2, 30) },
	"stochastic.PoissonInt": func() (out []float64) {
		for range 40 {
			out = append(out, float64(stochastic.PoissonInt(4)))
		}
		return
	},
	"stochastic.Cloud": func() (out []float64) {
		c := stochastic.NewCloud(stochastic.Config{
			Density:  func(step int) float64 { return 1 + float64(step)/4 },
			Pitch:    stochastic.Uniform(48, 72),
			Velocity: stochastic.Uniform(0.3, 0.9),
			Duration: stochastic.Uniform(1, 4),
		})
		for _, e := range c.Render(12) {
			out = append(out, e.Pitch, e.Velocity, e.Start, e.Duration)
		}
		return
	},
	"stochastic.Density": func() (out []float64) {
		g := stochastic.NewDensity(constant.New(0.5))
		for range 60 {
			out = append(out, b2f(g.Generate()))
		}
		return
	},
	"stochastic.Free": func() (out []float64) {
		for _, e := range stochastic.Free(20, 2, 1.5, 60, 4, scale.New(0, scale.Major)) {
			out = append(out, e.Pitch, e.Velocity, e.Start, e.Duration)
		}
		return
	},
	"automaton.NewLifeRandom": func() (out []float64) {
		l := automaton.NewLifeRandom(10, 10, automaton.Conway, 0.4)
		g := automaton.NewRowGate(l, 3)
		for range 40 {
			out = append(out, b2f(g.Generate()))
		}
		return
	},
	"evolve": func() (out []float64) {
		ga := evolve.New(evolve.Config{
			Population: 20, Length: 8, Min: 0, Max: 14,
			Fitness:      evolve.Weighted(evolve.Weight(evolve.Smoothness(), 1), evolve.Weight(evolve.IntervalVariety(), 1)),
			MutationRate: 0.15, Elitism: 1,
		})
		p := evolve.NewPlayer(ga, 3)
		for range 40 {
			out = append(out, float64(p.Generate()))
		}
		return
	},
}

func TestStochasticGeneratorsAreSeeded(t *testing.T) {
	for name, f := range stochasticGenerators {
		genny.Seed(31)
		a := f()
		genny.Seed(31)
		b := f()
		genny.Seed(32)
		c := f()
		if len(a) == 0 {
			t.Errorf("%s: no values", name)
		}
		if !slices.Equal(a, b) {
			t.Errorf("%s: the same seed gave other values", name)
		}
		if slices.Equal(a, c) {
			t.Errorf("%s: another seed gave the same values", name)
		}
	}
}
