// Package counterpoint writes counterpoint to a cantus firmus (a given line of pitches) with the rules of
// the strict style, using backtracking with random choices from the genny stream, so the same seed gives
// the same line.
//
// First species (Generate, Check) is note against note. Second species (GenerateSecond, CheckSecond) is two
// notes against one; its rules are listed below and at GenerateSecond. It is a faithful but not a complete
// version of Fux's second species (see the limitations there).
//
// Pitches are integers in semitones (MIDI note numbers).
//
// The rules of first species, as checked by Check:
//   - consonant intervals only: unison, minor and major third, fifth, minor and major sixth, octave and
//     their compound forms (intervals of 0, 3, 4, 7, 8 or 9 semitones modulo 12)
//   - the first note is a perfect consonance (unison, fifth or octave), the last note a unison or octave
//   - no unison in the middle of the line, and the line stays on its side of the cantus (above or below)
//   - no parallel or direct (hidden) fifths and octaves: no similar motion into a perfect consonance
//   - no leap larger than a major sixth (9 semitones), and no more than two leaps in a row (a leap is
//     more than a whole tone)
//   - no note repeated more than twice in a row
//   - one climax: the highest note of the line occurs once
//
// Second species (CheckSecond) adds to these, for the notes on the beat (strong) and off the beat (weak):
//   - strong notes obey the first species rules (consonant, perfect first and last note, no parallel or
//     direct perfect intervals between successive beats, and no fifth or octave on two successive beats
//     even in contrary motion)
//   - a weak note is consonant or a dissonant passing tone (reached and left by step in the same
//     direction) or neighbour tone (reached and left by step, back to the note it came from)
//   - the last note is approached by step from the weak note before it (the penultimate approach)
//
// Contrary and oblique motion are not rules but preferences: the generator tries those first
package counterpoint

import (
	"errors"
	"fmt"
	"slices"
	"sort"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/note"
	"github.com/almerlucke/genny/scale"
)

// Side tells on which side of the cantus firmus the counterpoint is written
type Side int

const (
	// Above writes the counterpoint above the cantus firmus
	Above Side = iota
	// Below writes it below
	Below
)

// ErrNoSolution is returned by Generate and GenerateSecond if no line satisfies all rules. The line that is
// returned with it is the best effort: it is consonant and starts and ends well, but breaks other
// rules (see Check)
var ErrNoSolution = errors.New("counterpoint: no line satisfies all rules")

// Config is the setup of the generator. The zero value is counterpoint above, with a maximal leap of a
// major sixth, a distance of at most a compound fifth (19 semitones) to the cantus, no range limits and the
// pitch classes of the cantus as its scale
type Config struct {
	// Side of the counterpoint
	Side Side
	// Scale limits the pitches of the counterpoint to a scale (in any octave). With nil it uses the pitch
	// classes that occur in the cantus firmus, which is right for a modal cantus
	Scale *scale.Scale
	// MaxLeap is the largest leap in semitones (default 9)
	MaxLeap int
	// MaxDistance is the largest distance to the cantus in semitones (default 19)
	MaxDistance int
	// Low and High limit the pitches of the counterpoint; they only apply if High > Low
	Low, High int
	// Filter, if set, limits the pitches of the line further: pitch p is only used for note j (an index into
	// the counterpoint) when Filter(j, p) is true. It is for rules about other voices, for instance to stay
	// consonant with a third voice. Check does not know it
	Filter func(j, pitch int) bool
	// Invertible writes a line that can swap places with the cantus by moving one of them an octave: no
	// interval of a fifth anywhere (it would become a dissonant fourth) and a distance of less than
	// an octave, except on the first and last note (so the voices do not cross or meet when one
	// moves). Thirds, sixths and octaves stay consonant
	Invertible bool
	// Budget is the maximal number of search steps before the generator gives up (default 200000)
	Budget int
}

func (c Config) maxLeap() int {
	if c.MaxLeap <= 0 {
		return 9
	}
	return c.MaxLeap
}

func (c Config) maxDistance() int {
	if c.MaxDistance <= 0 {
		return 19
	}
	return c.MaxDistance
}

// Violation is a broken rule at index Index of the counterpoint
type Violation struct {
	Index   int
	Rule    string
	Message string
}

func (v Violation) String() string { return fmt.Sprintf("note %d: %s: %s", v.Index, v.Rule, v.Message) }

// The names of the rules in Violation.Rule
const (
	RuleLength      = "length"
	RuleConsonance  = "consonance"
	RuleFirst       = "first"
	RuleLast        = "last"
	RuleSide        = "side"
	RuleUnison      = "unison"
	RuleDistance    = "distance"
	RuleParallel    = "parallel"
	RuleDirect      = "direct"
	RuleLeap        = "leap"
	RuleLeaps       = "leaps"
	RuleRepeat      = "repeat"
	RuleClimax      = "climax"
	RulePassing     = "passing"
	RulePenultimate = "penultimate"
	RuleInvertible  = "invertible"
)

