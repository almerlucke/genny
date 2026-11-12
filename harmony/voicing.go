package harmony

import (
	"math"
	"slices"

	"github.com/almerlucke/genny"
)

// Voice is a voice of the four part texture
type Voice int

const (
	Soprano Voice = iota
	Alto
	Tenor
	Bass
)

func (v Voice) String() string { return [...]string{"soprano", "alto", "tenor", "bass"}[v] }

// Range is the lowest and highest MIDI note of a voice
type Range struct{ Low, High int }

// Ranges holds a range per voice, indexed by Voice
type Ranges [4]Range

// DefaultRanges are the usual choral ranges: soprano C4-A5, alto G3-D5, tenor C3-A4, bass E2-D4
var DefaultRanges = Ranges{{60, 81}, {55, 74}, {48, 69}, {40, 62}}

// Voicing is a chord with the MIDI notes of the four voices, indexed by Voice
type Voicing struct {
	Chord Chord
	Notes [4]int
}

// VoiceConfig configures Realize
type VoiceConfig struct {
	// Ranges of the voices (default DefaultRanges)
	Ranges Ranges
	// Variety adds a random cost in [0, Variety) to each possible voicing, drawn from the genny
	// stream, so that different seeds choose among voicings of about the same movement. 0 is always
	// the same smoothest choice
	Variety float64
}

// the costs of what the realisation avoids. The hard rules are never broken unless there is no way
// at all; the soft rules are broken when the alternative is worse, and show up in the violations
const (
	penHard       = 2000.0
	penResolution = 400.0
	penLeap       = 120.0
	penDoubling   = 60.0
	penSecondInv  = 25.0
	penInversion  = 4.0
	penEndNotRoot = 30.0
)

type cand struct {
	notes [4]int
	cost  float64
}

// Realize writes the chords in four voices. It chooses for every chord the voicing that, together
// with its neighbours, moves least and obeys the rules of Check (found by dynamic programming over
// all voicings within the ranges). A rule that cannot be kept at a certain spot is broken there at
// the lowest cost, and the returned violations list what was accepted (the same as Check of the
// result). It panics on a range with Low above High. An empty list gives nil
func Realize(chords []Chord, cfg VoiceConfig) ([]Voicing, []Violation) {
	if len(chords) == 0 {
		return nil, nil
	}
	r := cfg.Ranges
	if r == (Ranges{}) {
		r = DefaultRanges
	}
	for _, x := range r {
		if x.Low > x.High {
			panic("harmony: voice range with Low above High")
		}
	}

	n := len(chords)
	cands := make([][]cand, n)
	for i, c := range chords {
		cands[i] = candidates(c, r)
		if len(cands[i]) == 0 {
			return nil, nil
		}
		for j := range cands[i] {
			cd := &cands[i][j]
			if cfg.Variety > 0 {
				cd.cost += genny.Float64() * cfg.Variety
			}
			if i == 0 || i == n-1 {
				if rootPC := c.PitchClasses()[0]; cd.notes[Bass]%12 != rootPC {
					cd.cost += penEndNotRoot
				}
			}
			if i == 0 {
				cd.cost += 0.3 * math.Abs(float64(cd.notes[Soprano]-67))
			}
		}
	}

	path := solve(chords, cands, true)
	if path == nil {
		path = solve(chords, cands, false)
	}
	out := make([]Voicing, n)
	for i := range out {
		out[i] = Voicing{Chord: chords[i], Notes: cands[i][path[i]].notes}
	}
	return out, Check(out)
}

// solve finds the cheapest path, or nil if hard rules leave none
func solve(chords []Chord, cands [][]cand, hard bool) []int {
	n := len(cands)
	cost := make([][]float64, n)
	from := make([][]int, n)
	cost[0] = make([]float64, len(cands[0]))
	for j, c := range cands[0] {
		cost[0][j] = c.cost
	}
	for i := 1; i < n; i++ {
		cost[i] = make([]float64, len(cands[i]))
		from[i] = make([]int, len(cands[i]))
		for j, cur := range cands[i] {
			best, arg := math.Inf(1), -1
			for k, prev := range cands[i-1] {
				if math.IsInf(cost[i-1][k], 1) {
					continue
				}
				t := stepCost(chords[i-1], prev.notes, chords[i], cur.notes, hard)
				if v := cost[i-1][k] + t; v < best {
					best, arg = v, k
				}
			}
			cost[i][j] = best + cur.cost
			from[i][j] = arg
		}
	}
	best, arg := math.Inf(1), -1
	for j, v := range cost[n-1] {
		if v < best {
			best, arg = v, j
		}
	}
	if arg < 0 {
		return nil
	}
	path := make([]int, n)
	for i := n - 1; i >= 0; i-- {
		path[i] = arg
		if i > 0 {
			arg = from[i][arg]
		}
	}
	return path
}

