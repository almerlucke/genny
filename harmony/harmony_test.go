package harmony

import (
	"reflect"
	"testing"

	"github.com/almerlucke/genny"
)

func numerals(cs []Chord) []string {
	out := make([]string, len(cs))
	for i, c := range cs {
		out[i] = c.Numeral
	}
	return out
}

func TestDiatonic(t *testing.T) {
	tests := []struct {
		key      Key
		degree   int
		seventh  bool
		numeral  string
		name     string
		function Function
	}{
		{Key{0, false}, 0, false, "I", "C", Tonic},
		{Key{0, false}, 1, false, "ii", "Dm", Subdominant},
		{Key{0, false}, 4, true, "V7", "G7", Dominant},
		{Key{0, false}, 6, false, "vii°", "Bdim", Dominant},
		{Key{7, false}, 5, false, "vi", "Em", Tonic},
		{Key{9, true}, 0, false, "i", "Am", Tonic},
		{Key{9, true}, 4, false, "V", "E", Dominant},
		{Key{9, true}, 1, true, "iiø7", "Bm7b5", Subdominant},
		{Key{9, true}, 6, false, "vii°", "G#dim", Dominant},
		{Key{0, false}, 7, false, "I", "C", Tonic},  // degree wraps
		{Key{12, false}, 0, false, "I", "C", Tonic}, // tonic wraps
	}
	for _, tt := range tests {
		c := Diatonic(tt.key, tt.degree, tt.seventh)
		if c.Numeral != tt.numeral || c.Name() != tt.name || c.Function != tt.function {
			t.Errorf("%v degree %d: got %s %s %s, want %s %s %s", tt.key, tt.degree, c.Numeral, c.Name(), c.Function, tt.numeral, tt.name, tt.function)
		}
	}
}

func TestLeadingToneAndSeventh(t *testing.T) {
	v7 := Diatonic(Key{0, false}, 4, true)
	if lt, ok := v7.LeadingTone(); !ok || lt != 11 {
		t.Errorf("V7 leading tone = %d %v", lt, ok)
	}
	if s, ok := v7.Seventh(); !ok || s != 5 {
		t.Errorf("V7 seventh = %d %v", s, ok)
	}
	if lt, ok := Diatonic(Key{0, false}, 6, false).LeadingTone(); !ok || lt != 11 {
		t.Errorf("vii° leading tone = %d %v", lt, ok)
	}
	if _, ok := Diatonic(Key{0, false}, 0, false).LeadingTone(); ok {
		t.Error("I has no leading tone")
	}
	sec, _ := SecondaryDominant(Key{0, false}, 1, false) // A major, leading tone is the C#
	if lt, ok := sec.LeadingTone(); !ok || lt != 1 {
		t.Errorf("V/ii leading tone = %d %v", lt, ok)
	}
}

func TestSecondaryDominant(t *testing.T) {
	c, ok := SecondaryDominant(Key{0, false}, 4, true)
	if !ok || c.Numeral != "V7/V" || c.Name() != "D7" || c.Of != "V" {
		t.Errorf("V7/V = %v %v", c, ok)
	}
	for _, d := range []int{0, 3, 6} {
		if _, ok := SecondaryDominant(Key{0, false}, d, false); ok {
			t.Errorf("degree %d cannot be tonicised", d)
		}
	}
	if _, ok := SecondaryDominant(Key{9, true}, 1, false); ok {
		t.Error("ii° cannot be tonicised")
	}
	if c, ok := SecondaryDominant(Key{9, true}, 3, false); !ok || c.Name() != "A" {
		t.Errorf("V/iv in A minor = %v", c)
	}
}

func TestBorrowed(t *testing.T) {
	c, ok := Borrowed(Key{0, false}, 3)
	if !ok || c.Name() != "Fm" || !c.Borrowed {
		t.Errorf("iv = %v", c)
	}
	if c, _ := Borrowed(Key{0, false}, 5); c.Name() != "G#" {
		t.Errorf("bVI = %v", c.Name())
	}
	if _, ok := Borrowed(Key{0, true}, 3); ok {
		t.Error("no borrowing in minor")
	}
	if _, ok := Borrowed(Key{0, false}, 1); ok {
		t.Error("only iv and bVI")
	}
}