func mod12(x int) int { return ((x % 12) + 12) % 12 }

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func sign(x int) int {
	switch {
	case x > 0:
		return 1
	case x < 0:
		return -1
	}
	return 0
}

func consonant(interval int) bool {
	switch mod12(interval) {
	case 0, 3, 4, 7, 8, 9:
		return true
	}
	return false
}

// Consonant tells if an interval in semitones (compound intervals and either direction included) is a
// consonance in the strict style: unison, third, fifth, sixth and octave. The fourth is a dissonance
func Consonant(interval int) bool { return consonant(interval) }

func perfect(interval int) bool {
	m := mod12(interval)
	return m == 0 || m == 7
}

// inferSide guesses the side of the counterpoint from the first interval that is not a unison
func inferSide(cf, cp []int) Side {
	for i := range min(len(cf), len(cp)) {
		if cp[i] > cf[i] {
			return Above
		}
		if cp[i] < cf[i] {
			return Below
		}
	}
	return Above
}

// Check returns the violations of the rules of first species in counterpoint cp to cantus firmus cf, in
// order of position (an empty result means the line is correct). The side is taken from the line itself.
// Both must have the same length. It uses the default limits of Config; Config.Check uses your own
func Check(cf, cp []int) []Violation {
	return Config{Side: inferSide(cf, cp)}.Check(cf, cp)
}

// Check is like the function Check but with the side and limits of the config
func (c Config) Check(cf, cp []int) []Violation {
	return c.check(cf, cp, 1)
}

// CheckSecond checks second species: cp has two notes for every note of cf, except the last note of cf which
// has one (2*len(cf)-1 notes). See GenerateSecond for the rules
func CheckSecond(cf, cp []int) []Violation {
	return Config{Side: inferSide2(cf, cp)}.CheckSecond(cf, cp)
}

func inferSide2(cf, cp []int) Side {
	for j := range cp {
		if j/2 < len(cf) {
			if cp[j] > cf[j/2] {
				return Above
			}
			if cp[j] < cf[j/2] {
				return Below
			}
		}
	}
	return Above
}

// CheckSecond is like the function CheckSecond but with the side and limits of the config
func (c Config) CheckSecond(cf, cp []int) []Violation {
	return c.check(cf, cp, 2)
}

func (c Config) check(cf, cp []int, spp int) []Violation {
	want := len(cf) * spp
	if spp == 2 && len(cf) > 0 {
		want--
	}
	if len(cp) != want {
		return []Violation{{0, RuleLength, fmt.Sprintf("%d notes against a cantus of %d, want %d", len(cp), len(cf), want)}}
	}
	var out []Violation
	for j := range cp {
		out = append(out, c.violationsAt(cf, cp, j, spp)...)
	}
	// one climax
	if len(cp) > 1 {
		top := slices.Max(cp)
		count := 0
		for _, p := range cp {
			if p == top {
				count++
			}
		}
		if count > 1 {
			out = append(out, Violation{slices.Index(cp, top), RuleClimax, fmt.Sprintf("the highest note %d occurs %d times", top, count)})
		}
	}
	slices.SortStableFunc(out, func(a, b Violation) int { return a.Index - b.Index })
	return out
}

// strong tells if slot j is on the beat (always in first species)
func strong(j, spp int) bool { return j%spp == 0 }

