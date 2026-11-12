// Package funnel provides Funnel, a shaper that snaps its input to the nearest value of an allowed
// set: a sieve, a scale or any list. For input x between the neighbours a <= x <= b of the set the
// output is a if x-a < b-x, b if b-x < x-a, and one of the two on a tie, decided by the Tie option:
// Lower (the default), Upper, or Random (a draw from the genny random stream, so genny.Seed makes
// it repeatable; the stream is only touched on a real tie). Input beyond the smallest or largest value
// snaps to that value.
//
// An empty set (also a sieve without members) is not an error: the input is returned unchanged.
// Non finite values in a list are dropped, non finite input is returned unchanged.
package funnel

import (
	"math"
	"sort"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/scale"
	"github.com/almerlucke/genny/sieve"
)

// Tie decides which neighbour is taken when the input is exactly halfway between two values
type Tie int

// The tie rules
const (
	// Lower rounds a tie down
	Lower Tie = iota
	// Upper rounds a tie up
	Upper
	// Random picks either with equal probability from the genny stream
	Random
)

// Funnel snaps to the nearest allowed value
type Funnel struct {
	// bracket returns the largest allowed value <= x and the smallest >= x, ok flags tell if they exist
	bracket func(x float64) (lo, hi float64, hasLo, hasHi bool)
	tie     Tie
	empty   bool
}

// New creates a Funnel from a list of allowed values (any order, duplicates allowed, copied). An
// empty list gives a Funnel that returns its input
func New(values []float64) *Funnel {
	set := make([]float64, 0, len(values))
	for _, v := range values {
		if !math.IsNaN(v) && !math.IsInf(v, 0) {
			set = append(set, v)
		}
	}
	sort.Float64s(set)
	if len(set) == 0 {
		return &Funnel{empty: true}
	}
	return &Funnel{bracket: func(x float64) (float64, float64, bool, bool) {
		i := sort.SearchFloat64s(set, x) // first index with set[i] >= x
		if i < len(set) && set[i] == x {
			return x, x, true, true
		}
		var lo, hi float64
		if i > 0 {
			lo = set[i-1]
		}
		if i < len(set) {
			hi = set[i]
		}
		return lo, hi, i > 0, i < len(set)
	}}
}

// FromSieve creates a Funnel from the members of a sieve as pitches base + member
func FromSieve(s sieve.Sieve, base float64) *Funnel { return New(sieve.Scale(s, base)) }

// FromScale creates a Funnel of all pitches of a scale, in every octave. It panics if sc is nil
func FromScale(sc *scale.Scale) *Funnel {
	if sc == nil {
		panic("funnel: FromScale: scale is nil")
	}
	return &Funnel{bracket: func(x float64) (float64, float64, bool, bool) {
		d := sc.DegreeOf(x)
		lo, hi := math.Inf(-1), math.Inf(1)
		for _, deg := range []int{d - 1, d, d + 1} {
			v := sc.Degree(deg)
			if v <= x && v > lo {
				lo = v
			}
			if v >= x && v < hi {
				hi = v
			}
		}
		return lo, hi, !math.IsInf(lo, 0), !math.IsInf(hi, 0)
	}}
}

// WithTie sets the tie rule and returns the Funnel
func (f *Funnel) WithTie(t Tie) *Funnel {
	f.tie = t
	return f
}

// Shape snaps x to the nearest allowed value
func (f *Funnel) Shape(x float64) float64 {
	// beyond 1e12 a float64 has no room for a scale degree computation, such values are left alone
	if f.empty || math.IsNaN(x) || math.IsInf(x, 0) || math.Abs(x) > 1e12 {
		return x
	}
	lo, hi, hasLo, hasHi := f.bracket(x)
	switch {
	case !hasLo && !hasHi:
		return x
	case !hasLo:
		return hi
	case !hasHi:
		return lo
	}
	dl, dh := x-lo, hi-x
	switch {
	case dl < dh:
		return lo
	case dh < dl:
		return hi
	}
	switch f.tie {
	case Upper:
		return hi
	case Random:
		if genny.Float64() < 0.5 {
			return lo
		}
		return hi
	}
	return lo
}
