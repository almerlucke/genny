package harmony

import (
	"fmt"
	"slices"
)

// Rule names a rule of four part writing
type Rule string

const (
	RuleRange          Rule = "range"            // a note outside the range of its voice
	RuleChordTone      Rule = "chord tone"       // a note that is not in the chord
	RuleIncomplete     Rule = "incomplete"       // root, third (and fifth, or seventh) missing
	RuleCrossing       Rule = "crossing"         // a lower voice above a higher voice
	RuleSpacing        Rule = "spacing"          // more than an octave between soprano-alto or alto-tenor
	RuleParallelFifths Rule = "parallel fifths"  // two voices a fifth apart move to a fifth
	RuleParallelOctave Rule = "parallel octaves" // two voices an octave (or unison) apart move to one
	RuleOverlap        Rule = "overlap"          // a voice moves past where its neighbour was
	RuleLeadingDoubled Rule = "leading tone doubled"
	RuleLeadingResolve Rule = "leading tone" // a leading tone in an upper voice does not rise a semitone
	RuleSeventh        Rule = "seventh"      // a seventh does not fall a step
	RuleLeap           Rule = "leap"         // upper voice over a fifth (octave excepted), bass over an octave
	RuleDoubling       Rule = "doubling"     // in a triad the root is doubled; with the fifth left out of a seventh chord too
)

// Violation is one broken rule. Index is the position in the list of voicings (see Check for the
// rules about moves), Voices the voices involved and Detail says what happened
type Violation struct {
	Index  int
	Rule   Rule
	Voices []Voice
	Detail string
}

func (v Violation) String() string {
	return fmt.Sprintf("chord %d: %s %v: %s", v.Index+1, v.Rule, v.Voices, v.Detail)
}

// Check lists the violations of the rules of four part writing in a list of voicings, using
// DefaultRanges. For a rule about a move (parallels, overlap, resolutions, leaps) Index is the chord
// the voices move to, except for the resolution of leading tones and sevenths, where it is the chord
// that has the note that should have resolved. The rules:
//
//   - every note is in its range and in the chord, the voices keep their order and the upper three
//     are within an octave of each other
//   - triads have root, third and fifth, seventh chords have root, third and seventh
//   - no parallel fifths or octaves between any two voices, and no overlap of neighbours
//   - a leading tone (the third of a major dominant, the root of vii diminished) is not doubled, and
//     in an upper voice rises a semitone when the next chord has the tonic note to rise to
//   - a seventh falls by a step
//   - upper voices leap at most a fifth (or an octave), the bass at most an octave
//   - the doubled note of a triad is the root (the third in vii diminished), and a seventh chord
//     without fifth doubles its root
func Check(v []Voicing) []Violation { return CheckRanges(v, DefaultRanges) }

