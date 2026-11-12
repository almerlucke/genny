// Package microtonal plays music in other tunings (31-EDO, just intonation, Bohlen-Pierce, any Scala
// scale) through ordinary MIDI. A Tuning turns scale degrees into fractional pitches (semitones); Render
// maps those to MIDI notes with a 14 bit pitch bend that fills the cents between the equal tempered
// notes. Two strategies exist: Single puts a whole line on one channel (monophonic music only, as a
// pitch bend moves every sounding note of its channel) and MPE gives every simultaneous note a channel of
// its own, so chords in 31-EDO or just intonation work. The tracks are meant for midi.File
//
// The midi package writes the bend itself (midi.Track.PitchBend), so the bend is written before each note
// start and not reset at the end: the next note on that channel sets its own
package microtonal

import (
	"math"

	"github.com/almerlucke/genny/scale"
)

// Tuning is a scale that repeats every Period cents. Degree 0 is the Root pitch
type Tuning struct {
	// Root is the pitch of degree 0 in semitones (60 is middle C)
	Root float64
	// Cents are the degrees within one period, in cents above the root: ascending, the first is 0
	Cents []float64
	// Period is the interval at which the scale repeats in cents (1200 for an octave)
	Period float64
}

func check(period float64, cents []float64) {
	if !(period > 0) || math.IsInf(period, 0) || len(cents) == 0 || cents[0] != 0 {
		panic("microtonal: a tuning needs a positive period and cents starting at 0")
	}
	for i, c := range cents {
		if math.IsNaN(c) || (i > 0 && c <= cents[i-1]) || c >= period {
			panic("microtonal: cents must ascend strictly and stay below the period")
		}
	}
}

// FromCents creates a tuning from degrees in cents (starting with 0) that repeats every period cents. It
// panics if the cents do not ascend strictly from 0 to below the period
func FromCents(root, period float64, cents ...float64) Tuning {
	check(period, cents)
	return Tuning{Root: root, Cents: append([]float64(nil), cents...), Period: period}
}

// EDO divides the octave in n equal steps: 12 is the usual scale, 19, 24 and 31 are popular. It panics if n < 1
func EDO(root float64, n int) Tuning { return EqualDivision(root, n, 1200) }

// EqualDivision divides any period (in cents) in n equal steps, for instance 13 steps of the tritave
// (1901.955 cents) for Bohlen-Pierce. It panics if n < 1 or the period is not positive
func EqualDivision(root float64, n int, period float64) Tuning {
	if n < 1 {
		panic("microtonal: divisions must be at least 1")
	}
	cents := make([]float64, n)
	for i := range cents {
		cents[i] = period * float64(i) / float64(n)
	}
	return FromCents(root, period, cents...)
}

// Just creates a tuning from frequency ratios: period is the ratio at which it repeats (2 for an octave,
// 3 for a tritave) and the ratios are the degrees, the first of which is 1. It panics under the conditions of FromCents
func Just(root, period float64, ratios ...float64) Tuning {
	cents := make([]float64, len(ratios))
	for i, r := range ratios {
		cents[i] = 1200 * math.Log2(r)
	}
	return FromCents(root, 1200*math.Log2(period), cents...)
}

// FromScl creates a tuning from a Scala file (see scale.ParseScl). Scl.Cents ends with the period; the
// pitches before it are the degrees. It panics if the Scl is empty or does not ascend
func FromScl(root float64, s scale.Scl) Tuning {
	if len(s.Cents) == 0 {
		panic("microtonal: empty scl")
	}
	n := len(s.Cents) - 1
	return FromCents(root, s.Cents[n], append([]float64{0}, s.Cents[:n]...)...)
}

// Len is the number of degrees in a period
func (t Tuning) Len() int { return len(t.Cents) }

func floorDiv(a, b int) int {
	q := a / b
	if a%b != 0 && (a < 0) != (b < 0) {
		q--
	}
	return q
}

// Cent returns the cents of a degree above the root; degrees beyond the period wrap into the next
// period and negative degrees go below the root
func (t Tuning) Cent(degree int) float64 {
	n := len(t.Cents)
	if n == 0 {
		return 0
	}
	o := floorDiv(degree, n)
	return t.Cents[degree-o*n] + float64(o)*t.Period
}

// Pitch returns the pitch in semitones of a degree: Root plus the cents of the degree divided by 100
func (t Tuning) Pitch(degree int) float64 { return t.Root + t.Cent(degree)/100 }

// Note returns the pitch of a degree in an octave (period) above the root, so Note(3, -1) is degree 3 one period down
func (t Tuning) Note(degree, octave int) float64 { return t.Pitch(degree + octave*len(t.Cents)) }

// Nearest returns the degree whose pitch is closest to a pitch in semitones
func (t Tuning) Nearest(pitch float64) int {
	n := len(t.Cents)
	if n == 0 || math.IsNaN(pitch) || math.IsInf(pitch, 0) {
		return 0
	}
	o := int(math.Floor((pitch - t.Root) * 100 / t.Period))
	best, bestD := 0, math.Inf(1)
	for d := (o - 1) * n; d <= (o+2)*n; d++ {
		if dist := math.Abs(t.Pitch(d) - pitch); dist < bestD {
			best, bestD = d, dist
		}
	}
	return best
}

// Snap moves a pitch to the nearest degree of the tuning
func (t Tuning) Snap(pitch float64) float64 { return t.Pitch(t.Nearest(pitch)) }