// violationsAt returns the rules broken by the note cp[j] given the notes before it. Notes after j are not
// looked at, so the generator can use it while building the line
func (c Config) violationsAt(cf, cp []int, j, spp int) []Violation {
	var out []Violation
	add := func(rule, format string, args ...any) {
		out = append(out, Violation{j, rule, fmt.Sprintf(format, args...)})
	}
	n := len(cf)*spp - (spp - 1)
	if len(cf) == 0 {
		n = 0
	}
	p, f := cp[j], cf[j/spp]
	iv := p - f
	if c.Side == Below {
		iv = f - p
	}

	if iv < 0 {
		add(RuleSide, "the line crosses the cantus (%d against %d)", p, f)
	}
	if abs(iv) > c.maxDistance() {
		add(RuleDistance, "%d semitones from the cantus, at most %d", abs(iv), c.maxDistance())
	}
	if c.High > c.Low && (p < c.Low || p > c.High) {
		add(RuleDistance, "pitch %d outside the range %d..%d", p, c.Low, c.High)
	}

	if c.Invertible {
		if mod12(iv) == 7 {
			add(RuleInvertible, "a fifth becomes a fourth when the voices swap")
		}
		if abs(iv) > 12 {
			add(RuleInvertible, "%d semitones from the cantus, more than an octave", abs(iv))
		}
		if abs(iv) == 12 && j > 0 && j < n-1 {
			add(RuleInvertible, "an octave in the middle of the line becomes a unison when the voices swap")
		}
	}

	st := strong(j, spp)
	cons := consonant(iv)
	switch {
	case j == 0 && !perfect(iv):
		add(RuleFirst, "the first note must be a unison, fifth or octave, got %d semitones", iv)
	case j == n-1 && mod12(iv) != 0:
		add(RuleLast, "the last note must be a unison or octave, got %d semitones", iv)
	case st && !cons:
		add(RuleConsonance, "a dissonance of %d semitones on the beat", iv)
	case !st && !cons:
		// a weak dissonance is only allowed as a passing tone, checked when the next note is there
		if j == n-1 {
			add(RulePassing, "the line ends on a dissonance")
		}
	}
	if j > 0 && j < n-1 && iv == 0 {
		add(RuleUnison, "a unison in the middle of the line")
	}

	// a dissonance on a weak note must have been a passing tone (reached and left by step in one direction)
	// or a neighbour tone (reached and left by step, returning to the note it came from)
	if j >= 2 && !strong(j-1, spp) && !consonant(prevInterval(cf, cp, j-1, spp, c.Side)) {
		a, b := cp[j-1]-cp[j-2], cp[j]-cp[j-1]
		stepwise := abs(a) >= 1 && abs(a) <= 2 && abs(b) >= 1 && abs(b) <= 2
		passing := sign(a) == sign(b)
		neighbour := sign(a) == -sign(b) && p == cp[j-2]
		if !stepwise || !(passing || neighbour) {
			add(RulePassing, "the dissonance on note %d is not a passing or neighbour tone", j-1)
		}
	}

	// the penultimate approach: the last note is reached by step
	if spp > 1 && j == n-1 && j > 0 && (abs(p-cp[j-1]) < 1 || abs(p-cp[j-1]) > 2) {
		add(RulePenultimate, "the last note is approached by a step, not %d semitones", abs(p-cp[j-1]))
	}

	if j > 0 {
		step := p - cp[j-1]
		if abs(step) > c.maxLeap() {
			add(RuleLeap, "a leap of %d semitones, at most %d", abs(step), c.maxLeap())
		}
		if j >= 3 && abs(step) > 2 && abs(cp[j-1]-cp[j-2]) > 2 && abs(cp[j-2]-cp[j-3]) > 2 {
			add(RuleLeaps, "three leaps in a row")
		}
		if j >= 2 && p == cp[j-1] && p == cp[j-2] {
			add(RuleRepeat, "note %d repeated three times", p)
		}
	}

	// motion between this and the previous beat
	if st && j >= spp {
		pj := j - spp
		pf := cf[pj/spp]
		pIv := cp[pj] - pf
		if c.Side == Below {
			pIv = pf - cp[pj]
		}
		dcf, dcp := f-pf, p-cp[pj]
		similar := sign(dcf)*sign(dcp) > 0
		if perfect(iv) && similar {
			if mod12(iv) == mod12(pIv) {
				add(RuleParallel, "parallel %s", perfectName(iv))
			} else {
				add(RuleDirect, "direct motion into a %s", perfectName(iv))
			}
		}
		if spp > 1 && perfect(iv) && perfect(pIv) && mod12(iv) == mod12(pIv) && !similar && dcf != 0 && dcp != 0 {
			add(RuleParallel, "%s on successive beats", perfectName(iv))
		}
	}

	return out
}

func prevInterval(cf, cp []int, j, spp int, side Side) int {
	iv := cp[j] - cf[j/spp]
	if side == Below {
		iv = -iv
	}
	return iv
}

func perfectName(iv int) string {
	if mod12(iv) == 0 {
		return "octave"
	}
	return "fifth"
}

// Generate writes first species counterpoint to the cantus firmus: one note against each note. It
// returns ErrNoSolution with the best effort line when the rules cannot all be met (for instance
// for a very short or very narrow cantus); the line is then consonant on every beat but Check tells
// what is wrong. It returns nil and no error for an empty cantus. It never panics
func Generate(cf []int, cfg Config) ([]int, error) { return generate(cf, cfg, 1) }

// GenerateSecond writes second species counterpoint: two notes against each note of the
// cantus, 2*len(cf)-1 notes in all (the last cantus note gets one). On the beat (every first note of
// a pair) the rules of first species apply: consonance, perfect first and last note, no parallel or
// direct perfect intervals between successive beats (and no fifth or octave on two successive beats at
// all), one climax. The second note of a pair is consonant or a dissonant passing tone (reached and left
// by step in the same direction) or neighbour tone (reached and left by step, returning). The last note
// is approached by step. Leaps, repeated notes and the range are limited as in first species.
//
// Limitations: the line always begins on the beat (no opening rest, as Fux has it), weak notes are not
// checked for parallels with the cantus, and the cantus firmus needs no particular form, so a very
// short or narrow one can fail with ErrNoSolution (see CheckSecond for what is then broken)
func GenerateSecond(cf []int, cfg Config) ([]int, error) { return generate(cf, cfg, 2) }

