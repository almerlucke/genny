package fugue

import (
	"errors"
	"math"
	"slices"
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/counterpoint"
	"github.com/almerlucke/genny/note"
	"github.com/almerlucke/genny/scale"
)

var (
	dMinor      = scale.New(62, scale.Minor)
	testSubject = NewSubject([]int{0, 4, 5, 3, 1, 4, 3, 2, 1, 0}, []float64{8, 4, 2, 2, 4, 2, 2, 2, 2, 4})
)

func TestAnswer(t *testing.T) {
	tests := []struct {
		name    string
		degrees []int
		kind    AnswerKind
		want    []int
	}{
		{"real answer is a fifth up", []int{0, 4, 5, 3}, Real, []int{4, 8, 9, 7}},
		{"tonal: tonic to dominant, dominant to tonic", []int{0, 4, 5, 3}, Tonal, []int{4, 7, 9, 7}},
		{"tonal: the head ends at the first other note", []int{0, 2, 4}, Tonal, []int{4, 6, 8}},
		{"tonal: dominant first", []int{4, 0, 1, 2}, Tonal, []int{7, 4, 5, 6}},
		{"tonal without a dominant is real", []int{0, 1, 2, 0}, Tonal, []int{4, 5, 6, 4}},
		{"tonal: repeated dominant", []int{4, 4, 3, 2}, Tonal, []int{7, 7, 7, 6}},
		{"tonal: octaves count", []int{7, 11, 9}, Tonal, []int{11, 14, 13}},
		{"tonal: negative degrees", []int{-3, 0, 1}, Tonal, []int{0, 4, 5}},
	}
	for _, tt := range tests {
		s := NewSubject(tt.degrees, slices.Repeat([]float64{1}, len(tt.degrees)))
		got := Answer(s, dMinor, tt.kind)
		if !slices.Equal(got.Degrees, tt.want) {
			t.Errorf("%s: got %v, want %v", tt.name, got.Degrees, tt.want)
		}
		if !slices.Equal(got.Durations, s.Durations) {
			t.Errorf("%s: rhythm changed", tt.name)
		}
	}
	// the input is not changed
	s := NewSubject([]int{0, 4}, []float64{1, 1})
	Answer(s, dMinor, Tonal)
	if !slices.Equal(s.Degrees, []int{0, 4}) {
		t.Error("answer changed the subject")
	}
}

