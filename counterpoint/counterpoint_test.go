package counterpoint

import (
	"errors"
	"slices"
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/scale"
)

// D Dorian, 12 notes
var dorian = []int{62, 65, 64, 67, 65, 69, 67, 65, 64, 62, 60, 62}

func rules(vs []Violation) []string {
	var out []string
	for _, v := range vs {
		out = append(out, v.Rule)
	}
	return out
}

func TestCheck(t *testing.T) {
	cf := []int{62, 65, 64, 62}
	tests := []struct {
		name string
		cp   []int
		want []string // rules that must appear; nil means none
	}{
		{"good", []int{69, 73, 72, 74}, nil},
		{"wrong length", []int{69, 72}, []string{RuleLength}},
		{"dissonance", []int{69, 71, 71, 74}, []string{RuleConsonance}},
		{"first not perfect", []int{66, 72, 71, 74}, []string{RuleFirst}},
		{"last not octave", []int{69, 72, 71, 69}, []string{RuleLast}},
		{"parallel fifths", []int{69, 72, 71, 74}, []string{RuleParallel}},
		{"unison in the middle", []int{69, 65, 71, 74}, []string{RuleUnison}},
		{"leap too big", []int{69, 80, 71, 74}, []string{RuleLeap}},
		{"crossing", []int{69, 60, 71, 74}, []string{RuleSide}},
	}
	for _, tt := range tests {
		got := rules(Check(cf, tt.cp))
		for _, w := range tt.want {
			if !slices.Contains(got, w) {
				t.Errorf("%s: want rule %q, got %v", tt.name, w, got)
			}
		}
		if tt.want == nil && tt.name == "good" && len(got) != 0 {
			t.Errorf("%s: unexpected %v", tt.name, got)
		}
	}
}

func TestCheckMotion(t *testing.T) {
	// cantus up a step, counterpoint up a step: fifth to fifth is parallel
	cf := []int{60, 62, 60}
	if got := rules(Check(cf, []int{67, 69, 72})); !slices.Contains(got, RuleParallel) {
		t.Errorf("parallel fifths: %v", got)
	}
	// similar motion into an octave: direct
	cf = []int{60, 62, 60, 60}
	if got := rules(Check(cf, []int{64, 74, 76, 72})); !slices.Contains(got, RuleDirect) {
		t.Errorf("direct octave: %v", got)
	}
	// contrary motion into a fifth is fine
	if got := rules(Check([]int{60, 62}, []int{72, 69})); slices.Contains(got, RuleDirect) || slices.Contains(got, RuleParallel) {
		t.Errorf("contrary: %v", got)
	}
	// three leaps in a row, and the same note three times, and two climaxes
	if got := rules(Check([]int{60, 60, 60, 60, 60}, []int{67, 72, 67, 72, 72})); !slices.Contains(got, RuleLeaps) {
		t.Errorf("leaps: %v", got)
	}
	if got := rules(Check([]int{60, 62, 64, 65, 60}, []int{64, 64, 64, 64, 72})); !slices.Contains(got, RuleRepeat) {
		t.Errorf("repeat: %v", got)
	}
	if got := rules(Check([]int{60, 62, 64, 62, 60}, []int{72, 74, 76, 76, 72})); !slices.Contains(got, RuleClimax) {
		t.Errorf("climax: %v", got)
	}
}

func TestGenerateSatisfiesCheck(t *testing.T) {
	for _, side := range []Side{Above, Below} {
		for seed := uint64(1); seed <= 30; seed++ {
			genny.Seed(seed)
			cfg := Config{Side: side}
			cp, err := Generate(dorian, cfg)
			if err != nil {
				t.Errorf("seed %d side %d: %v", seed, side, err)
				continue
			}
			if len(cp) != len(dorian) {
				t.Fatalf("length %d", len(cp))
			}
			if vs := cfg.Check(dorian, cp); len(vs) != 0 {
				t.Errorf("seed %d side %d: %v violations: %v", seed, side, len(vs), vs)
			}
			if vs := Check(dorian, cp); len(vs) != 0 {
				t.Errorf("seed %d side %d: inferred side check: %v", seed, side, vs)
			}
		}
	}
}

func TestGenerateSeeded(t *testing.T) {
	genny.Seed(5)
	a, _ := Generate(dorian, Config{})
	genny.Seed(5)
	b, _ := Generate(dorian, Config{})
	if !slices.Equal(a, b) {
		t.Error("same seed, other line")
	}
	differ := false
	for seed := uint64(6); seed < 12; seed++ {
		genny.Seed(seed)
		c, _ := Generate(dorian, Config{})
		differ = differ || !slices.Equal(a, c)
	}
	if !differ {
		t.Error("seeds do not vary the line")
	}
}

