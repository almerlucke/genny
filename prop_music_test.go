package genny_test

import (
	"math"
	"slices"
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/chord"
	"github.com/almerlucke/genny/counterpoint"
	"github.com/almerlucke/genny/evolve"
	"github.com/almerlucke/genny/float/tendency"
	"github.com/almerlucke/genny/melody"
	"github.com/almerlucke/genny/ngram"
	"github.com/almerlucke/genny/ngram/style"
	"github.com/almerlucke/genny/note"
	"github.com/almerlucke/genny/row"
	"github.com/almerlucke/genny/scale"
	"github.com/almerlucke/genny/sieve"
	"github.com/almerlucke/genny/stochastic"
)

const propSeeds = 200

// counterpoint searches are slow (milliseconds each), so they get fewer seeds
const cpSeeds = 40

// randomCantus draws a stepwise cantus firmus of 6..14 notes from the genny stream
func randomCantus() []int {
	n := 6 + genny.Intn(9)
	cf := make([]int, n)
	cf[0] = 60
	for i := 1; i < n; i++ {
		cf[i] = min(max(cf[i-1]+[]int{-4, -2, -1, 1, 2, 4}[genny.Intn(6)], 48), 72)
	}
	cf[n-1] = 60
	return cf
}

// checkEvents tests the documented ranges of note events
func checkEvents(t *testing.T, name string, seed uint64, events []note.Event, lo, hi float64) {
	t.Helper()
	for i, e := range events {
		switch {
		case math.IsNaN(e.Pitch) || e.Pitch < lo || e.Pitch > hi:
			t.Fatalf("%s seed %d event %d: pitch %v outside [%v, %v]", name, seed, i, e.Pitch, lo, hi)
		case !(e.Velocity >= 0 && e.Velocity <= 1):
			t.Fatalf("%s seed %d event %d: velocity %v", name, seed, i, e.Velocity)
		case !(e.Duration > 0):
			t.Fatalf("%s seed %d event %d: duration %v", name, seed, i, e.Duration)
		case !(e.Start >= 0):
			t.Fatalf("%s seed %d event %d: start %v", name, seed, i, e.Start)
		}
	}
}

func TestPropCounterpointPassesOwnChecker(t *testing.T) {
	solved := 0
	for seed := uint64(1); seed <= cpSeeds; seed++ {
		genny.Seed(seed)
		cf := randomCantus()
		for _, side := range []counterpoint.Side{counterpoint.Above, counterpoint.Below} {
			cfg := counterpoint.Config{Side: side}
			cp, err := counterpoint.Generate(cf, cfg)
			if len(cp) != len(cf) {
				t.Fatalf("seed %d: %d notes for a cantus of %d", seed, len(cp), len(cf))
			}
			if err != nil {
				continue
			}
			solved++
			if v := cfg.Check(cf, cp); len(v) != 0 {
				t.Fatalf("seed %d side %d: %v\ncf %v\ncp %v", seed, side, v, cf, cp)
			}
		}
	}
	if solved < cpSeeds {
		t.Errorf("only %d of %d lines were solved", solved, 2*cpSeeds)
	}
}

func TestPropCounterpointReproducible(t *testing.T) {
	for seed := uint64(1); seed <= cpSeeds; seed++ {
		run := func() []int {
			genny.Seed(seed)
			cp, _ := counterpoint.Generate(randomCantus(), counterpoint.Config{})
			return cp
		}
		if a, b := run(), run(); !slices.Equal(a, b) {
			t.Fatalf("seed %d: %v != %v", seed, a, b)
		}
	}
}