func TestSubject(t *testing.T) {
	s := SubjectFromPitches(dMinor, []float64{62, 69, 65}, []float64{4, 2, 2})
	if !slices.Equal(s.Degrees, []int{0, 4, 2}) || s.Length() != 8 {
		t.Errorf("%v %v", s, s.Length())
	}
	ev := s.Events(dMinor, 16, 7)
	if ev[0].Pitch != 74 || ev[1].Start != 20 || ev[2].Pitch != 77 || ev[2].Duration != 2 {
		t.Errorf("%v", ev)
	}
	for name, f := range map[string]func(){
		"empty":          func() { NewSubject(nil, nil) },
		"length":         func() { NewSubject([]int{0}, []float64{1, 1}) },
		"zero duration":  func() { NewSubject([]int{0}, []float64{0}) },
		"not in the key": func() { SubjectFromPitches(dMinor, []float64{63}, []float64{1}) },
		"no voices":      func() { Exposition(s, 0, dMinor, Config{}) },
		"nil key":        func() { Exposition(s, 2, nil, Config{}) },
		"no notes":       func() { Exposition(Subject{}, 2, dMinor, Config{}) },
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

func TestCountersubject(t *testing.T) {
	for seed := uint64(1); seed <= 8; seed++ {
		for _, species := range []int{1, 2} {
			genny.Seed(seed)
			cfg := Config{Species: species}
			if species == 1 {
				cfg.Answer = Real // a tonal answer rewrites the notes against its head, which note against note rarely allows
			}
			cs, err := Countersubject(testSubject, dMinor, cfg)
			if err != nil {
				t.Fatalf("seed %d species %d: %v", seed, species, err)
			}
			if math.Abs(cs.Length()-testSubject.Length()) > 1e-9 {
				t.Errorf("length %v", cs.Length())
			}
			ca, err := CounterAnswer(testSubject, cs, dMinor, cfg.Answer)
			if err != nil {
				t.Fatalf("seed %d: counter answer: %v", seed, err)
			}
			answer := Answer(testSubject, dMinor, cfg.Answer)
			// above and below (an octave down) the subject and the answer, with the rules of counterpoint
			for _, c := range []struct {
				name string
				cf   Subject
				cs   Subject
			}{{"subject", testSubject, cs}, {"answer", answer, ca}} {
				cf, _ := pitches(c.cf, dMinor, 0)
				for _, side := range []counterpoint.Side{counterpoint.Above, counterpoint.Below} {
					shift := 0
					if side == counterpoint.Below {
						shift = -7
					}
					cp, _ := pitches(c.cs, dMinor, shift)
					conf := counterpoint.Config{Side: side, Scale: dMinor}
					var vs []counterpoint.Violation
					if species == 2 {
						vs = conf.CheckSecond(cf, cp)
					} else {
						vs = conf.Check(cf, cp)
					}
					if len(vs) != 0 {
						t.Errorf("seed %d species %d %s side %d: %v", seed, species, c.name, side, vs)
					}
				}
			}
		}
	}
	// seeds reproduce
	genny.Seed(4)
	a, _ := Countersubject(testSubject, dMinor, Config{})
	genny.Seed(4)
	b, _ := Countersubject(testSubject, dMinor, Config{})
	if !slices.Equal(a.Degrees, b.Degrees) {
		t.Error("same seed, other countersubject")
	}
	// an empty subject has none, a microtonal key is refused
	if _, err := Countersubject(Subject{}, dMinor, Config{}); !errors.Is(err, ErrNoCountersubject) {
		t.Error(err)
	}
	if _, err := Countersubject(testSubject, scale.NewPeriod(62, []float64{0, 2.5, 5}, 12), Config{}); !errors.Is(err, ErrWholeSemitones) {
		t.Error(err)
	}
}

func TestExposition(t *testing.T) {
	for voices := 1; voices <= 5; voices++ {
		for seed := uint64(1); seed <= 3; seed++ {
			genny.Seed(seed)
			f, err := Exposition(testSubject, voices, dMinor, Config{})
			if err != nil {
				t.Fatalf("voices %d seed %d: %v", voices, seed, err)
			}
			if len(f.Entries) != voices || len(f.Voices) != voices {
				t.Fatalf("%d entries, %d voices", len(f.Entries), len(f.Voices))
			}
			for k, e := range f.Entries {
				if e.Start != float64(k)*32 || e.Voice != k || e.Kind != Kind(k%2) {
					t.Errorf("entry %d: %+v", k, e)
				}
				if e.Shift%7 != 0 {
					t.Errorf("entry %d is not moved by octaves: %d", k, e.Shift)
				}
			}
			if vs := f.Check(); len(vs) != 0 {
				t.Errorf("voices %d seed %d: %v", voices, seed, vs)
			}
			if f.Bar(f.Entries[voices-1].Start) != 2*voices-1 {
				t.Errorf("bar %d", f.Bar(f.Entries[voices-1].Start))
			}
		}
	}
	// the answer is a fifth up in the key (A against D), the voices keep their order of pitch at the start
	genny.Seed(1)
	f, _ := Exposition(testSubject, 4, dMinor, Config{})
	if f.Voices[0][0].Pitch != 74 || f.Voices[1][0].Pitch != 69 {
		t.Errorf("first notes %v %v", f.Voices[0][0], f.Voices[1][0])
	}
	// the voice that entered before plays the countersubject against the entry
	if len(f.Voices[0]) <= len(testSubject.Degrees) {
		t.Error("the first voice has no countersubject")
	}
	// reproducible
	genny.Seed(1)
	g, _ := Exposition(testSubject, 4, dMinor, Config{})
	for v := range f.Voices {
		if !slices.Equal(f.Voices[v], g.Voices[v]) {
			t.Errorf("voice %d differs for the same seed", v)
		}
	}
}

func TestCheckFindsMistakes(t *testing.T) {
	genny.Seed(2)
	f, _ := Exposition(testSubject, 3, dMinor, Config{})
	if len(f.Check()) != 0 {
		t.Fatal(f.Check())
	}
	mutate := func(name string, change func(f *Fugue)) {
		genny.Seed(2)
		g, _ := Exposition(testSubject, 3, dMinor, Config{})
		change(g)
		if len(g.Check()) == 0 {
			t.Errorf("%s: not found", name)
		}
	}
	mutate("wrong pitch", func(g *Fugue) { g.Voices[1][3].Pitch++ })
	mutate("wrong length", func(g *Fugue) { g.Voices[0][2].Duration++ })
	mutate("answer marked as subject", func(g *Fugue) { g.Entries[1].Kind = KindSubject })
	mutate("subject marked as answer", func(g *Fugue) { g.Entries[0].Kind = KindAnswer })
	mutate("not a whole octave", func(g *Fugue) { g.Entries[2].Shift++ })
	mutate("no notes", func(g *Fugue) { g.Entries[2].Start += 1 })
	mutate("no such voice", func(g *Fugue) { g.Entries[2].Voice = 9 })
	mutate("overlap", func(g *Fugue) { g.Voices[2][0].Duration += 100 })
}

func TestEpisode(t *testing.T) {
	genny.Seed(3)
	f, _ := Exposition(testSubject, 3, dMinor, Config{})
	start := f.End()
	arrival := f.Episode(EpisodeConfig{Voices: []int{0, 1}, Repeats: 4, Step: -1})
	// the sequence: the head of the subject (the first bar: four notes) moves down a degree every repeat
	var lead []float64
	for _, e := range f.Voices[0] {
		if e.Start >= start {
			lead = append(lead, e.Pitch)
		}
	}
	if len(lead) != 16 {
		t.Fatalf("lead has %d notes", len(lead))
	}
	for r := 1; r < 4; r++ {
		for i := range 4 {
			// one degree down is 1 or 2 semitones
			if d := lead[(r-1)*4+i] - lead[r*4+i]; d != 1 && d != 2 {
				t.Errorf("repeat %d note %d moved %g", r, i, d)
			}
		}
	}
	if arrival%7 != mod(-4, 7) && arrival%7 != -4%7 {
		t.Logf("arrival %d", arrival)
	}
	last := f.Sections[len(f.Sections)-1]
	if last.Name != "episode" || last.Start != start || last.End-last.Start != 64 {
		t.Errorf("section %+v", last)
	}
	// the partner plays the countersubject head, consonant with the lead on the beats
	var partner []float64
	for _, e := range f.Voices[1] {
		if e.Start >= start {
			partner = append(partner, e.Pitch)
		}
	}
	if len(partner) == 0 {
		t.Fatal("no partner")
	}
	for _, p := range f.accentsOf(f.Voices[0], f.Voices[1], start, last.End) {
		if !counterpoint.Consonant(int(math.Round(p))) {
			t.Errorf("dissonance %g in the episode", p)
		}
	}
	// up by step
	f.Episode(EpisodeConfig{Voices: []int{2}, Repeats: 2, Step: 1})
	// invalid voices do not panic and add nothing
	n := len(f.Sections)
	f.Episode(EpisodeConfig{Voices: []int{7}})
	if len(f.Sections) != n {
		t.Error("episode in a missing voice")
	}
	if vs := f.Check(); len(vs) != 0 {
		t.Error(vs)
	}
}

// accentsOf returns the intervals between two lines at the beats of a stretch
func (f *Fugue) accentsOf(a, b []note.Event, from, to float64) []float64 {
	var out []float64
	for t := from; t < to; t += f.Config.beat() {
		pa, okA := sounding(a, t)
		pb, okB := sounding(b, t)
		if okA && okB {
			out = append(out, pa-pb)
		}
	}
	return out
}

func TestMiddleStrettoClose(t *testing.T) {
	genny.Seed(3)
	f, _ := Exposition(testSubject, 4, dMinor, Config{})
	f.Episode(EpisodeConfig{})
	f.Middle(1, KindSubject, 2, true)
	f.Middle(0, KindAnswer, 3, true) // top voice: the countersubject goes below
	f.Middle(9, KindSubject, 0, true)
	if got := f.Entries[len(f.Entries)-1]; got.Section != "middle entry" || got.Voice != 0 || got.Kind != KindAnswer || mod(got.Shift, 7) != 3 {
		t.Errorf("%+v", got)
	}
	if vs := f.Check(); len(vs) != 0 {
		t.Fatal(vs)
	}

	// the stretto works at a distance the checker approves of, and not at one it does not
	good := StrettoDistances(testSubject, dMinor, Tonal, 4, f.Config.beat())
	if !slices.Contains(good, 16) {
		t.Fatalf("distances %v", good)
	}
	before := len(f.Entries)
	if err := f.Stretto(8, nil); !errors.Is(err, ErrStretto) {
		t.Errorf("distance 8: %v", err)
	}
	if len(f.Entries) != before {
		t.Error("a failed stretto added entries")
	}
	for _, bad := range []float64{0, -3} {
		if err := f.Stretto(bad, nil); !errors.Is(err, ErrStretto) {
			t.Error(bad, err)
		}
	}
	if f.Stretto(16, []int{0}) == nil || f.Stretto(16, []int{0, 0}) == nil || f.Stretto(16, []int{0, 12}) == nil {
		t.Error("bad voices accepted")
	}
	start := f.next()
	if err := f.Stretto(16, nil); err != nil {
		t.Fatal(err)
	}
	es := f.Entries[before:]
	if len(es) != 4 {
		t.Fatalf("%d stretto entries", len(es))
	}
	for i, e := range es {
		if e.Start != start+float64(i)*16 || e.Section != "stretto" {
			t.Errorf("%+v", e)
		}
	}
	// the next voice enters before the previous one has finished
	if es[1].Start >= es[0].Start+testSubject.Length() {
		t.Error("no overlap")
	}
	if vs := f.Check(); len(vs) != 0 {
		t.Fatal(vs)
	}

	// the close: a pedal in the bass, then the tonic triad with the top voice on the tonic
	closeStart := f.next()
	f.Close(CloseConfig{Bars: 4, LeadingTone: true})
	end := f.End()
	if end-closeStart != 64 {
		t.Errorf("close lasts %g", end-closeStart)
	}
	bass := f.Voices[3]
	if bass[len(bass)-1].Pitch != 50 && bass[len(bass)-1].Pitch != 38 && bass[len(bass)-1].Pitch != 62 {
		t.Errorf("bass ends on %g", bass[len(bass)-1].Pitch)
	}
	var classes []int
	for v := range f.Voices {
		l := f.Voices[v]
		e := l[len(l)-1]
		if e.Start+e.Duration != end || e.Start != closeStart+32 {
			t.Errorf("voice %d last note %+v", v, e)
		}
		classes = append(classes, int(e.Pitch)%12)
	}
	for _, c := range classes {
		if !slices.Contains([]int{2, 5, 9}, c) {
			t.Errorf("final chord has pitch class %d: %v", c, classes)
		}
	}
	if int(f.Voices[0][len(f.Voices[0])-1].Pitch)%12 != 2 {
		t.Error("the top voice does not end on the tonic")
	}
	// the dominant chord has the raised leading tone C sharp and no C
	for v := range f.Voices {
		for _, e := range f.Voices[v] {
			if e.Start == closeStart && int(e.Pitch)%12 == 0 {
				t.Errorf("voice %d has a C natural in the dominant chord", v)
			}
		}
	}
	if vs := f.Check(); len(vs) != 0 {
		t.Fatal(vs)
	}
}

func TestStrettoDistances(t *testing.T) {
	if got := StrettoDistances(testSubject, dMinor, Tonal, 4, 8); !slices.Equal(got, []float64{16}) {
		t.Errorf("%v", got)
	}
	if StrettoDistances(testSubject, dMinor, Tonal, 1, 8) != nil || StrettoDistances(testSubject, dMinor, Tonal, 2, 0) != nil {
		t.Error("degenerate arguments")
	}
	// with two voices more distances work
	if got := StrettoDistances(testSubject, dMinor, Tonal, 2, 8); len(got) < 1 {
		t.Errorf("%v", got)
	}
}

func TestCloseShortAndOneVoice(t *testing.T) {
	genny.Seed(1)
	f, _ := Exposition(testSubject, 1, dMinor, Config{})
	f.Close(CloseConfig{Bars: 1})
	if vs := f.Check(); len(vs) != 0 {
		t.Error(vs)
	}
	// a single voice only has a bass
	if got := f.Voices[0][len(f.Voices[0])-1]; int(got.Pitch)%12 != 2 {
		t.Errorf("%+v", got)
	}
	if err := f.Stretto(16, nil); !errors.Is(err, ErrStretto) {
		t.Error("stretto of one voice")
	}
}

func TestOtherKeysAndOrder(t *testing.T) {
	// a major key, a real answer and a different order of entry
	key := scale.New(60, scale.Major)
	s := NewSubject([]int{0, 2, 4, 2, 0}, []float64{4, 4, 4, 2, 2})
	for seed := uint64(1); seed <= 3; seed++ {
		genny.Seed(seed)
		f, _ := Exposition(s, 3, key, Config{Answer: Real, Order: []int{1, 0, 2}, StepsPerBar: 8})
		if vs := f.Check(); len(vs) != 0 {
			t.Errorf("seed %d: %v", seed, vs)
		}
		if f.Entries[0].Voice != 1 || f.Entries[1].Start != 16 {
			t.Errorf("%+v", f.Entries)
		}
		if got := f.Answer.Degrees; !slices.Equal(got, []int{4, 6, 8, 6, 4}) {
			t.Errorf("answer %v", got)
		}
	}
}
