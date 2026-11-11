package main

import (
	"fmt"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/listtools"
	"github.com/almerlucke/genny/pulse"
	"github.com/almerlucke/genny/sequence"
	"github.com/almerlucke/genny/sieve"
	"github.com/almerlucke/genny/split"
)

// plotsRhythm writes the plots of the rhythm tools: sieves as gates, the pulse hierarchy and its
// thinning, split durations and the list tools
func plotsRhythm(dir string) {
	genny.Seed(11)

	// sieves as membership strips: two residue classes, their union and intersection
	thirds := sieve.New(3, 0, 0, 47)
	fives := sieve.New(5, 1, 0, 47)
	strip := func(s sieve.Sieve, height float64) func() float64 {
		return gate(sieve.NewGate(s, 48), height)
	}
	lines(dir, "sieve_gates.png", "Sieves as gates (48 steps)", 48,
		[]string{"intersection", "union", "5n+1", "3n"},
		strip(thirds.Intersection(fives), 0), strip(thirds.Union(fives), 1.2),
		strip(fives, 2.4), strip(thirds, 3.6))

	// the hierarchy of 3+2+2 and 4/4 in sixteenths, as importance per pulse
	h12 := pulse.Hierarchy(3, 2, 2)
	h16 := pulse.Hierarchy(2, 2, 2, 2)
	floats := func(ind []int) func() float64 {
		g := sequence.NewLoop(func() []float64 {
			out := make([]float64, len(ind))
			for i, v := range ind {
				out[i] = float64(v)
			}
			return out
		}()...)
		return g.Generate
	}
	lines(dir, "pulse_hierarchy.png", "Importance of the pulses (two bars)", 32,
		[]string{"2,2,2,2", "3,2,2"}, floats(h16), floats(h12))

	// the same rhythm thinned at several levels: higher is sparser
	var gates []func() float64
	var names []string
	for i, level := range []int{14, 12, 8, 0} {
		gates = append(gates, gate(pulse.NewConstGate(h16, level), float64(i)*1.2))
		names = append(names, fmt.Sprintf("level %d", level))
	}
	lines(dir, "pulse_thin.png", "A rhythm thinned by importance (two bars)", 32, names, gates...)

	// a level that follows a ramp: the rhythm fades in over 8 bars of 16
	level := 16.0
	fadeIn := pulse.NewGate(h16, funcGen(func() float64 {
		level = max(level-0.125, 0)
		return level
	}))
	lines(dir, "pulse_fade.png", "A rhythm fading in as the level drops", 128,
		[]string{"fade", "level / 8"}, gate(fadeIn, 3), func() float64 { return level / 8 })

	// recursive splitting: durations of a few cycles, and two metric divisions
	for _, p := range []struct {
		name string
		prob float64
	}{{"split_low.png", 0.35}, {"split_high.png", 0.8}} {
		genny.Seed(4)
		g := split.NewGen(8, 0.25, p.prob, 2)
		lines(dir, p.name, "Split durations of 8, probability "+fmt.Sprint(p.prob), 40,
			[]string{"duration"}, g.Generate)
	}
	m := sequence.NewLoop(append(split.Metric(8, 3), split.Metric(8, 4)...)...)
	lines(dir, "split_metric.png", "Metric: 3 and 4 equal parts of 8", 14, []string{"duration"}, m.Generate)

	// the list tools: jump, grow, stutter, group
	notes := make([]float64, 12)
	for i := range notes {
		notes[i] = float64(60 + 2*i)
	}
	jump := listtools.NewJump(notes, 3, 1, 4)
	lines(dir, "jump.png", "Jump: 3 values, then a leap of 4", 36, []string{"pitch"}, jump.Generate)

	geo := listtools.NewGrow(1, listtools.Geometric, 1.5, 0, 12, listtools.Restart)
	ari := listtools.NewGrow(0, listtools.Arithmetic, 1.25, 0, 10, listtools.Clip)
	lines(dir, "grow.png", "Grow: geometric (restart) and arithmetic (clip)", 30,
		[]string{"geometric", "arithmetic"}, geo.Generate, ari.Generate)

	st := listtools.NewStutter[float64](sequence.NewLoop(1.0, 3, 2, 5, 4), 2, 3)
	gr := listtools.NewGroup[float64](sequence.NewLoop(1.0, 2, 3, 4), sequence.NewLoop(1, 2, 3))
	lines(dir, "stutter_group.png", "Stutter (after 2, 3 times) and group (1, 2, 3 times)", 36,
		[]string{"group", "stutter"}, gr.Generate, st.Generate)
}

// funcGen is a generator of floats from a function
type funcGen func() float64

func (f funcGen) Generate() float64 { return f() }
func (f funcGen) Continuous() bool  { return true }
func (f funcGen) Done() bool        { return false }
func (f funcGen) Reset()            {}
