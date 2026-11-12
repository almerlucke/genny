package main

import (
	"github.com/almerlucke/genny/float/export/plot"
	"log"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/chord"
	"github.com/almerlucke/genny/constrain"
	"github.com/almerlucke/genny/float/iter"
	"github.com/almerlucke/genny/float/iter/updaters/chaos"
	"github.com/almerlucke/genny/function"
	"github.com/almerlucke/genny/grammar"
	"github.com/almerlucke/genny/scale"
)

// consSave draws a plot to a PNG file
func consSave(p *plot.Plot, dir, name string, w, h plot.Length) {
	if err := p.Save(w, h, path(dir, name)); err != nil {
		log.Fatalf("%s: %v", name, err)
	}
}

// consPortrait plots the states of an iterator against each other (x against y) as dots
func consPortrait(dir, name, title, xLabel, yLabel string, orbits ...[][2]float64) {
	p := plot.New()
	p.Title.Text, p.X.Label.Text, p.Y.Label.Text = title, xLabel, yLabel
	for _, orbit := range orbits {
		pts := make(plot.XYs, len(orbit))
		for i, v := range orbit {
			pts[i] = plot.XY{X: v[0], Y: v[1]}
		}
		s, err := plot.NewScatter(pts)
		if err != nil {
			log.Fatal(err)
		}
		s.GlyphStyle.Radius = plot.Points(0.7)
		p.Add(s)
	}
	consSave(p, dir, name, 12*plot.Centimeter, 12*plot.Centimeter)
}

// consOrbit runs an iterator and collects its first two values after a number of steps to settle
func consOrbit(it *iter.Iterator, settle, n int) [][2]float64 {
	for range settle {
		it.Generate()
	}
	out := make([][2]float64, n)
	for i := range out {
		f := it.Generate()
		out[i] = [2]float64{f[0], f[1]}
	}
	return out
}

// consStems draws vertical stems (position, height)
func consStems(dir, name, title, xLabel, yLabel string, xs, ys []float64) {
	p := plot.New()
	p.Title.Text, p.X.Label.Text, p.Y.Label.Text = title, xLabel, yLabel
	for i := range xs {
		l, err := plot.NewLine(plot.XYs{{X: xs[i], Y: 0}, {X: xs[i], Y: ys[i]}})
		if err != nil {
			log.Fatal(err)
		}
		p.Add(l)
	}
	consSave(p, dir, name, 20*plot.Centimeter, 8*plot.Centimeter)
}

