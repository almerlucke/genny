package evolve_test

import (
	"math"
	"slices"
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/evolve"
	"github.com/almerlucke/genny/scale"
)

func config() evolve.Config {
	return evolve.Config{
		Population: 30, Length: 12, Min: 0, Max: 14,
		Fitness:      evolve.Weighted(evolve.Weight(evolve.Contour(evolve.Arch(12)), 2), evolve.Weight(evolve.Smoothness(), 1)),
		MutationRate: 0.1, Elitism: 2,
	}
}

func TestNewPanics(t *testing.T) {
	good := config()
	for name, mod := range map[string]func(*evolve.Config){
		"population 0":    func(c *evolve.Config) { c.Population = 0 },
		"population < 0":  func(c *evolve.Config) { c.Population = -3 },
		"negative length": func(c *evolve.Config) { c.Length = -1 },
		"max below min":   func(c *evolve.Config) { c.Min, c.Max = 5, 4 },
		"no fitness":      func(c *evolve.Config) { c.Fitness = nil },
	} {
		cfg := good
		mod(&cfg)
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: no panic", name)
				}
			}()
			evolve.New(cfg)
		}()
	}
}

func TestFitnessNeverDropsWithElitism(t *testing.T) {
	genny.Seed(1)
	ga := evolve.New(config())
	prev := ga.Best().Fitness
	first := prev
	for gen := 1; gen <= 100; gen++ {
		ga.Step()
		b := ga.Best()
		if b.Fitness < prev {
			t.Fatalf("generation %d: fitness fell from %v to %v", gen, prev, b.Fitness)
		}
		if b.Generation != gen || ga.Generation() != gen {
			t.Fatalf("generation count %d, %d", b.Generation, ga.Generation())
		}
		prev = b.Fitness
	}
	if prev < first+0.1 {
		t.Errorf("fitness went from %v to %v, want a clear improvement", first, prev)
	}
}

func TestRunAndRange(t *testing.T) {
	genny.Seed(2)
	cfg := config()
	cfg.Min, cfg.Max = 3, 9
	ga := evolve.New(cfg)
	b := ga.Run(40)
	if b.Generation != 40 || len(b.Phrase) != 12 {
		t.Fatalf("best %+v", b)
	}
	pop, fit := ga.Population()
	if len(pop) != 30 || len(fit) != 30 {
		t.Fatal("population size changed")
	}
	for i, p := range pop {
		if i > 0 && fit[i] > fit[i-1] {
			t.Fatal("population is not ordered best first")
		}
		for _, v := range p {
			if v < 3 || v > 9 {
				t.Fatalf("note %d outside the range", v)
			}
		}
	}
	if !slices.Equal(pop[0], b.Phrase) {
		t.Error("best is not the first of the population")
	}
	if got := ga.Run(0); got.Generation != 40 {
		t.Error("Run(0) should not step")
	}
	if got := ga.Run(-5); got.Generation != 40 {
		t.Error("Run(-5) should not step")
	}
	b.Phrase[0] = 99 // a copy: the population is not changed
	if ga.Best().Phrase[0] == 99 {
		t.Error("Best shares memory with the population")
	}
}

func TestDeterministicUnderSeed(t *testing.T) {
	run := func(seed uint64) evolve.Best {
		genny.Seed(seed)
		return evolve.New(config()).Run(25)
	}
	a, b, c := run(5), run(5), run(6)
	if !slices.Equal(a.Phrase, b.Phrase) || a.Fitness != b.Fitness {
		t.Error("the same seed gave other results")
	}
	if slices.Equal(a.Phrase, c.Phrase) && a.Fitness == c.Fitness {
		t.Error("another seed gave the same result")
	}
}

func TestEmptyGenome(t *testing.T) {
	genny.Seed(3)
	cfg := config()
	cfg.Length = 0
	ga := evolve.New(cfg)
	b := ga.Run(5)
	if len(b.Phrase) != 0 {
		t.Errorf("phrase %v", b.Phrase)
	}
	p := evolve.NewPlayer(ga, 1)
	if !p.Done() || p.Continuous() || p.Generate() != 0 {
		t.Error("a player of empty phrases should be done and give 0")
	}
}