// musicGens builds a generator that draws from the stream and plays n values as floats
var musicGens = map[string]func() []float64{
	"tendency": func() []float64 {
		m := tendency.NewMask(true, tendency.Segment{N: 4, A1: 0, A2: 1, Z1: 5, Z2: 9}, tendency.Segment{N: 3, A1: 3, A2: 3, Z1: 3, Z2: 4})
		return takeAny[float64](tendency.NewValue(m), 50)
	},
	"sieve": func() (out []float64) {
		s := sieve.New(3, genny.Intn(3), 0, 30).Union(sieve.New(5, genny.Intn(5), 0, 30))
		return sieve.Funnel(takeFloats(40), sieve.Scale(s, 0))
	},
	"melody": func() (out []float64) {
		for _, e := range melody.ChordTonesList(propProgression(), 6, melody.Config{Scale: scale.New(60, scale.Major)}) {
			out = append(out, e.Pitch, e.Velocity, e.Start, e.Duration)
		}
		return
	},
	"stochastic": func() (out []float64) {
		c := stochastic.NewCloud(stochastic.Config{
			Density: func(int) float64 { return 2 }, Pitch: stochastic.Uniform(48, 72),
			Velocity: stochastic.Uniform(0, 1), Duration: stochastic.Uniform(0.1, 2),
		})
		for _, e := range c.Render(20) {
			out = append(out, e.Pitch, e.Velocity, e.Start, e.Duration)
		}
		return
	},
	"evolve": func() (out []float64) {
		ga := evolve.New(evolve.Config{Population: 20, Length: 8, Min: 0, Max: 12, Fitness: evolve.Smoothness(), MutationRate: 0.2, Elitism: 1})
		return toFloats(ga.Run(15).Phrase)
	},
	"ngram": func() []float64 {
		return takeAny[float64](ngram.Learn([]float64{1, 2, 3, 2, 1, 2, 3, 4, 3, 2}, 2), 40)
	},
	"style": func() (out []float64) {
		for _, n := range propStyle().Phrase(30) {
			out = append(out, n.Pitch, n.Duration)
		}
		return
	},
	"row": func() []float64 { return row.Random(12) },
	"row.Spray": func() []float64 {
		return tendency.Spray([]float64{1, 2, 3, 4}, 0.5)
	},
}

func takeFloats(n int) []float64 {
	out := make([]float64, n)
	for i := range out {
		out[i] = genny.Float64() * 30
	}
	return out
}

func propProgression() []chord.Chord {
	c := scale.New(60, scale.Major)
	return []chord.Chord{chord.Diatonic(c, 0, 3), chord.Diatonic(c, 3, 3), chord.Diatonic(c, 4, 3), chord.Diatonic(c, 0, 3)}
}

func propStyle() *style.Style {
	var ev []note.Event
	for i, p := range []float64{60, 62, 64, 65, 67, 65, 64, 62, 60, 64, 67, 72} {
		ev = append(ev, note.Event{Pitch: p, Velocity: 0.8, Start: float64(i), Duration: 1})
	}
	return style.Learn(ev, 2, style.Intervals())
}

func TestPropSeededStreamsReproduce(t *testing.T) {
	for name, f := range musicGens {
		for seed := uint64(1); seed <= propSeeds; seed++ {
			genny.Seed(seed)
			a := f()
			genny.Seed(seed)
			b := f()
			if !slices.Equal(a, b) {
				t.Fatalf("%s seed %d: the same seed gave other output", name, seed)
			}
		}
	}
}

func TestPropRowIsPermutation(t *testing.T) {
	for seed := uint64(1); seed <= propSeeds; seed++ {
		genny.Seed(seed)
		r := row.Random(12)
		s := slices.Clone(r)
		slices.Sort(s)
		for i, v := range s {
			if v != float64(i) {
				t.Fatalf("seed %d: %v is not a permutation", seed, r)
			}
		}
	}
}