// CheckRanges is Check with other voice ranges
func CheckRanges(v []Voicing, r Ranges) []Violation {
	var out []Violation
	add := func(i int, rule Rule, d string, voices ...Voice) {
		out = append(out, Violation{Index: i, Rule: rule, Voices: voices, Detail: d})
	}

	for i, x := range v {
		pcs := x.Chord.PitchClasses()
		lt, hasLT := x.Chord.LeadingTone()

		// ranges, chord tones, order, spacing
		pos := make([]int, 4) // chord position of each voice, -1 if not a chord tone
		for k := range 4 {
			n := x.Notes[k]
			if n < r[k].Low || n > r[k].High {
				add(i, RuleRange, fmt.Sprintf("%d outside %d-%d", n, r[k].Low, r[k].High), Voice(k))
			}
			pos[k] = slices.Index(pcs, ((n%12)+12)%12)
			if pos[k] < 0 {
				add(i, RuleChordTone, fmt.Sprintf("%d is not in %s", n, x.Chord.Name()), Voice(k))
			}
		}
		for k := 0; k < 3; k++ {
			if x.Notes[k] < x.Notes[k+1] {
				add(i, RuleCrossing, fmt.Sprintf("%d below %d", x.Notes[k], x.Notes[k+1]), Voice(k), Voice(k+1))
			}
		}
		for k := 0; k < 2; k++ {
			if x.Notes[k]-x.Notes[k+1] > 12 {
				add(i, RuleSpacing, fmt.Sprintf("%d semitones apart", x.Notes[k]-x.Notes[k+1]), Voice(k), Voice(k+1))
			}
		}

		// completeness and doubling (only when all notes are chord tones)
		if !slices.Contains(pos, -1) {
			need := []int{0, 1, 2}
			if x.Chord.IsSeventh() {
				need = []int{0, 1, 3}
			}
			complete := true
			for _, p := range need {
				if !slices.Contains(pos, p) {
					complete = false
					add(i, RuleIncomplete, fmt.Sprintf("no %s", [...]string{"root", "third", "fifth", "seventh"}[p]))
				}
			}
			if complete {
				counts := make([]int, 4)
				for _, p := range pos {
					counts[p]++
				}
				for p, n := range counts {
					if n < 2 || (hasLT && pcs[p] == lt) {
						continue
					}
					want := 0
					if hasLT && lt == pcs[0] {
						want = 1
					}
					if p != want && (!x.Chord.IsSeventh() || counts[2] == 0) {
						add(i, RuleDoubling, fmt.Sprintf("%s is doubled", [...]string{"root", "third", "fifth", "seventh"}[p]))
					}
				}
			}
		}
		if hasLT {
			var who []Voice
			for k := range 4 {
				if ((x.Notes[k]%12)+12)%12 == lt {
					who = append(who, Voice(k))
				}
			}
			if len(who) > 1 {
				add(i, RuleLeadingDoubled, "leading tone is in several voices", who...)
			}
		}

		// resolutions need the next chord
		if i+1 < len(v) {
			next := v[i+1]
			if hasLT && slices.Contains(next.Chord.PitchClasses(), (lt+1)%12) {
				for k := 0; k < 3; k++ {
					if ((x.Notes[k]%12)+12)%12 == lt && next.Notes[k] != x.Notes[k]+1 {
						add(i, RuleLeadingResolve, fmt.Sprintf("%d goes to %d, not up a semitone", x.Notes[k], next.Notes[k]), Voice(k))
					}
				}
			}
			if sev, ok := x.Chord.Seventh(); ok {
				for k := range 4 {
					if ((x.Notes[k]%12)+12)%12 == sev {
						if d := next.Notes[k] - x.Notes[k]; d != -1 && d != -2 {
							add(i, RuleSeventh, fmt.Sprintf("%d goes to %d, not down a step", x.Notes[k], next.Notes[k]), Voice(k))
						}
					}
				}
			}
		}
		if i == 0 {
			continue
		}

		// moves from the previous chord
		p := v[i-1]
		for a := 0; a < 4; a++ {
			for b := a + 1; b < 4; b++ {
				if p.Notes[a] == x.Notes[a] || p.Notes[b] == x.Notes[b] {
					continue // an oblique or static pair is not parallel
				}
				before := ((p.Notes[a]-p.Notes[b])%12 + 12) % 12
				after := ((x.Notes[a]-x.Notes[b])%12 + 12) % 12
				if before != after {
					continue
				}
				switch before {
				case 7:
					add(i, RuleParallelFifths, fmt.Sprintf("%d,%d to %d,%d", p.Notes[a], p.Notes[b], x.Notes[a], x.Notes[b]), Voice(a), Voice(b))
				case 0:
					add(i, RuleParallelOctave, fmt.Sprintf("%d,%d to %d,%d", p.Notes[a], p.Notes[b], x.Notes[a], x.Notes[b]), Voice(a), Voice(b))
				}
			}
		}
		for a := 0; a < 3; a++ {
			if x.Notes[a] < p.Notes[a+1] || x.Notes[a+1] > p.Notes[a] {
				add(i, RuleOverlap, fmt.Sprintf("%d, %d after %d, %d", x.Notes[a], x.Notes[a+1], p.Notes[a], p.Notes[a+1]), Voice(a), Voice(a+1))
			}
		}
		for k := range 4 {
			d := x.Notes[k] - p.Notes[k]
			if d < 0 {
				d = -d
			}
			limit := 7
			if k == int(Bass) {
				limit = 12
			}
			if d > limit && !(k < 3 && d == 12) {
				add(i, RuleLeap, fmt.Sprintf("%d semitones", d), Voice(k))
			}
		}
	}
	return out
}
