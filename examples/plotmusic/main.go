// plotmusic plots the music generators to see that they behave: go run ./examples/plotmusic [dir]
// writes PNG files in dir (default "plots")
package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/automaton"
	"github.com/almerlucke/genny/chord"
	"github.com/almerlucke/genny/euclid"
	"github.com/almerlucke/genny/float/dist"
	"github.com/almerlucke/genny/float/export/plot"
	"github.com/almerlucke/genny/float/iter"
	"github.com/almerlucke/genny/float/iter/updaters/chaos"
	"github.com/almerlucke/genny/float/lfo"
	"github.com/almerlucke/genny/float/lissajous"
	"github.com/almerlucke/genny/float/noise"
	"github.com/almerlucke/genny/float/rwalk"
	"github.com/almerlucke/genny/groove"
	"github.com/almerlucke/genny/mutate"
	"github.com/almerlucke/genny/numseq"
	"github.com/almerlucke/genny/progression"
	"github.com/almerlucke/genny/scale"
	"gonum.org/v1/plot/vg"
)

// frames is a float.FrameGenerator made of one function that returns the values of a step
type frames struct {
	dims int
	f    func() []float64
}

func (g *frames) Generate() []float64 { return g.f() }
func (g *frames) Dimensions() int     { return g.dims }
func (g *frames) Continuous() bool    { return true }
func (g *frames) Done() bool          { return false }
func (g *frames) Reset()              {}

// lines plots the generators next to each other, one line each
func lines(dir, name, title string, n int, names []string, gens ...func() float64) {
	g := &frames{dims: len(gens), f: func() []float64 {
		out := make([]float64, len(gens))
		for i, gen := range gens {
			out[i] = gen()
		}
		return out
	}}
	err := plot.PlotX(g, n, title, "step", "", names, 20*vg.Centimeter, 8*vg.Centimeter, path(dir, name))
	if err != nil {
		log.Fatalf("%s: %v", name, err)
	}
}

func path(dir, name string) string { return filepath.Join(dir, name) }

// gate draws a gate at a height so several rhythms can share one plot
func gate(g interface{ Generate() bool }, height float64) func() float64 {
	return func() float64 {
		if g.Generate() {
			return height + 0.8
		}
		return height
	}
}