func TestModulate(t *testing.T) {
	got := numerals(Modulate(Key{7, false}, Key{2, false})) // G to D
	if want := []string{"vi=ii", "V7"}; !reflect.DeepEqual(got, want) {
		t.Errorf("G to D = %v, want %v", got, want)
	}
	got = numerals(Modulate(Key{2, false}, Key{7, false})) // D to G: no pivot
	if want := []string{"V7"}; !reflect.DeepEqual(got, want) {
		t.Errorf("D to G = %v, want %v", got, want)
	}
	rel := Modulate(Key{7, false}, Key{7, false}.Relative()) // G to E minor
	if last := rel[len(rel)-1]; last.Name() != "B7" || last.Key.Minor != true {
		t.Errorf("G to Em ends on %v", last)
	}
	for _, c := range Modulate(Key{7, false}, Key{2, false}) {
		if c.Key != (Key{2, false}) {
			t.Errorf("%v not in the new key", c)
		}
	}
}

func TestKey(t *testing.T) {
	if k := (Key{7, false}); k.Dominant() != (Key{2, false}) || k.Relative() != (Key{4, true}) || k.String() != "G major" {
		t.Errorf("key helpers wrong: %v %v %v", k.Dominant(), k.Relative(), k)
	}
	if k := (Key{9, true}); k.Relative() != (Key{0, false}) || k.String() != "A minor" {
		t.Errorf("minor key helpers wrong: %v %v", k.Relative(), k)
	}
}

func TestNewPanics(t *testing.T) {
	for name, cfg := range map[string]Config{
		"length 1":     {PhraseLength: 1},
		"length -3":    {PhraseLength: -3},
		"probability":  {SeventhProb: 1.5},
		"cadence":      {Cadences: map[Cadence]float64{Half: -1}},
		"move weights": {Moves: map[Function]map[Function]float64{Tonic: {Dominant: -1}}},
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: expected panic", name)
				}
			}()
			New(cfg)
		}()
	}
}

func TestPhraseCadences(t *testing.T) {
	for _, minor := range []bool{false, true} {
		for _, n := range []int{2, 3, 4, 6, 9} {
			for c := Authentic; c <= Deceptive; c++ {
				for seed := uint64(0); seed < 20; seed++ {
					genny.Seed(seed)
					g := New(Config{Key: Key{7, minor}, SeventhProb: 0.5, SecondaryProb: 0.5, BorrowedProb: 0.5})
					p := g.Phrase(n, c)
					if len(p) != n {
						t.Fatalf("length %d, want %d", len(p), n)
					}
					last, pre := p[n-1], p[n-2]
					switch c {
					case Authentic:
						if pre.Numeral != "V" && pre.Numeral != "V7" || last.Root != 0 {
							t.Errorf("authentic ends %v %v", pre, last)
						}
					case Half:
						if last.Numeral != "V" {
							t.Errorf("half cadence ends on %v", last)
						}
					case Plagal:
						if last.Root != 0 || pre.Root != 5 {
							t.Errorf("plagal ends %v %v", pre, last)
						}
					case Deceptive:
						if pre.Root != 7 || last.Function != Tonic || last.Root == 0 {
							t.Errorf("deceptive ends %v %v", pre, last)
						}
					}
					if n > 2 && p[0].Numeral != "I" && p[0].Numeral != "i" {
						t.Errorf("phrase starts on %v", p[0])
					}
				}
			}
		}
	}
}

