package main

import (
	"fmt"
	"github.com/almerlucke/genny/float/export/plot"
	"log"
	"math"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/automaton"
	"github.com/almerlucke/genny/evolve"
	"github.com/almerlucke/genny/scale"
	"github.com/almerlucke/genny/stochastic"
)

// stochScatter plots points as dots or squares
func stochScatter(dir, name, title, xLabel, yLabel string, w, h plot.Length, series ...plot.XYs) {
	p := plot.New()
	p.Title.Text, p.X.Label.Text, p.Y.Label.Text = title, xLabel, yLabel
	for _, pts := range series {
		s, err := plot.NewScatter(pts)
		if err != nil {
			log.Fatalf("%s: %v", name, err)
		}
		s.GlyphStyle.Shape = plot.BoxGlyph{}
		s.GlyphStyle.Radius = plot.Points(2)
		p.Add(s)
	}
	consSave(p, dir, name, w, h)
}

// plotsStochastic writes the plots of the stochastic tools, Life, the evolution and the Messiaen modes to dir
func plotsStochastic(dir string) {
	// a Poisson process and a cloud follow a density curve: the number of events per step against the
	// density, and the onsets of a steady Poisson process
	genny.Seed(51)
	const steps = 96
	density := func(step int) float64 { return 0.2 + 3*math.Pow(math.Sin(math.Pi*float64(step)/steps), 2) }
	cloud := stochastic.NewCloud(stochastic.Config{Density: density, Pitch: stochastic.Uniform(48, 72)})
	counts := make([]float64, steps)
	for _, e := range cloud.Render(steps) {
		counts[int(e.Start)]++
	}
	step, step2 := 0, 0
	lines(dir, "stochastic_cloud.png", "Cloud: events per step against the density", steps,
		[]string{"events", "density"},
		func() float64 { step++; return counts[step-1] },
		func() float64 { step2++; return density(step2 - 1) })

	// onsets of Poisson processes of rates 0.5, 1 and 2 as stems on three lines
	p := plot.New()
	p.Title.Text, p.X.Label.Text, p.Y.Label.Text = "Poisson onsets (rate 0.5, 1, 2 per step)", "step", "rate"
	for i, rate := range []float64{0.5, 1, 2} {
		pts := plot.XYs{}
		for _, t := range stochastic.Poisson(rate, 64) {
			pts = append(pts, plot.XY{X: t, Y: float64(i)})
		}
		s, err := plot.NewScatter(pts)
		if err != nil {
			log.Fatal(err)
		}
		s.GlyphStyle.Radius = plot.Points(2)
		p.Add(s)
	}
	consSave(p, dir, "stochastic_poisson.png", 20*plot.Centimeter, 6*plot.Centimeter)

	// Life: one row of a random grid over 64 generations (a dot is a live cell, time runs upwards)
	genny.Seed(52)
	life := automaton.NewLifeRandom(48, 24, automaton.Conway, 0.35)
	var strip plot.XYs
	var population []float64
	for gen := range 64 {
		for x, alive := range life.Row(12) {
			if alive {
				strip = append(strip, plot.XY{X: float64(x), Y: float64(gen)})
			}
		}
		population = append(population, float64(life.Population()))
		life.Step()
	}
	stochScatter(dir, "life_strip.png", "Game of Life: row 12 over 64 generations", "column", "generation",
		20*plot.Centimeter, 12*plot.Centimeter, strip)
	i := 0
	lines(dir, "life_population.png", "Game of Life: live cells", 64, []string{"population"},
		func() float64 { i++; return population[i-1] })

	// the fitness of an evolution towards an arch: best and mean of the population
	genny.Seed(53)
	ga := evolve.New(evolve.Config{
		Population: 30, Length: 12, Min: 0, Max: 14,
		Fitness: evolve.Weighted(
			evolve.Weight(evolve.Contour(evolve.Arch(12)), 3),
			evolve.Weight(evolve.Smoothness(), 1),
			evolve.Weight(evolve.InScale(scale.New(0, scale.Major)), 0.5)),
		MutationRate: 0.1, Elitism: 2,
	})
	const gens = 80
	best, mean := make([]float64, gens), make([]float64, gens)
	for g := range gens {
		_, fit := ga.Population()
		sum := 0.0
		for _, f := range fit {
			sum += f
		}
		best[g], mean[g] = fit[0], sum/float64(len(fit))
		ga.Step()
	}
	b, m := 0, 0
	lines(dir, "evolve_fitness.png", "Evolution towards an arch: fitness per generation", gens,
		[]string{"best", "mean"},
		func() float64 { b++; return best[b-1] }, func() float64 { m++; return mean[m-1] })

	// the Messiaen modes as pitch sets over one octave (the line of a mode, a dot per pitch class)
	var modes []plot.XYs
	for i, mode := range [][]float64{scale.Mode1, scale.Mode2, scale.Mode3, scale.Mode4, scale.Mode5, scale.Mode6, scale.Mode7} {
		pts := plot.XYs{}
		for _, s := range mode {
			pts = append(pts, plot.XY{X: s, Y: float64(i + 1)})
		}
		modes = append(modes, pts)
	}
	stochScatter(dir, "messiaen_modes.png", fmt.Sprintf("Messiaen's modes (bottom to top: 1 to 7; transpositions %d, %d, %d, %d, %d, %d, %d)",
		scale.Transpositions(scale.Mode1), scale.Transpositions(scale.Mode2), scale.Transpositions(scale.Mode3),
		scale.Transpositions(scale.Mode4), scale.Transpositions(scale.Mode5), scale.Transpositions(scale.Mode6),
		scale.Transpositions(scale.Mode7)), "semitones above the root", "mode", 20*plot.Centimeter, 9*plot.Centimeter, modes...)
}
