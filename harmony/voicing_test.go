package harmony

import (
	"fmt"
	"slices"
	"testing"

	"github.com/almerlucke/genny"
)

var cMajor = Key{0, false}

func vo(c Chord, s, a, t, b int) Voicing { return Voicing{Chord: c, Notes: [4]int{s, a, t, b}} }

func rules(vs []Violation) []Rule {
	var out []Rule
	for _, v := range vs {
		if !slices.Contains(out, v.Rule) {
			out = append(out, v.Rule)
		}
	}
	return out
}

func hasRule(vs []Violation, r Rule) bool { return slices.Contains(rules(vs), r) }

func TestCheckValid(t *testing.T) {
	I, V := Diatonic(cMajor, 0, false), Diatonic(cMajor, 4, false)
	good := []Voicing{
		vo(I, 72, 64, 55, 48),
		vo(V, 71, 62, 55, 43),
		vo(I, 72, 64, 55, 48),
	}
	if vs := Check(good); len(vs) != 0 {
		t.Errorf("valid progression has violations: %v", vs)
	}
	if vs := Check(nil); len(vs) != 0 {
		t.Errorf("nil: %v", vs)
	}
	if vs := Check(good[:1]); len(vs) != 0 {
		t.Errorf("single chord: %v", vs)
	}
}

func TestCheckClassicParallelFifths(t *testing.T) {
	I, V := Diatonic(cMajor, 0, false), Diatonic(cMajor, 4, false)
	// tenor and bass both rise a fifth: C-G becomes G-D
	bad := []Voicing{vo(I, 72, 64, 55, 48), vo(V, 71, 67, 62, 55)}
	vs := Check(bad)
	if len(vs) != 1 || vs[0].Rule != RuleParallelFifths || vs[0].Index != 1 || !slices.Equal(vs[0].Voices, []Voice{Tenor, Bass}) {
		t.Errorf("want one parallel fifth between tenor and bass, got %v", vs)
	}
}

func TestCheckInvalid(t *testing.T) {
	I, V := Diatonic(cMajor, 0, false), Diatonic(cMajor, 4, false)
	V7 := Diatonic(cMajor, 4, true)
	base := vo(I, 72, 64, 55, 48)
	tests := []struct {
		name string
		seq  []Voicing
		want Rule
	}{
		{"parallel octaves", []Voicing{base, vo(V, 79, 71, 62, 55)}, RuleParallelOctave},
		{"crossing", []Voicing{vo(I, 72, 55, 64, 48)}, RuleCrossing},
		{"spacing", []Voicing{vo(I, 79, 64, 55, 48)}, RuleSpacing},
		{"range", []Voicing{vo(I, 84, 64, 55, 48)}, RuleRange},
		{"not in chord", []Voicing{vo(I, 73, 64, 55, 48)}, RuleChordTone},
		{"incomplete", []Voicing{vo(I, 72, 67, 55, 48)}, RuleIncomplete},
		{"overlap", []Voicing{base, vo(V, 74, 71, 67, 43)}, RuleOverlap},
		{"leading tone doubled", []Voicing{vo(V, 71, 59, 55, 43)}, RuleLeadingDoubled},
		{"leading tone does not rise", []Voicing{vo(V, 71, 62, 55, 43), vo(I, 67, 64, 55, 48)}, RuleLeadingResolve},
		{"seventh does not fall", []Voicing{vo(V7, 71, 62, 53, 43), vo(I, 72, 64, 55, 48)}, RuleSeventh},
		{"big leap", []Voicing{base, vo(V, 62, 59, 55, 43)}, RuleLeap},
		{"bass leap", []Voicing{vo(I, 72, 64, 55, 48), vo(V, 71, 62, 55, 55)}, RuleLeap}, // 7 is fine, see next
		{"doubled third", []Voicing{vo(I, 76, 67, 60, 52)}, RuleDoubling},
	}
	for _, tt := range tests {
		if tt.name == "bass leap" {
			// the bass leaps two octaves: a violation; a fifth is not
			seq := []Voicing{vo(I, 72, 64, 55, 62), vo(V, 71, 62, 55, 43)}
			if !hasRule(Check(seq), RuleLeap) {
				t.Errorf("bass leap of 19 not found: %v", Check(seq))
			}
			if hasRule(Check(tt.seq), RuleLeap) {
				t.Errorf("bass leap of 7 is allowed")
			}
			continue
		}
		if vs := Check(tt.seq); !hasRule(vs, tt.want) {
			t.Errorf("%s: want rule %q, got %v", tt.name, tt.want, vs)
		}
	}
}

func TestCheckViolationString(t *testing.T) {
	V := Violation{Index: 1, Rule: RuleParallelFifths, Voices: []Voice{Tenor, Bass}, Detail: "x"}
	if got := V.String(); got != "chord 2: parallel fifths [tenor bass]: x" {
		t.Errorf("String = %q", got)
	}
}

func TestRealizeEdge(t *testing.T) {
	if v, e := Realize(nil, VoiceConfig{}); v != nil || e != nil {
		t.Error("empty input gives nil")
	}
	defer func() {
		if recover() == nil {
			t.Error("bad range should panic")
		}
	}()
	Realize([]Chord{Diatonic(cMajor, 0, false)}, VoiceConfig{Ranges: Ranges{{70, 60}, {50, 60}, {40, 50}, {30, 40}}})
}