func TestGenerateEdgeCases(t *testing.T) {
	genny.Seed(1)
	if cp, err := Generate(nil, Config{}); cp != nil || err != nil {
		t.Error("empty cantus", cp, err)
	}
	// one note: an octave or a unison
	cp, _ := Generate([]int{60}, Config{})
	if len(cp) != 1 || mod12(cp[0]-60) != 0 {
		t.Error("single note", cp)
	}
	// a cantus that cannot get a line (a constant cantus: the line would repeat itself or reuse its climax)
	cp, err := Generate([]int{60, 60, 60, 60, 60, 60}, Config{MaxDistance: 3})
	if !errors.Is(err, ErrNoSolution) {
		t.Errorf("want ErrNoSolution, got %v (%v)", err, cp)
	}
	// the budget gives up without panic
	cp, err = Generate(dorian, Config{Budget: 1})
	if err == nil {
		t.Error("budget 1 should fail", cp)
	}
	// impossible range
	if cp, err = Generate(dorian, Config{Low: 1, High: 2}); err == nil || cp != nil {
		t.Error("impossible range", cp, err)
	}
}

func TestScaleOption(t *testing.T) {
	// with a scale only its pitch classes are used, also those that are not in the cantus
	genny.Seed(2)
	key := scale.New(62, scale.Dorian)
	cp, err := Generate(dorian, Config{Scale: key})
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range cp {
		if !slices.Contains([]int{2, 4, 5, 7, 9, 11, 0}, mod12(p)) {
			t.Errorf("pitch %d is not in D dorian", p)
		}
	}
}

func TestCheckSecondTable(t *testing.T) {
	cf := []int{62, 65, 64, 62}
	flat := []int{60, 60, 60}
	tests := []struct {
		name string
		cf   []int
		cp   []int
		want []string // the rules that must appear; nil means the line must be correct
	}{
		{"valid with a passing tone", cf, []int{69, 71, 69, 70, 71, 76, 74}, nil},
		{"valid", cf, []int{69, 71, 69, 72, 71, 76, 74}, nil},
		{"valid with neighbour tones", flat, []int{72, 74, 72, 71, 72}, nil},
		{"wrong length", cf, []int{69, 71, 69}, []string{RuleLength}},
		{"dissonance on the beat", cf, []int{69, 71, 70, 72, 71, 76, 74}, []string{RuleConsonance}},
		{"weak dissonance reached by leap", cf, []int{69, 72, 69, 72, 71, 76, 74}, []string{RulePassing}},
		{"weak dissonance left by leap", flat, []int{72, 74, 69, 71, 72}, []string{RulePassing}},
		{"weak dissonance a repeated note", flat, []int{72, 74, 74, 71, 72}, []string{RulePassing}},
		{"first note not perfect", cf, []int{66, 71, 69, 72, 71, 76, 74}, []string{RuleFirst}},
		{"last note not octave", cf, []int{69, 71, 69, 72, 71, 76, 71}, []string{RuleLast}},
		{"last note not reached by step", cf, []int{69, 71, 69, 72, 71, 64, 74}, []string{RulePenultimate}},
		{"unison in the middle", cf, []int{69, 62, 62, 72, 71, 76, 74}, []string{RuleUnison}},
		{"leap too big", cf, []int{69, 80, 69, 72, 71, 76, 74}, []string{RuleLeap}},
		{"crossing the cantus", cf, []int{69, 60, 69, 72, 71, 76, 74}, []string{RuleSide}},
		{"three times the same note", flat, []int{72, 72, 72, 71, 72}, []string{RuleRepeat}},
	}
	for _, tt := range tests {
		got := rules(CheckSecond(tt.cf, tt.cp))
		if tt.want == nil {
			if len(got) != 0 {
				t.Errorf("%s: unexpected %v", tt.name, got)
			}
			continue
		}
		for _, w := range tt.want {
			if !slices.Contains(got, w) {
				t.Errorf("%s: want rule %q, got %v", tt.name, w, got)
			}
		}
	}
}

func TestCheckSecondParallels(t *testing.T) {
	// beats: fifth (67 over 60), fifth (69 over 62) in similar motion: parallel fifths
	cf := []int{60, 62, 60}
	if got := rules(CheckSecond(cf, []int{67, 65, 69, 71, 72})); !slices.Contains(got, RuleParallel) {
		t.Errorf("parallel fifths: %v", got)
	}
	// a fifth followed by a fifth in contrary motion is also refused in second species
	if got := rules(CheckSecond([]int{62, 60, 62}, []int{69, 67, 67, 71, 74})); !slices.Contains(got, RuleParallel) {
		t.Errorf("fifths on successive beats: %v", got)
	}
	// direct octave: the cantus and the line both rise into the octave
	if got := rules(CheckSecond([]int{60, 62, 60}, []int{64, 67, 74, 73, 72})); !slices.Contains(got, RuleDirect) {
		t.Errorf("direct octave: %v", got)
	}
}