type candidate struct {
	p    int
	cost float64
}

func generate(cf []int, cfg Config, spp int) ([]int, error) {
	if len(cf) == 0 {
		return nil, nil
	}
	n := len(cf) * spp
	if spp == 2 {
		n--
	}
	allowed := cfg.pitchClasses(cf)
	budget := cfg.Budget
	if budget <= 0 {
		budget = 200000
	}

	cp := make([]int, 0, n)
	maxSoFar := 0
	steps := 0

	// the candidates for slot j, best first with a random touch
	candidates := func(j int, relaxed bool) []candidate {
		f := cf[j/spp]
		lo, hi := f, f+cfg.maxDistance()
		if cfg.Side == Below {
			lo, hi = f-cfg.maxDistance(), f
		}
		var out []candidate
		for p := lo; p <= hi; p++ {
			if !allowed[mod12(p)] || (cfg.Filter != nil && !cfg.Filter(j, p)) {
				continue
			}
			cp = append(cp, p)
			var bad []Violation
			if relaxed {
				for _, v := range cfg.violationsAt(cf, cp, j, spp) {
					if v.Rule == RuleConsonance || v.Rule == RuleFirst || v.Rule == RuleLast || v.Rule == RuleSide || v.Rule == RuleDistance {
						bad = append(bad, v)
					}
				}
			} else {
				bad = cfg.violationsAt(cf, cp, j, spp)
				// a climax once: never meet the highest note so far again (it may only be exceeded)
				if len(bad) == 0 && j > 0 && p == maxSoFar {
					bad = append(bad, Violation{j, RuleClimax, ""})
				}
			}
			cp = cp[:len(cp)-1]
			if len(bad) > 0 {
				continue
			}
			out = append(out, candidate{p, cfg.cost(cf, cp, j, spp, p)})
		}
		genny.Shuffle(len(out), func(a, b int) { out[a], out[b] = out[b], out[a] })
		for i := range out {
			out[i].cost += genny.Float64() * 1.5
		}
		sort.SliceStable(out, func(a, b int) bool { return out[a].cost < out[b].cost })
		return out
	}

	var search func(j int, relaxed bool) bool
	search = func(j int, relaxed bool) bool {
		if j == n {
			return true
		}
		for _, c := range candidates(j, relaxed) {
			if steps++; steps > budget {
				return false
			}
			prevMax := maxSoFar
			cp = append(cp, c.p)
			if j == 0 || c.p > maxSoFar {
				maxSoFar = c.p
			}
			if search(j+1, relaxed) {
				return true
			}
			cp = cp[:len(cp)-1]
			maxSoFar = prevMax
			if steps > budget {
				return false
			}
		}
		return false
	}

	if search(0, false) {
		return slices.Clone(cp), nil
	}

	// best effort: only the beat rules
	cp, steps, maxSoFar = cp[:0], 0, 0
	if search(0, true) {
		return slices.Clone(cp), ErrNoSolution
	}
	return nil, ErrNoSolution
}

// pitchClasses are the allowed pitch classes
func (c Config) pitchClasses(cf []int) [12]bool {
	var out [12]bool
	if c.Scale != nil {
		for i := range c.Scale.Len() {
			out[mod12(int(c.Scale.Degree(i)+0.5))] = true
		}
		return out
	}
	for _, p := range cf {
		out[mod12(p)] = true
	}
	return out
}

// cost orders the candidates: contrary motion best, then oblique, then similar; small leaps are
// better than large ones
func (c Config) cost(cf, cp []int, j, spp, p int) float64 {
	if j == 0 {
		return 0
	}
	cost := 0.0
	dcp := p - cp[j-1]
	if strong(j, spp) {
		dcf := cf[j/spp] - cf[(j-1)/spp]
		switch m := sign(dcf) * sign(dcp); {
		case m < 0:
			cost += 0
		case m == 0:
			cost += 1
		default:
			cost += 2
		}
	}
	cost += float64(abs(dcp)) / 4
	if dcp == 0 {
		cost += 1
	}
	return cost
}

// Line turns pitches into note events of the given length in steps, one after the other
func Line(pitches []int, length float64) []note.Event {
	out := make([]note.Event, len(pitches))
	for i, p := range pitches {
		out[i] = note.Event{Pitch: float64(p), Velocity: 0.8, Start: float64(i) * length, Duration: length}
	}
	return out
}
