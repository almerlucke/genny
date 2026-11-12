// Package evolve is a small genetic algorithm for phrases. A genome is a phrase of integers: scale
// degrees (play them with scale.NewDegrees) or, just as well, MIDI pitches. A population of random phrases
// is improved generation by generation by tournament selection, one point crossover, mutation of single
// notes and elitism (the best phrases survive unchanged, so the best fitness never drops). What is good is
// told by a fitness function (higher is better), for which there are helpers (Smoothness, Contour, InScale,
// IntervalVariety, Target) that combine with Weighted. All random draws come from the stream of package
// genny, so a seed repeats an evolution. NewPlayer plays the best phrase while it evolves further
package evolve

import (
	"math"
	"slices"

	"github.com/almerlucke/genny"
)

// Fitness rates a phrase: higher is better. A NaN counts as the worst possible
type Fitness func(phrase []int) float64

// Config describes a genetic algorithm
type Config struct {
	// Population is the number of phrases, at least 1
	Population int
	// Length is the number of notes of a phrase; 0 gives empty phrases
	Length int
	// Min and Max are the lowest and highest value of a note (inclusive)
	Min, Max int
	// Fitness rates a phrase, required
	Fitness Fitness
	// Tournament is the number of random phrases that compete for each parent (default 3). A larger
	// tournament presses harder
	Tournament int
	// CrossoverRate is the chance that two parents are combined with one point crossover instead of
	// copied (default 0.9 for 0; a negative value never crosses)
	CrossoverRate float64
	// MutationRate is the chance per note that it is changed: half of the time by a step of one up or
	// down, otherwise to a random value in the range. 0 means no mutation, it is clamped to 1
	MutationRate float64
	// Elitism is the number of best phrases that go to the next generation unchanged (clamped to the
	// population). With at least 1 the best fitness never gets worse
	Elitism int
	// Start phrases are put in the initial population (the rest is random); they are clipped or padded
	// to Length and to the range of notes
	Start [][]int
}

// Best is the best phrase of a population
type Best struct {
	// Phrase is a copy of the best phrase
	Phrase []int
	// Fitness of the phrase
	Fitness float64
	// Generation is the generation it was found in (0 is the initial population)
	Generation int
}

// GA is a genetic algorithm
type GA struct {
	cfg        Config
	pop        [][]int
	fit        []float64
	generation int
}

// New creates the algorithm with a random initial population (create it after genny.Seed). It panics if
// the population is smaller than 1, the length is negative, Max is below Min or there is no fitness function
func New(cfg Config) *GA {
	if cfg.Population < 1 {
		panic("evolve: population must be at least 1")
	}
	if cfg.Length < 0 {
		panic("evolve: length must not be negative")
	}
	if cfg.Max < cfg.Min {
		panic("evolve: max must not be below min")
	}
	if cfg.Fitness == nil {
		panic("evolve: a fitness function is needed")
	}
	if cfg.Tournament < 1 {
		cfg.Tournament = 3
	}
	if cfg.CrossoverRate == 0 {
		cfg.CrossoverRate = 0.9
	}
	cfg.MutationRate = min(max(cfg.MutationRate, 0), 1)
	cfg.Elitism = min(max(cfg.Elitism, 0), cfg.Population)

	g := &GA{cfg: cfg, fit: make([]float64, cfg.Population)}
	for i := range cfg.Population {
		p := make([]int, cfg.Length)
		if i < len(cfg.Start) {
			for j := range p {
				if j < len(cfg.Start[i]) {
					p[j] = min(max(cfg.Start[i][j], cfg.Min), cfg.Max)
				} else {
					p[j] = g.random()
				}
			}
		} else {
			for j := range p {
				p[j] = g.random()
			}
		}
		g.pop = append(g.pop, p)
	}
	g.evaluate()
	return g
}

func (g *GA) random() int { return g.cfg.Min + genny.Intn(g.cfg.Max-g.cfg.Min+1) }