func TestSingleGeneAndTinyPopulation(t *testing.T) {
	genny.Seed(4)
	cfg := config()
	cfg.Length, cfg.Population, cfg.Elitism = 1, 1, 5 // elitism is clamped to the population
	ga := evolve.New(cfg)
	ga.Run(10)
	if len(ga.Best().Phrase) != 1 {
		t.Error("single gene phrase")
	}
}

func TestFixedRangeAndStart(t *testing.T) {
	genny.Seed(5)
	cfg := config()
	cfg.Min, cfg.Max = 7, 7
	ga := evolve.New(cfg)
	ga.Run(3)
	for _, v := range ga.Best().Phrase {
		if v != 7 {
			t.Fatalf("note %d in a range of one value", v)
		}
	}

	// a start phrase that is clipped to the range and padded
	cfg = config()
	cfg.Fitness = evolve.Target([]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11})
	cfg.Start = [][]int{{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, {100, -100}}
	ga = evolve.New(cfg)
	if b := ga.Best(); b.Fitness != 1 || b.Generation != 0 {
		t.Errorf("start phrase not used: %+v", b)
	}
	pop, _ := ga.Population()
	for _, p := range pop {
		for _, v := range p {
			if v < 0 || v > 14 {
				t.Fatalf("start was not clipped: %v", p)
			}
		}
	}
}

func TestNaNFitnessIsWorst(t *testing.T) {
	genny.Seed(6)
	cfg := config()
	cfg.Fitness = func(p []int) float64 {
		if p[0]%2 == 0 {
			return math.NaN()
		}
		return float64(p[0])
	}
	b := evolve.New(cfg).Run(20)
	if b.Phrase[0]%2 == 0 || math.IsNaN(b.Fitness) {
		t.Errorf("best %+v", b)
	}
}

func TestPlayerPlaysAndImproves(t *testing.T) {
	genny.Seed(7)
	cfg := config()
	cfg.Start = nil
	ga := evolve.New(cfg)
	initial := ga.Best().Phrase
	p := evolve.NewPlayer(ga, 20)
	if !p.Continuous() || p.Done() {
		t.Fatal("player should be endless")
	}
	var first []int
	for range 12 {
		first = append(first, p.Generate())
	}
	if !slices.Equal(first, initial) {
		t.Errorf("the first cycle should be the best of the initial population: %v vs %v", first, initial)
	}
	if p.Cycles() != 1 || ga.Generation() != 20 {
		t.Errorf("after one cycle: %d cycles, generation %d", p.Cycles(), ga.Generation())
	}
	if want := ga.Best().Phrase; !slices.Equal(p.Phrase(), want) {
		t.Error("the next phrase should be the new best")
	}
	for range 12 * 5 {
		p.Generate()
	}
	if ga.Generation() != 120 {
		t.Errorf("generation %d, want 120", ga.Generation())
	}
	p.Generate()
	p.Reset()
	if got := p.Generate(); got != p.Phrase()[0] {
		t.Error("Reset should restart the phrase")
	}
	if evolve.NewPlayer(ga, -3) == nil {
		t.Error("negative generations are clamped")
	}
}

func TestNoCrossoverStillMutates(t *testing.T) {
	genny.Seed(8)
	cfg := config()
	cfg.CrossoverRate = -1
	cfg.MutationRate = 5 // clamped to 1
	ga := evolve.New(cfg)
	ga.Run(10)
	if len(ga.Best().Phrase) != 12 {
		t.Error("phrase length changed")
	}
}