func main() {
	dir := "plots"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Fatal(err)
	}
	genny.Seed(1)

	// LFO shapes, two cycles of 50 steps
	var shapes []func() float64
	for _, s := range []lfo.Shape{lfo.Sine, lfo.Triangle, lfo.Saw, lfo.Square, lfo.SmoothRandom} {
		shapes = append(shapes, lfo.New(s, 1, 50).Generate)
	}
	lines(dir, "lfo.png", "LFO shapes", 100,
		[]string{"sine", "triangle", "saw", "square", "smooth random"}, shapes...)

	// sample and hold, and the same with a slew limiter
	sh := lfo.New(lfo.SampleHold, 1, 25)
	slewed := lfo.NewSlew(lfo.New(lfo.SampleHold, 1, 25), 0.1, 0.1)
	genny.Seed(2)
	lines(dir, "slew.png", "Sample and hold, slewed", 100,
		[]string{"sample & hold", "slewed"}, sh.Generate, slewed.Generate)

	// noise with more octaves has more detail
	genny.Seed(3)
	var noises []func() float64
	for _, oct := range []int{1, 2, 4} {
		noises = append(noises, noise.New(0.02, oct).Generate)
	}
	lines(dir, "noise.png", "Smooth noise", 400, []string{"1 octave", "2 octaves", "4 octaves"}, noises...)

	// attractors
	for _, a := range []struct {
		name string
		it   *iter.Iterator
		n    int
	}{
		{"lorenz", iter.New([]float64{1, 1, 1}, chaos.NewLorenz(10, 28, 8.0/3, 0.01)), 2500},
		{"rossler", iter.New([]float64{1, 1, 1}, chaos.NewRossler(0.2, 0.2, 5.7, 0.05)), 2500},
	} {
		err := plot.PlotX(a.it, a.n, a.name, "step", "", []string{"x", "y", "z"}, 20*vg.Centimeter, 8*vg.Centimeter,
			path(dir, a.name+".png"))
		if err != nil {
			log.Fatal(err)
		}
	}

	// a random walk quantized to a scale, and a walk over scale degrees
	genny.Seed(4)
	minor := scale.New(57, scale.Minor)
	walk := rwalk.New(64, 52, 76, 1.2, rwalk.Bounce)
	var raw float64
	lines(dir, "quantize.png", "Random walk quantized to A minor", 80, []string{"walk", "quantized"},
		func() float64 { raw = walk.Generate(); return raw },
		func() float64 { return minor.Quantize(raw) })

	dw := scale.NewWalk(minor, 7, 0, 14, 1, 0.15, 4)
	lines(dir, "scalewalk.png", "Walk over scale degrees", 80, []string{"pitch"}, dw.Generate)

	// rhythms: euclidean 5/16 and 7/16, rule 30 and rule 90 as gates
	lines(dir, "rhythm.png", "Rhythms", 64, []string{"E(5,16)", "E(7,16)", "rule 30", "rule 90"},
		gate(euclid.New(16, 5, 0), 0), gate(euclid.New(16, 7, 0), 1.2),
		gate(automaton.NewGate(automaton.New(30, 16)), 2.4),
		gate(automaton.NewGate(automaton.New(90, 16)), 3.6))

	// a phrase that mutates: the notes of every pass, against the original
	genny.Seed(5)
	phrase := []float64{0, 2, 4, 7, 4, 2, 0, -3}
	mutating := mutate.New(phrase, mutate.Config[float64]{
		SwapProb: 0.6, RotateProb: 0.3, ReplaceProb: 0.4,
		Replace: func() float64 { return float64(genny.Intn(10) - 3) },
	})
	i := 0
	lines(dir, "mutate.png", "Mutating phrase", 8*12, []string{"original", "mutated"},
		func() float64 { i++; return phrase[(i-1)%len(phrase)] }, mutating.Generate)

	// voice leading: three voices through a progression, in root position and led
	genny.Seed(6)
	prog := progression.New(progression.Pop, 0)
	leader := chord.NewVoiceLeader(48, 84)
	var bars [][]float64
	var led [][]float64
	for range 16 {
		c := chord.Diatonic(minor, prog.Generate(), 3)
		bars = append(bars, c)
		led = append(led, leader.Transform(c))
	}
	step := 0
	lines(dir, "voiceleading.png", "Voice leading: root position vs led", 16,
		[]string{"low", "mid", "high", "low (led)", "mid (led)", "high (led)"},
		func() float64 { return bars[step][0] }, func() float64 { return bars[step][1] },
		func() float64 { return bars[step][2] },
		func() float64 { return led[step][0] }, func() float64 { return led[step][1] },
		func() float64 { s := led[step][2]; step++; return s })

	// the accents of a bar of 7/8 (2+2+3) and of 4/4 in eighths
	a78, a44 := groove.NewAccent(2, 2, 3), groove.NewAccent(2, 2, 2, 2)
	lines(dir, "accent.png", "Accents", 28, []string{"7/8 (2+2+3)", "4/4 (2+2+2+2)"}, a78.Generate, a44.Generate)

	// distributions
	genny.Seed(7)
	lines(dir, "distributions.png", "Gaussian and exponential", 100, []string{"gaussian(0, 1)", "exponential(1)"},
		dist.NewGaussian(0, 1).Generate, dist.NewExponential(1).Generate)

	// number sequences as pitch classes
	fib, primes, tm := numseq.NewFibonacci(12), numseq.NewPrimes(), numseq.NewThueMorse()
	lines(dir, "numseq.png", "Number sequences", 48, []string{"fibonacci mod 12", "primes mod 12", "thue-morse x 6"},
		func() float64 { return float64(fib.Generate()) },
		func() float64 { return float64(primes.Generate() % 12) },
		func() float64 { return float64(tm.Generate() * 6) })

	// a lissajous curve: its two outputs over time
	err := plot.PlotX(lissajous.New(3, 2, math.Pi/2, 100), 200, "Lissajous 3:2", "step", "", []string{"x", "y"},
		20*vg.Centimeter, 8*vg.Centimeter, path(dir, "lissajous.png"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("plots written to", dir)
}