// evaluate rates the population and orders it, best first (a stable order, so equal phrases keep theirs)
func (g *GA) evaluate() {
	idx := make([]int, len(g.pop))
	fit := make([]float64, len(g.pop))
	for i, p := range g.pop {
		idx[i] = i
		fit[i] = g.cfg.Fitness(p)
		if math.IsNaN(fit[i]) {
			fit[i] = math.Inf(-1)
		}
	}
	slices.SortStableFunc(idx, func(a, b int) int {
		switch {
		case fit[a] > fit[b]:
			return -1
		case fit[a] < fit[b]:
			return 1
		}
		return 0
	})
	pop := make([][]int, len(g.pop))
	for i, j := range idx {
		pop[i], g.fit[i] = g.pop[j], fit[j]
	}
	g.pop = pop
}

// tournament picks the best of a few random phrases (the population is ordered, so the lowest index wins)
func (g *GA) tournament() []int {
	best := len(g.pop)
	for range g.cfg.Tournament {
		best = min(best, genny.Intn(len(g.pop)))
	}
	return g.pop[best]
}

// Step makes the next generation
func (g *GA) Step() {
	next := make([][]int, 0, len(g.pop))
	for _, p := range g.pop[:g.cfg.Elitism] {
		next = append(next, slices.Clone(p))
	}
	for len(next) < len(g.pop) {
		a, b := g.tournament(), g.tournament()
		child := slices.Clone(a)
		if g.cfg.Length >= 2 && genny.Float64() < g.cfg.CrossoverRate {
			cut := 1 + genny.Intn(g.cfg.Length-1)
			copy(child[cut:], b[cut:])
		}
		for i := range child {
			if genny.Float64() < g.cfg.MutationRate {
				if genny.Float64() < 0.5 {
					d := 1
					if genny.Float64() < 0.5 {
						d = -1
					}
					child[i] = min(max(child[i]+d, g.cfg.Min), g.cfg.Max)
				} else {
					child[i] = g.random()
				}
			}
		}
		next = append(next, child)
	}
	g.pop = next
	g.generation++
	g.evaluate()
}

// Run makes a number of generations (none for 0 or less) and returns the best phrase
func (g *GA) Run(generations int) Best {
	for range generations {
		g.Step()
	}
	return g.Best()
}

// Best returns the best phrase of the current generation. With elitism it is the best ever found
func (g *GA) Best() Best {
	return Best{Phrase: slices.Clone(g.pop[0]), Fitness: g.fit[0], Generation: g.generation}
}

// Generation is the number of steps taken
func (g *GA) Generation() int { return g.generation }

// Population returns copies of the phrases, best first, and their fitness
func (g *GA) Population() ([][]int, []float64) {
	pop := make([][]int, len(g.pop))
	for i, p := range g.pop {
		pop[i] = slices.Clone(p)
	}
	return pop, slices.Clone(g.fit)
}

// Player is a generator that plays the best phrase of an evolving population, note by note. After every
// complete phrase the population evolves for a number of generations and the new best phrase follows, so
// the music gets better as it plays
type Player struct {
	ga     *GA
	perCyc int
	phrase []int
	pos    int
	cycles int
}

// NewPlayer creates a player that evolves generationsPerCycle generations (0 or less: none) after every
// phrase. The first phrase is the best of the initial population
func NewPlayer(ga *GA, generationsPerCycle int) *Player {
	return &Player{ga: ga, perCyc: max(generationsPerCycle, 0), phrase: ga.Best().Phrase}
}

// Generate returns the next note of the phrase. With a phrase length of 0 it is done at once and returns 0
func (p *Player) Generate() int {
	if len(p.phrase) == 0 {
		return 0
	}
	v := p.phrase[p.pos]
	p.pos++
	if p.pos == len(p.phrase) {
		p.pos = 0
		p.cycles++
		p.phrase = p.ga.Run(p.perCyc).Phrase
	}
	return v
}

// Cycles is the number of complete phrases played
func (p *Player) Cycles() int { return p.cycles }

// Phrase returns a copy of the phrase that is playing
func (p *Player) Phrase() []int { return slices.Clone(p.phrase) }

func (p *Player) Continuous() bool { return len(p.phrase) > 0 }

func (p *Player) Done() bool { return len(p.phrase) == 0 }

// Reset starts the current phrase again; the population keeps what it has evolved
func (p *Player) Reset() { p.pos = 0 }
