package constrain

import (
	"slices"

	"github.com/almerlucke/genny"
)

// Walk is a melodic walk through intervals, see ScaleIntervals and RandomIntervals. It is a
// genny.Generator[int] that never ends (unless it has no intervals: then it is done at once)
type Walk struct {
	first, cur int
	begun      bool
	empty      bool
	step       func(cur int) int
	reset      func()
}

// Generate returns the first note, then the following notes of the walk
func (w *Walk) Generate() int {
	if w.empty {
		return 0
	}
	if !w.begun {
		w.begun, w.cur = true, w.first
		return w.cur
	}
	w.cur = w.step(w.cur)
	return w.cur
}

func (w *Walk) Continuous() bool { return !w.empty }
func (w *Walk) Done() bool       { return w.empty }

// Reset goes back to the first note and the first interval
func (w *Walk) Reset() {
	w.begun = false
	if w.reset != nil {
		w.reset()
	}
}

var _ genny.Generator[int] = (*Walk)(nil)

func pc(n int) int { return ((n % 12) + 12) % 12 }

// ScaleIntervals is a melodic walk: it starts on first, then adds the intervals one after the other (and
// cycles through them), and only accepts a result that is in the allowed set (the pitches of a scale,
// over as many octaves as the walk will reach). A result that is not allowed is skipped, the next
// interval is tried instead, up to Stop times (default 100): if none is allowed the walk repeats
// the current note. The first note is given as it is, allowed or not. With PitchClass the allowed
// set holds for all octaves (compared mod 12). An empty allowed set allows everything. Without
// intervals the walk is done at once and generates 0
func ScaleIntervals(first int, intervals []int, allowed []int, opts ...Option) *Walk {
	o := newOptions(opts)
	ok := set(allowed)
	if o.pitchClass {
		ok = set(nil)
		for _, a := range allowed {
			ok[pc(a)] = true
		}
	}
	intervals = slices.Clone(intervals)
	idx := 0
	return &Walk{
		first: first,
		empty: len(intervals) == 0,
		step: func(cur int) int {
			for range o.stop {
				cand := cur + intervals[idx%len(intervals)]
				idx++
				key := cand
				if o.pitchClass {
					key = pc(cand)
				}
				if len(allowed) == 0 || ok[key] {
					return cand
				}
			}
			return cur
		},
		reset: func() { idx = 0 },
	}
}

// RandomIntervals is a melodic walk that moves up or down by one of the intervals, chosen at random
// (with the random stream of genny), and stays within [lo, hi]. It starts in the middle of the range.
// When a move would leave the range the other direction is tried, then another interval, up to Stop
// times: if nothing fits the walk repeats the note. It panics if lo > hi, if there are no intervals or
// if an interval is negative
func RandomIntervals(lo, hi int, intervals ...int) *Walk {
	return RandomIntervalsOpt(lo, hi, intervals)
}

// RandomIntervalsOpt is RandomIntervals with options (Stop)
func RandomIntervalsOpt(lo, hi int, intervals []int, opts ...Option) *Walk {
	if lo > hi {
		panic("constrain: RandomIntervals needs lo <= hi")
	}
	if len(intervals) == 0 {
		panic("constrain: RandomIntervals needs at least one interval")
	}
	if slices.Min(intervals) < 0 {
		panic("constrain: RandomIntervals intervals must not be negative")
	}
	o := newOptions(opts)
	intervals = slices.Clone(intervals)
	return &Walk{
		first: lo + (hi-lo)/2,
		step: func(cur int) int {
			for range o.stop {
				iv := intervals[genny.Intn(len(intervals))]
				if genny.Intn(2) == 0 {
					iv = -iv
				}
				if c := cur + iv; c >= lo && c <= hi {
					return c
				}
				if c := cur - iv; c >= lo && c <= hi {
					return c
				}
			}
			return cur
		},
	}
}