func TestNoRetrogression(t *testing.T) {
	// a dominant is never followed by a subdominant, except a secondary dominant that leads to one
	for seed := uint64(0); seed < 100; seed++ {
		genny.Seed(seed)
		g := New(Config{Key: Key{0, seed%2 == 0}, PhraseLength: 6, SecondaryProb: 0.4, SeventhProb: 0.3})
		prev := g.Generate()
		for range 60 {
			c := g.Generate()
			if prev.Function == Dominant && c.Function == Subdominant && prev.Of == "" {
				t.Fatalf("seed %d: %v -> %v", seed, prev, c)
			}
			if prev.Of != "" && c.Numeral != prev.Of && c.Numeral != prev.Of+"7" {
				// the target of a secondary dominant follows at once (it may be a seventh chord)
				t.Fatalf("seed %d: %v does not resolve to %s but to %v", seed, prev, prev.Of, c)
			}
			prev = c
		}
	}
}

func TestSecondariesAppear(t *testing.T) {
	genny.Seed(2)
	g := New(Config{PhraseLength: 8, SecondaryProb: 1, SeventhProb: 0})
	n := 0
	for range 80 {
		if g.Generate().Of != "" {
			n++
		}
	}
	if n == 0 {
		t.Error("no secondary dominants with SecondaryProb 1")
	}
	genny.Seed(2)
	g = New(Config{PhraseLength: 8})
	for range 80 {
		c := g.Generate()
		if c.Of != "" || c.Borrowed || c.IsSeventh() {
			t.Fatalf("plain config made %v", c)
		}
	}
}

func TestCadenceAlternation(t *testing.T) {
	genny.Seed(1)
	g := New(Config{PhraseLength: 3})
	var ends []string
	for i := 0; i < 12; i++ {
		c := g.Generate()
		if i%3 == 2 {
			ends = append(ends, c.Numeral)
		}
	}
	if want := []string{"V", "I", "V", "I"}; !reflect.DeepEqual(ends, want) {
		t.Errorf("phrase ends %v, want %v", ends, want)
	}
}

func TestCadenceWeights(t *testing.T) {
	genny.Seed(5)
	g := New(Config{PhraseLength: 3, Cadences: map[Cadence]float64{Plagal: 1}})
	for range 5 {
		g.Generate()
		g.Generate()
		if c := g.Generate(); c.Numeral != "I" {
			t.Fatalf("plagal phrase ends on %v", c)
		}
	}
}

func TestMovesOverride(t *testing.T) {
	genny.Seed(3)
	// only tonic and subdominant chords in the body
	g := New(Config{PhraseLength: 6, Moves: map[Function]map[Function]float64{
		Tonic: {Subdominant: 1}, Subdominant: {Tonic: 1},
	}})
	p := g.Phrase(6, Half)
	for i, c := range p[:4] {
		if c.Function == Dominant {
			t.Errorf("chord %d is dominant: %v", i, numerals(p))
		}
	}
}

func TestSeedReproducible(t *testing.T) {
	run := func(seed uint64) []string {
		genny.Seed(seed)
		g := New(Config{Key: Key{7, false}, SecondaryProb: 0.3, SeventhProb: 0.3, BorrowedProb: 0.2})
		var out []string
		for range 30 {
			out = append(out, g.Generate().Numeral)
		}
		return out
	}
	if a, b := run(4), run(4); !reflect.DeepEqual(a, b) {
		t.Errorf("same seed differs:\n%v\n%v", a, b)
	}
	if a, b := run(4), run(5); reflect.DeepEqual(a, b) {
		t.Error("different seeds gave the same chords")
	}
}

func TestGeneratorInterface(t *testing.T) {
	var g genny.Generator[Chord] = New(Config{})
	if !g.Continuous() || g.Done() {
		t.Error("generator is continuous")
	}
	genny.Seed(9)
	a := []Chord{g.Generate(), g.Generate()}
	g.Reset()
	genny.Seed(9)
	b := []Chord{g.Generate(), g.Generate()}
	if !reflect.DeepEqual(a, b) {
		t.Errorf("reset: %v vs %v", a, b)
	}
}

func TestSetKey(t *testing.T) {
	genny.Seed(1)
	g := New(Config{Key: Key{0, false}})
	g.Phrase(4, Authentic)
	g.SetKey(Key{14, false})
	for _, c := range g.Phrase(4, Authentic) {
		if c.Key != (Key{2, false}) {
			t.Errorf("%v not in D", c)
		}
	}
}
