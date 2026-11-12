package harmony

import (
	"github.com/almerlucke/genny"
)

// Cadence is the way a phrase ends
type Cadence int

const (
	// Authentic ends V (or V7) - I
	Authentic Cadence = iota
	// Half ends on V, after a subdominant or the tonic
	Half
	// Plagal ends IV - I
	Plagal
	// Deceptive ends V - vi (VI in minor)
	Deceptive
)

func (c Cadence) String() string {
	return [...]string{"authentic", "half", "plagal", "deceptive"}[c]
}

// Config configures a Generator. The zero value is a four chord phrase in C major that alternates a
// half and an authentic cadence
type Config struct {
	Key Key
	// PhraseLength is the number of chords of a phrase, at least 2 (default 4)
	PhraseLength int
	// Cadences gives the weights of the cadences that end the phrases. Without weights the phrases
	// alternate half and authentic, which makes balanced antecedent and consequent phrases
	Cadences map[Cadence]float64
	// Moves overrides the weights of the grammar for the function moves, per function you give.
	// The defaults are T: T 1, S 3, D 1.5; S: S 1, D 4, T 0.5; D: T 5, D 0.3 (there is no D -> S)
	Moves map[Function]map[Function]float64
	// SeventhProb is the chance of a seventh on ii and V (default 0)
	SeventhProb float64
	// SecondaryProb is the chance that a chord becomes the dominant of the chord after it (default 0)
	SecondaryProb float64
	// BorrowedProb is the chance of iv instead of IV in a major key (default 0)
	BorrowedProb float64
}

var defaultMoves = map[Function]map[Function]float64{
	Tonic:       {Tonic: 1, Subdominant: 3, Dominant: 1.5},
	Subdominant: {Subdominant: 1, Dominant: 4, Tonic: 0.5},
	Dominant:    {Tonic: 5, Dominant: 0.3},
}

// weights of the degrees per function: tonic, subdominant, dominant
var degreeWeights = map[Function][][2]float64{
	Tonic:       {{0, 4}, {5, 2}, {2, 0.6}},
	Subdominant: {{3, 3}, {1, 3}},
	Dominant:    {{4, 6}, {6, 1}},
}

// Generator generates the chords of phrases that end in cadences. It implements
// genny.Generator[Chord] and never ends. All its randomness comes from the genny stream (genny.Seed)
type Generator struct {
	cfg     Config
	queue   []Chord
	phrases int
}

// New creates a Generator. It panics on a PhraseLength of 1 or less than 0 (0 means the default),
// a probability outside [0, 1] or a negative weight
func New(cfg Config) *Generator {
	if cfg.PhraseLength < 0 || cfg.PhraseLength == 1 {
		panic("harmony: PhraseLength must be at least 2")
	}
	if cfg.PhraseLength == 0 {
		cfg.PhraseLength = 4
	}
	for _, p := range []float64{cfg.SeventhProb, cfg.SecondaryProb, cfg.BorrowedProb} {
		if p < 0 || p > 1 {
			panic("harmony: probability outside [0, 1]")
		}
	}
	for _, w := range cfg.Cadences {
		if w < 0 {
			panic("harmony: negative cadence weight")
		}
	}
	for _, m := range cfg.Moves {
		for _, w := range m {
			if w < 0 {
				panic("harmony: negative move weight")
			}
		}
	}
	cfg.Key.Tonic = mod12(cfg.Key.Tonic)
	return &Generator{cfg: cfg}
}

// SetKey changes the key from the next phrase on; chords already queued keep their key
func (g *Generator) SetKey(k Key) {
	k.Tonic = mod12(k.Tonic)
	g.cfg.Key = k
}

// Generate returns the next chord, starting a new phrase when the last one is used up
func (g *Generator) Generate() Chord {
	if len(g.queue) == 0 {
		g.queue = g.Phrase(g.cfg.PhraseLength, g.nextCadence())
		g.phrases++
	}
	c := g.queue[0]
	g.queue = g.queue[1:]
	return c
}

func (g *Generator) Continuous() bool { return true }
func (g *Generator) Done() bool       { return false }

// Reset starts again with a new phrase (the alternation of the cadences starts over)
func (g *Generator) Reset() {
	g.queue = nil
	g.phrases = 0
}

func (g *Generator) nextCadence() Cadence {
	if len(g.cfg.Cadences) == 0 {
		if g.phrases%2 == 0 {
			return Half
		}
		return Authentic
	}
	var ws []float64
	for c := Authentic; c <= Deceptive; c++ {
		ws = append(ws, g.cfg.Cadences[c])
	}
	if i := pick(ws); i >= 0 {
		return Cadence(i)
	}
	return Authentic
}