// candidates are all ways to give the chord tones to the four voices within the ranges, in order
// from bass to soprano, with no more than an octave between the upper voices, all required chord
// tones present and the leading tone not doubled
func candidates(c Chord, r Ranges) []cand {
	pcs := c.PitchClasses()
	lt, hasLT := c.LeadingTone()
	var out []cand
	inChord := func(p int) bool { return slices.Contains(pcs, p%12) }

	var notes [4]int
	for b := r[Bass].Low; b <= r[Bass].High; b++ {
		if !inChord(b) {
			continue
		}
		notes[Bass] = b
		for t := max(r[Tenor].Low, b+1); t <= r[Tenor].High; t++ {
			if !inChord(t) {
				continue
			}
			notes[Tenor] = t
			for a := max(r[Alto].Low, t+1); a <= min(r[Alto].High, t+12); a++ {
				if !inChord(a) {
					continue
				}
				notes[Alto] = a
				for s := max(r[Soprano].Low, a+1); s <= min(r[Soprano].High, a+12); s++ {
					if !inChord(s) {
						continue
					}
					notes[Soprano] = s
					cost, ok := voicingCost(c, pcs, notes, lt, hasLT)
					if ok {
						out = append(out, cand{notes, cost})
					}
				}
			}
		}
	}
	return out
}

func voicingCost(c Chord, pcs []int, notes [4]int, lt int, hasLT bool) (float64, bool) {
	var count [4]int
	for _, nt := range notes {
		count[slices.Index(pcs, nt%12)]++
	}
	if count[0] == 0 || count[1] == 0 {
		return 0, false
	}
	seventh := len(pcs) == 4
	if seventh && count[3] == 0 || !seventh && count[2] == 0 {
		return 0, false
	}
	if hasLT {
		k := 0
		for _, nt := range notes {
			if nt%12 == lt {
				k++
			}
		}
		if k > 1 {
			return 0, false
		}
	}
	cost := 0.0
	doubled := -1
	for i, k := range count {
		if k > 1 {
			doubled = i
		}
	}
	if doubled >= 0 {
		want := 0
		if hasLT && lt == pcs[0] {
			want = 1
		}
		if doubled != want {
			cost += penDoubling
		}
	}
	switch notes[Bass] % 12 {
	case pcs[1]:
		cost += penInversion
	case pcs[2]:
		cost += penSecondInv
	default:
		if seventh && notes[Bass]%12 == pcs[3] {
			cost += penSecondInv
		}
	}
	return cost, true
}

// stepCost is the cost of moving from one voicing to the next: the movement plus the penalty of each
// rule broken. With hard set, parallels and overlaps are not possible (infinite cost)
func stepCost(pc Chord, p [4]int, cc Chord, q [4]int, hard bool) float64 {
	cost := 0.0
	for v := range 4 {
		d := q[v] - p[v]
		if d < 0 {
			d = -d
		}
		if v == int(Bass) {
			cost += 0.5 * float64(d)
			if d > 12 {
				cost += penLeap
			}
		} else {
			cost += float64(d)
			if d > 7 && d != 12 {
				cost += penLeap
			}
		}
	}
	bad := 0.0
	for i := range 4 {
		for j := i + 1; j < 4; j++ {
			ic := (p[i] - p[j]) % 12
			if (ic == 7 || ic == 0) && (q[i]-q[j])%12 == ic && p[i] != q[i] && p[j] != q[j] {
				bad = penHard
			}
		}
		if i < 3 && (q[i] < p[i+1] || q[i+1] > p[i]) {
			bad = penHard
		}
	}
	if bad > 0 {
		if hard {
			return math.Inf(1)
		}
		cost += bad
	}
	if lt, ok := pc.LeadingTone(); ok && slices.Contains(cc.PitchClasses(), (lt+1)%12) {
		for v := 0; v < 3; v++ {
			if p[v]%12 == lt && q[v] != p[v]+1 {
				cost += penResolution
			}
		}
	}
	if sev, ok := pc.Seventh(); ok {
		for v := range 4 {
			if p[v]%12 == sev {
				if d := q[v] - p[v]; d != -1 && d != -2 {
					cost += penResolution
				}
			}
		}
	}
	return cost
}