// plotsConstraints writes the plots of the constraints, the grammar, the chaotic modulators and the
// spectral chords to dir
func plotsConstraints(dir string) {
	// the same wild source, unconstrained and with the constraints on top, and the two walks
	wild := func() *function.Function[int] { return function.New(func() int { return 48 + genny.Intn(25) }) }
	genny.Seed(41)
	raw := wild()
	genny.Seed(41)
	allowed := constrain.AllowInterval[int](wild(), []int{1, 2, 3, 4})
	genny.Seed(41)
	noRepeat := constrain.RemoveRepetitions[int](constrain.AllowInterval[int](wild(), []int{0, 5, 7, 12}))
	minor := scale.New(48, scale.Minor)
	var notes []int
	for p := 48; p <= 72; p++ {
		if minor.Quantize(float64(p)) == float64(p) {
			notes = append(notes, p)
		}
	}
	genny.Seed(41)
	stepwise := constrain.ScaleIntervals(60, []int{2, -1, 3, -4, 1, 2, -3}, notes)
	genny.Seed(41)
	leaps := constrain.RandomIntervals(48, 72, 2, 3, 7, 12)
	asFloat := func(g genny.Generator[int]) func() float64 { return func() float64 { return float64(g.Generate()) } }
	lines(dir, "constrain_walks.png", "Constrained pitch lines", 60,
		[]string{"unconstrained", "AllowInterval 1-4", "AllowInterval 0,5,7,12", "ScaleIntervals in C minor", "RandomIntervals 2,3,7,12"},
		asFloat(raw), asFloat(allowed), asFloat(noRepeat), asFloat(stepwise), asFloat(leaps))

	// the chaotic modulators over time, and as phase portraits
	// two lines from one oscillator: the first advances it, the second reads the same frame
	var frame []float64
	step := func(it *iter.Iterator) func() float64 {
		return func() float64 { frame = append(frame[:0], it.Generate()...); return frame[0] }
	}
	second := func() float64 { return frame[1] }
	lines(dir, "duffing_time.png", "Duffing oscillator", 600, []string{"x", "velocity"},
		step(iter.New([]float64{1, 0}, chaos.NewDuffing())), second)
	consPortrait(dir, "duffing.png", "Duffing oscillator", "x", "velocity",
		consOrbit(iter.New([]float64{1, 0}, chaos.NewDuffing()), 200, 12000))

	lines(dir, "pendulum_time.png", "Driven pendulum", 600, []string{"angle", "angular velocity"},
		step(iter.New([]float64{0.2, 0}, chaos.NewPendulum())), second)
	consPortrait(dir, "pendulum.png", "Driven pendulum", "angle", "angular velocity",
		consOrbit(iter.New([]float64{0.2, 0}, chaos.NewPendulum()), 200, 12000))

	var orbits [][][2]float64
	for i := range 9 {
		start := []float64{0.3 + 0.7*float64(i), 0.2 + 0.6*float64(i%4)}
		orbits = append(orbits, consOrbit(iter.New(start, chaos.NewStandardMap(1.1)), 0, 1500))
	}
	consPortrait(dir, "standardmap.png", "Standard map, K = 1.1", "theta", "p", orbits...)

	// spectral chords: the partials as pitches, stretched and squeezed, and as stems of a spectrum
	partial := func(distortion float64, sub bool) func() float64 {
		n := 0
		return func() float64 {
			n++
			c := chord.Harmonic(55, []int{n}, distortion, sub)
			if len(c) == 0 {
				return 0
			}
			return c[0]
		}
	}
	lines(dir, "spectral_partials.png", "Pitch of the partials of 55 Hz", 16,
		[]string{"harmonic", "stretched 1.05", "squeezed 0.95", "subharmonic (from 440 Hz)"},
		partial(1, false), partial(1.05, false), partial(0.95, false), func() func() float64 {
			n := 0
			return func() float64 { n++; return chord.Harmonic(440, []int{n}, 1, true)[0] }
		}())
	var xs, ys []float64
	for n := 1; n <= 24; n++ {
		xs = append(xs, chord.Harmonic(110, []int{n}, 1, false)[0])
		ys = append(ys, 1/float64(n))
	}
	consStems(dir, "spectral_stems.png", "Harmonic series on 110 Hz (height 1/n)", "MIDI pitch", "amplitude", xs, ys)
	fm := chord.FM(440, 110, 5)
	var fx, fy []float64
	for _, p := range fm {
		fx = append(fx, p)
		fy = append(fy, 1)
	}
	consStems(dir, "fm.png", "FM 440 Hz by 110 Hz, 5 orders", "MIDI pitch", "", fx, fy)

	// a phrase from a grammar: three derivations in a row as scale degrees of C minor
	genny.Seed(42)
	g := grammar.New[int]().
		Rule("phrase", grammar.Choice(2, grammar.Ref[int]("motif"), grammar.Ref[int]("motif")),
			grammar.Choice(1, grammar.Ref[int]("motif"), grammar.Ref[int]("answer"), grammar.Ref[int]("motif"))).
		Rule("motif", grammar.Seq(0, 2, 4), grammar.Seq(4, 2, 0), grammar.Choice(1, grammar.Term(2), grammar.Term(4), grammar.Ref[int]("motif"))).
		Rule("answer", grammar.Seq(7, 4, 2, 0))
	var phrase []float64
	for range 3 {
		for _, d := range g.Expand("phrase", 4) {
			phrase = append(phrase, minor.Degree(d+7))
		}
	}
	i := 0
	lines(dir, "grammar_phrase.png", "Three derivations of one grammar", len(phrase), []string{"pitch"},
		func() float64 { v := phrase[i]; i++; return v })
}