// pick returns an index drawn by weight, or -1 if all weights are zero
func pick(ws []float64) int {
	total := 0.0
	for _, w := range ws {
		total += w
	}
	if total <= 0 {
		return -1
	}
	x := genny.Float64() * total
	for i, w := range ws {
		if x < w {
			return i
		}
		x -= w
	}
	for i := len(ws) - 1; i >= 0; i-- {
		if ws[i] > 0 {
			return i
		}
	}
	return -1
}

func (g *Generator) nextFunction(from Function) Function {
	m, ok := g.cfg.Moves[from]
	if !ok {
		m = defaultMoves[from]
	}
	ws := []float64{m[Tonic], m[Subdominant], m[Dominant]}
	if i := pick(ws); i >= 0 {
		return Function(i)
	}
	return Tonic
}

// chordFor picks a chord of a function. prev, if not nil, is not repeated
func (g *Generator) chordFor(fn Function, prev *Chord) Chord {
	opts := degreeWeights[fn]
	ws := make([]float64, len(opts))
	for i, o := range opts {
		ws[i] = o[1]
		if prev != nil && degreeOf(*prev) == int(o[0]) {
			ws[i] = 0
		}
	}
	i := pick(ws)
	if i < 0 {
		i = 0
	}
	return g.dress(int(opts[i][0]))
}

// dress makes the chord of a degree, with the chances of a seventh or borrowed chord
func (g *Generator) dress(degree int) Chord {
	k := g.cfg.Key
	if degree == 3 && !k.Minor && genny.Float64() < g.cfg.BorrowedProb {
		c, _ := Borrowed(k, 3)
		return c
	}
	seventh := (degree == 1 || degree == 4) && genny.Float64() < g.cfg.SeventhProb
	return Diatonic(k, degree, seventh)
}

// Phrase makes a phrase of n chords (at least 2) that starts on the tonic and ends in the given
// cadence. The last two chords are the cadence; the chords before follow the function grammar.
// It does not touch the queue of Generate
func (g *Generator) Phrase(n int, c Cadence) []Chord {
	if n < 2 {
		n = 2
	}
	k := g.cfg.Key
	body := n - 2
	out := make([]Chord, 0, n)

	fns := make([]Function, body)
	for i := range fns {
		if i > 0 {
			fns[i] = g.nextFunction(fns[i-1])
		}
	}
	// the chord before the cadence must lead into it
	if body > 0 {
		last := &fns[body-1]
		switch {
		case (c == Authentic || c == Deceptive) && *last == Dominant:
			*last = Subdominant
			if body > 1 && fns[body-2] == Dominant {
				*last = Tonic // no way back from a dominant to a subdominant
			}
		case c == Plagal && *last != Tonic:
			*last = Tonic
		}
	}
	for i, fn := range fns {
		var prev *Chord
		if i > 0 {
			prev = &out[i-1]
		}
		if i == 0 {
			out = append(out, Diatonic(k, 0, false))
			continue
		}
		out = append(out, g.chordFor(fn, prev))
	}

	var prev *Chord
	if len(out) > 0 {
		prev = &out[len(out)-1]
	}
	switch c {
	case Authentic:
		out = append(out, Diatonic(k, 4, genny.Float64() < 2*g.cfg.SeventhProb), Diatonic(k, 0, false))
	case Deceptive:
		out = append(out, Diatonic(k, 4, genny.Float64() < 2*g.cfg.SeventhProb), Diatonic(k, 5, false))
	case Plagal:
		iv, ok := Borrowed(k, 3)
		if !ok || genny.Float64() >= g.cfg.BorrowedProb {
			iv = Diatonic(k, 3, false)
		}
		out = append(out, iv, Diatonic(k, 0, false))
	case Half:
		pre := Subdominant
		if body > 0 && fns[body-1] == Dominant {
			pre = Tonic
		}
		var p Chord
		if pre == Tonic {
			p = Diatonic(k, 0, false)
		} else {
			p = g.chordFor(Subdominant, prev)
		}
		out = append(out, p, Diatonic(k, 4, false))
	}

	g.addSecondaries(out, c)
	return out
}

// addSecondaries turns chords into the dominant of the chord that follows them
func (g *Generator) addSecondaries(out []Chord, c Cadence) {
	n := len(out)
	for i := 1; i < n-1; i++ {
		if (c == Authentic || c == Deceptive) && i == n-2 {
			continue
		}
		if out[i-1].Function == Dominant || out[i].Borrowed {
			continue
		}
		d := degreeOf(out[i+1])
		if d < 1 || d > 5 || out[i+1].Borrowed {
			continue
		}
		if genny.Float64() >= g.cfg.SecondaryProb {
			continue
		}
		if s, ok := SecondaryDominant(g.cfg.Key, d, genny.Float64() < g.cfg.SeventhProb); ok {
			out[i] = s
			i++ // the target stays what it is
		}
	}
}
