// Package chord builds chords (lists of pitches in semitones), inversions and voicings, and
// chooses voicings that move as little as possible between chords
package chord

import (
	"math"
	"slices"

	"github.com/almerlucke/genny/scale"
)

// Intervals of common chords, in semitones above the root
var (
	Major          = []float64{0, 4, 7}
	Minor          = []float64{0, 3, 7}
	Diminished     = []float64{0, 3, 6}
	Augmented      = []float64{0, 4, 8}
	Sus2           = []float64{0, 2, 7}
	Sus4           = []float64{0, 5, 7}
	Major6         = []float64{0, 4, 7, 9}
	Minor6         = []float64{0, 3, 7, 9}
	Major7         = []float64{0, 4, 7, 11}
	Minor7         = []float64{0, 3, 7, 10}
	Dominant7      = []float64{0, 4, 7, 10}
	MinorMajor7    = []float64{0, 3, 7, 11}
	HalfDiminished = []float64{0, 3, 6, 10}
	Diminished7    = []float64{0, 3, 6, 9}
	Add9           = []float64{0, 4, 7, 14}
	Major9         = []float64{0, 4, 7, 11, 14}
	Minor9         = []float64{0, 3, 7, 10, 14}
	Dominant9      = []float64{0, 4, 7, 10, 14}
)

// Chord is a list of pitches in ascending order
type Chord []float64

// New creates a chord from a root and intervals above it
func New(root float64, intervals []float64) Chord {
	c := make(Chord, len(intervals))
	for i, iv := range intervals {
		c[i] = root + iv
	}
	slices.Sort(c)
	return c
}

// Diatonic builds the chord on a scale degree by stacking thirds of the scale (every second
// degree): size 3 is a triad, 4 a seventh chord
func Diatonic(s *scale.Scale, degree, size int) Chord {
	c := make(Chord, size)
	for i := range c {
		c[i] = s.Degree(degree + 2*i)
	}
	return c
}

// Transpose moves all notes by semitones
func (c Chord) Transpose(semitones float64) Chord {
	out := make(Chord, len(c))
	for i, p := range c {
		out[i] = p + semitones
	}
	return out
}

// Invert returns the n-th inversion: n times the lowest note moves an octave up. A negative n moves
// the highest note an octave down
func (c Chord) Invert(n int) Chord {
	out := slices.Clone(c)
	for ; n > 0; n-- {
		out = append(out[1:], out[0]+12)
	}
	for ; n < 0; n++ {
		out = append(Chord{out[len(out)-1] - 12}, out[:len(out)-1]...)
	}
	return out
}

// Open spreads the chord by dropping the second highest note an octave (drop 2 voicing)
func (c Chord) Open() Chord {
	if len(c) < 3 {
		return slices.Clone(c)
	}
	out := slices.Clone(c)
	out[len(out)-2] -= 12
	slices.Sort(out)
	return out
}

// VoiceLeader chooses for every chord the inversion and octave closest to the previous chord, for
// smooth voice leading. It is a transform.Transformer[[]float64]
type VoiceLeader struct {
	low, high float64
	prev      []float64
}

// NewVoiceLeader creates a voice leader that keeps the chord notes in [low, high] where possible
// (low >= high turns the range off)
func NewVoiceLeader(low, high float64) *VoiceLeader { return &VoiceLeader{low: low, high: high} }

func (v *VoiceLeader) Transform(in []float64) []float64 {
	if len(in) == 0 {
		return in
	}
	c := Chord(slices.Clone(in))
	slices.Sort(c)

	if v.prev == nil {
		v.prev = c
		return slices.Clone(c)
	}

	best, bestCost := c, math.Inf(1)
	for inv := range len(c) {
		for oct := -3; oct <= 3; oct++ {
			cand := c.Invert(inv).Transpose(float64(12 * oct))
			if cost := v.cost(cand); cost < bestCost {
				best, bestCost = cand, cost
			}
		}
	}

	v.prev = best
	return slices.Clone(best)
}

// cost is the distance the voices travel, with a penalty for leaving the range
func (v *VoiceLeader) cost(cand Chord) float64 {
	cost := 0.0
	for i := range min(len(cand), len(v.prev)) {
		cost += math.Abs(cand[i] - v.prev[i])
	}
	if v.low < v.high {
		for _, p := range cand {
			if p < v.low {
				cost += 1000 + v.low - p
			} else if p > v.high {
				cost += 1000 + p - v.high
			}
		}
	}
	return cost
}

func (v *VoiceLeader) Reset() { v.prev = nil }