func TestSecondSpecies(t *testing.T) {
	key := scale.New(62, scale.Dorian)
	cantus := [][]int{
		dorian,
		{62, 65, 64, 62},
		{62, 67, 65, 64, 62},
		{57, 60, 59, 57, 55, 57},
		{60, 62, 64, 65, 67, 65, 64, 62, 60},
	}
	for ci, cf := range cantus {
		for _, side := range []Side{Above, Below} {
			for seed := uint64(1); seed <= 60; seed++ {
				genny.Seed(seed)
				cfg := Config{Side: side, Scale: key}
				if ci == 4 {
					cfg.Scale = scale.New(60, scale.Major)
				}
				cp, err := GenerateSecond(cf, cfg)
				if err != nil {
					t.Errorf("cantus %d side %d seed %d: %v", ci, side, seed, err)
					continue
				}
				if len(cp) != 2*len(cf)-1 {
					t.Fatalf("seed %d: %d notes", seed, len(cp))
				}
				if vs := cfg.CheckSecond(cf, cp); len(vs) != 0 {
					t.Errorf("cantus %d side %d seed %d: %v", ci, side, seed, vs)
				}
				if vs := CheckSecond(cf, cp); len(vs) != 0 {
					t.Errorf("cantus %d side %d seed %d (inferred side): %v", ci, side, seed, vs)
				}
			}
		}
	}
	// without a scale the pitch classes of the cantus are used
	for seed := uint64(1); seed <= 20; seed++ {
		genny.Seed(seed)
		cp, err := GenerateSecond(dorian, Config{})
		if err != nil || len(CheckSecond(dorian, cp)) != 0 {
			t.Errorf("seed %d: %v %v", seed, err, CheckSecond(dorian, cp))
		}
	}
}

func TestSecondSpeciesSeededAndEdges(t *testing.T) {
	genny.Seed(4)
	a, _ := GenerateSecond(dorian, Config{})
	genny.Seed(4)
	b, _ := GenerateSecond(dorian, Config{})
	if !slices.Equal(a, b) {
		t.Error("same seed, other line")
	}
	if cp, err := GenerateSecond(nil, Config{}); cp != nil || err != nil {
		t.Error("empty cantus", cp, err)
	}
	cp, err := GenerateSecond([]int{60}, Config{})
	if err != nil || len(cp) != 1 || mod12(cp[0]-60) != 0 {
		t.Error("single note", cp, err)
	}
	// impossible: no solution, no panic
	if _, err := GenerateSecond(dorian, Config{Low: 1, High: 2}); !errors.Is(err, ErrNoSolution) {
		t.Error("impossible range", err)
	}
	if _, err := GenerateSecond(dorian, Config{Budget: 1}); err == nil {
		t.Error("budget 1 should fail")
	}
}

func TestLineAndViolationString(t *testing.T) {
	l := Line([]int{60, 62}, 4)
	if len(l) != 2 || l[1].Start != 4 || l[1].Duration != 4 {
		t.Error(l)
	}
	if (Violation{3, RuleLeap, "x"}).String() == "" {
		t.Error("string")
	}
}

func TestInvertible(t *testing.T) {
	key := scale.New(62, scale.Dorian)
	cf := []int{62, 65, 64, 62}
	for species := 1; species <= 2; species++ {
		for seed := uint64(1); seed <= 50; seed++ {
			genny.Seed(seed)
			cfg := Config{Scale: key, Invertible: true}
			gen, check := Generate, cfg.Check
			if species == 2 {
				gen, check = GenerateSecond, cfg.CheckSecond
			}
			cp, err := gen(cf, cfg)
			if err != nil {
				t.Errorf("species %d seed %d: %v", species, seed, err)
				continue
			}
			if vs := check(cf, cp); len(vs) != 0 {
				t.Errorf("species %d seed %d: %v", species, seed, vs)
			}
			for j, p := range cp {
				if mod12(p-cf[j/species]) == 7 {
					t.Errorf("fifth in %v", cp)
				}
			}
			// moved an octave down the line is below the cantus and still obeys the rules
			down := make([]int, len(cp))
			for i, p := range cp {
				down[i] = p - 12
			}
			if vs := (Config{Side: Below, Scale: key}).Check(cf, down); species == 1 && len(vs) != 0 {
				// only the climax and unison rules can differ for a line that touches the cantus
				for _, v := range vs {
					if v.Rule != RuleUnison && v.Rule != RuleSide {
						t.Errorf("seed %d inverted: %v", seed, v)
					}
				}
			}
		}
	}
	// a fifth is flagged
	if got := rules(Config{Invertible: true}.Check([]int{60, 62}, []int{67, 69})); !slices.Contains(got, RuleInvertible) {
		t.Error(got)
	}
	if !Consonant(16) || Consonant(5) {
		t.Error("Consonant")
	}
}