func TestPropOutputRanges(t *testing.T) {
	major := scale.New(60, scale.Major)
	for seed := uint64(1); seed <= propSeeds; seed++ {
		genny.Seed(seed)
		checkEvents(t, "melody", seed, melody.ChordTonesList(propProgression(), 6, melody.Config{Scale: major}), 60, 84)

		cloud := stochastic.NewCloud(stochastic.Config{
			Density: func(int) float64 { return 3 }, Pitch: stochastic.Uniform(36, 96),
			Velocity: stochastic.Uniform(-2, 3), Duration: stochastic.Uniform(-1, 2), Scale: major,
		})
		checkEvents(t, "cloud", seed, cloud.Render(20), 0, 127)
		checkEvents(t, "free", seed, stochastic.Free(30, 1.5, 1, 60, 5, major), 0, 127)

		st := propStyle()
		for i, n := range st.Phrase(40) {
			if n.Pitch < 55 || n.Pitch > 80 || !(n.Duration > 0) {
				t.Fatalf("style seed %d note %d: %+v", seed, i, n)
			}
		}
		ev := style.Events(st.Fill(32), 0, 0.7)
		checkEvents(t, "style", seed, ev, 55, 80)

	}
}

func TestPropDegenerateInput(t *testing.T) {
	genny.Seed(1)
	if cp, err := counterpoint.Generate(nil, counterpoint.Config{}); cp != nil || err != nil {
		t.Errorf("empty cantus: %v %v", cp, err)
	}
	for _, cf := range [][]int{{60}, {60, 60}, {0, 127}, {-5, 300, -5}, {60, 60, 60, 60, 60, 60}} {
		for _, side := range []counterpoint.Side{counterpoint.Above, counterpoint.Below} {
			counterpoint.Generate(cf, counterpoint.Config{Side: side})
			counterpoint.GenerateSecond(cf, counterpoint.Config{Side: side, MaxLeap: -3, Budget: 100})
		}
	}
	if ev := melody.ChordTonesList(nil, 4, melody.Config{Scale: scale.New(60, scale.Major)}); len(ev) != 0 {
		t.Errorf("no chords gave %d events", len(ev))
	}
	if ev := melody.ChordTonesList(propProgression(), 0, melody.Config{Scale: scale.New(60, scale.Major)}); len(ev) != 0 {
		t.Errorf("no bars gave %d events", len(ev))
	}
	g := sieve.NewGen(sieve.New(3, 0, 5, 1), true)
	if !g.Done() || g.Generate() != 0 {
		t.Error("empty sieve gen")
	}
	if got := sieve.Funnel([]float64{1, 2}, nil); !slices.Equal(got, []float64{1, 2}) {
		t.Errorf("funnel without values %v", got)
	}
	if ev := stochastic.NewCloud(stochastic.Config{Density: func(int) float64 { return 1 }, Pitch: stochastic.Uniform(0, 1)}).Render(-3); len(ev) != 0 {
		t.Error("negative steps")
	}
	m := tendency.NewMask(true)
	v := tendency.NewValue(m)
	_ = v.Generate()
	if tendency.Spray(nil, 1) != nil && len(tendency.Spray(nil, 1)) != 0 {
		t.Error("spray of nothing")
	}
	if len(tendency.Anchor(nil, 0, 0, 1, 0, 1)) != 0 {
		t.Error("anchor of zero values")
	}
	ga := evolve.New(evolve.Config{Population: 1, Length: 0, Fitness: evolve.Smoothness()})
	p := evolve.NewPlayer(ga, 3)
	if p.Generate() != 0 || !p.Done() {
		t.Error("empty genome player")
	}
	if len(row.Random(0)) != 0 {
		t.Error("row of 0")
	}
}

func TestPropVoiceLeaderRange(t *testing.T) {
	for seed := uint64(1); seed <= propSeeds; seed++ {
		genny.Seed(seed)
		vl := chord.NewVoiceLeader(48, 72)
		for range 6 {
			c := chord.New(float64(40+genny.Intn(30)), []float64{0, 4, 7})
			out := vl.Transform(c)
			if len(out) != len(c) {
				t.Fatalf("seed %d: %d notes became %d", seed, len(c), len(out))
			}
		}
	}
}