func TestFitnessHelpers(t *testing.T) {
	key := scale.New(0, scale.Major)
	flat := []int{5, 5, 5, 5, 5, 5, 5, 5}
	jumpy := []int{0, 12, 0, 12, 0, 12, 0, 12}
	arch := []int{0, 2, 4, 6, 6, 4, 2, 0}
	valley := []int{6, 4, 2, 0, 0, 2, 4, 6}
	for _, c := range []struct {
		name     string
		f        evolve.Fitness
		p        []int
		min, max float64
	}{
		{"smoothness flat", evolve.Smoothness(), flat, 1, 1},
		{"smoothness jumpy", evolve.Smoothness(), jumpy, 0.07, 0.09},
		{"smoothness one note", evolve.Smoothness(), []int{3}, 1, 1},
		{"smoothness empty", evolve.Smoothness(), nil, 1, 1},
		{"contour arch", evolve.Contour(evolve.Arch(8)), arch, 0.85, 1},
		{"contour valley", evolve.Contour(evolve.Arch(8)), valley, 0, 0.4},
		{"contour flat", evolve.Contour(evolve.Arch(8)), flat, 0.2, 0.8},
		{"contour stretched target", evolve.Contour([]float64{0, 1}), []int{0, 1, 2, 3, 4}, 0.99, 1},
		{"contour empty phrase", evolve.Contour(evolve.Arch(8)), nil, 1, 1},
		{"contour empty target", evolve.Contour(nil), arch, 1, 1},
		{"in scale all", evolve.InScale(key), []int{0, 2, 4, 5, 7, 12}, 1, 1},
		{"in scale half", evolve.InScale(key), []int{0, 1, 4, 6}, 0.5, 0.5},
		{"in scale empty", evolve.InScale(key), nil, 1, 1},
		{"variety none", evolve.IntervalVariety(), flat, 1.0 / 7, 1.0 / 7},
		{"variety all", evolve.IntervalVariety(), []int{0, 1, 3, 6, 10}, 1, 1},
		{"variety ignores direction", evolve.IntervalVariety(), []int{0, 2, 0, 2}, 1.0 / 3, 1.0 / 3},
		{"variety short", evolve.IntervalVariety(), []int{1}, 1, 1},
		{"target same", evolve.Target(arch), arch, 1, 1},
		{"target off by one", evolve.Target([]int{1, 1}), []int{2, 2}, 0.5, 0.5},
		{"target length differs", evolve.Target([]int{1, 1}), []int{1, 1, 1, 1}, 0.66, 0.67},
		{"target both empty", evolve.Target(nil), nil, 1, 1},
	} {
		if got := c.f(c.p); got < c.min-1e-9 || got > c.max+1e-9 {
			t.Errorf("%s: %v, want %v..%v", c.name, got, c.min, c.max)
		}
	}
}

func TestWeighted(t *testing.T) {
	one := func([]int) float64 { return 1 }
	zero := func([]int) float64 { return 0 }
	if got := evolve.Weighted(evolve.Weight(one, 3), evolve.Weight(zero, 1))(nil); got != 0.75 {
		t.Errorf("weighted mean %v, want 0.75", got)
	}
	if got := evolve.Weighted()(nil); got != 0 {
		t.Errorf("no parts: %v", got)
	}
	if got := evolve.Weighted(evolve.Weight(one, 0))(nil); got != 0 {
		t.Errorf("zero weights: %v", got)
	}
	for name, f := range map[string]func(){
		"negative": func() { evolve.Weighted(evolve.Weight(one, -1)) },
		"nan":      func() { evolve.Weighted(evolve.Weight(one, math.NaN())) },
		"nil":      func() { evolve.Weighted(evolve.Weight(nil, 1)) },
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: no panic", name)
				}
			}()
			f()
		}()
	}
}

func TestArch(t *testing.T) {
	a := evolve.Arch(5)
	if len(a) != 5 || a[0] != 0 || a[2] != 1 || math.Abs(a[4]) > 1e-12 {
		t.Errorf("arch %v", a)
	}
	if len(evolve.Arch(0)) != 0 || len(evolve.Arch(-2)) != 0 {
		t.Error("empty arch")
	}
	if got := evolve.Arch(1); len(got) != 1 || got[0] != 1 {
		t.Errorf("arch of one: %v", got)
	}
}