func TestRealizeNarrowRanges(t *testing.T) {
	// ranges in which no chord can be voiced give nil, never a panic
	r := Ranges{{60, 60}, {60, 60}, {60, 60}, {60, 60}}
	if v, _ := Realize([]Chord{Diatonic(cMajor, 0, false)}, VoiceConfig{Ranges: r}); v != nil {
		t.Errorf("got %v", v)
	}
}

func TestRealizeClassic(t *testing.T) {
	var cs []Chord
	for _, d := range []int{0, 3, 4, 0} {
		cs = append(cs, Diatonic(cMajor, d, false))
	}
	vs, exc := Realize(cs, VoiceConfig{})
	if len(vs) != 4 || len(exc) != 0 {
		t.Fatalf("I IV V I: %d voicings, violations %v", len(vs), exc)
	}
	if vs[0].Notes[Bass]%12 != 0 || vs[3].Notes[Bass]%12 != 0 {
		t.Errorf("starts and ends with the root in the bass: %v %v", vs[0].Notes, vs[3].Notes)
	}
}

func TestRealizeSeventhAndLeadingTone(t *testing.T) {
	cs := []Chord{
		Diatonic(cMajor, 0, false), Diatonic(cMajor, 1, true), Diatonic(cMajor, 4, true), Diatonic(cMajor, 0, false),
	}
	vs, exc := Realize(cs, VoiceConfig{})
	if len(exc) != 0 {
		t.Fatalf("violations: %v\n%v", exc, vs)
	}
	// the leading tone of V7 rises, the seventh falls
	for v := range 3 {
		if vs[2].Notes[v]%12 == 11 && vs[3].Notes[v] != vs[2].Notes[v]+1 {
			t.Errorf("leading tone in voice %d does not rise", v)
		}
	}
	for v := range 4 {
		if vs[2].Notes[v]%12 == 5 && vs[3].Notes[v]-vs[2].Notes[v] != -1 {
			t.Errorf("seventh in voice %d does not fall a semitone", v)
		}
	}
}

func TestRealizeReproducible(t *testing.T) {
	run := func(seed uint64) []Voicing {
		genny.Seed(seed)
		g := New(Config{Key: Key{7, false}, SecondaryProb: 0.3, SeventhProb: 0.3})
		var cs []Chord
		for range 12 {
			cs = append(cs, g.Generate())
		}
		v, _ := Realize(cs, VoiceConfig{Variety: 3})
		return v
	}
	a, b := run(3), run(3)
	for i := range a {
		if a[i].Notes != b[i].Notes {
			t.Fatalf("same seed differs at chord %d", i)
		}
	}
	c := run(4)
	same := true
	for i := range a {
		same = same && a[i].Notes == c[i].Notes
	}
	if same {
		t.Error("different seeds gave the same voicing")
	}
}

// Generated realisations must pass the independent checker. The rules that are never given up
// (ranges, chord tones, completeness, order, spacing, parallels, overlap, leading tone doubled) must
// always hold; the others (resolutions, leaps, doubling) may be exceptions but must be rare
func TestPropertyGeneratedPassCheck(t *testing.T) {
	hard := []Rule{RuleRange, RuleChordTone, RuleIncomplete, RuleCrossing, RuleSpacing,
		RuleParallelFifths, RuleParallelOctave, RuleOverlap, RuleLeadingDoubled}
	chords, exceptions, clean := 0, 0, 0
	const seeds = 300
	for seed := uint64(0); seed < seeds; seed++ {
		genny.Seed(seed)
		key := Key{int(seed % 12), seed%3 == 0}
		g := New(Config{Key: key, PhraseLength: 4 + int(seed%3), SeventhProb: 0.3, SecondaryProb: 0.3, BorrowedProb: 0.2})
		var cs []Chord
		for range 16 {
			cs = append(cs, g.Generate())
		}
		vs, exc := Realize(cs, VoiceConfig{Variety: float64(seed % 4)})
		if len(vs) != len(cs) {
			t.Fatalf("seed %d: %d voicings for %d chords", seed, len(vs), len(cs))
		}
		again := Check(vs)
		if len(again) != len(exc) {
			t.Fatalf("seed %d: Realize reports %d violations, Check %d", seed, len(exc), len(again))
		}
		for _, r := range hard {
			if hasRule(again, r) {
				t.Errorf("seed %d: hard rule %q broken: %v", seed, r, again)
			}
		}
		chords += len(cs)
		exceptions += len(exc)
		if len(exc) == 0 {
			clean++
		}
	}
	t.Logf("%d chords, %d exceptions, %d of %d seeds without any", chords, exceptions, clean, seeds)
	if float64(exceptions) > 0.03*float64(chords) {
		t.Errorf("too many accepted exceptions: %d of %d chords", exceptions, chords)
	}
}

func ExampleRealize() {
	cs := []Chord{
		Diatonic(Key{0, false}, 0, false), Diatonic(Key{0, false}, 3, false),
		Diatonic(Key{0, false}, 4, true), Diatonic(Key{0, false}, 0, false),
	}
	voicings, violations := Realize(cs, VoiceConfig{})
	for _, v := range voicings {
		fmt.Println(v.Chord.Numeral, v.Notes)
	}
	fmt.Println(len(violations), "violations")
	// Output:
	// I [67 60 52 48]
	// IV [69 60 53 41]
	// V7 [67 59 53 43]
	// I [67 60 52 48]
	// 0 violations
}
