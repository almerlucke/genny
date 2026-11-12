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

func TestSecondSpecies(t *testing.T) {
	for seed := uint64(1); seed <= 20; seed++ {
		genny.Seed(seed)
		cp, err := GenerateSecond(dorian, Config{})
		if err != nil {
			t.Errorf("seed %d: %v", seed, err)
			continue
		}
		if len(cp) != 2*len(dorian)-1 {
			t.Fatalf("seed %d: %d notes", seed, len(cp))
		}
		if vs := CheckSecond(dorian, cp); len(vs) != 0 {
			t.Errorf("seed %d: %v", seed, vs)
		}
	}
	if got := rules(CheckSecond([]int{60, 62}, []int{64})); !slices.Contains(got, RuleLength) {
		t.Error("length", got)
	}
	// a passing dissonance on the weak beat is fine: C against 60: 72 (octave) 71 (7th) 69...
	cf := []int{60, 60, 60}
	if got := rules(CheckSecond(cf, []int{72, 71, 69, 71, 72})); slices.Contains(got, RulePassing) || slices.Contains(got, RuleConsonance) {
		t.Errorf("passing tone: %v", got)
	}
	// a leapt-to dissonance is not
	if got := rules(CheckSecond(cf, []int{72, 74, 69, 71, 72})); !slices.Contains(got, RulePassing) {
		t.Errorf("not a passing tone: %v", got)
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
