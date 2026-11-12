// Package line works on whole lines of notes (slices of note.Event) instead of on values: transpose,
// stretch, invert, reverse, shift, merge, cut, loop, quantize, thin out and make canons. Every function
// returns a new slice and leaves its input alone
package line

import (
	"cmp"
	"math"
	"slices"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/note"
	"github.com/almerlucke/genny/scale"
)

// Transpose moves all pitches by semitones
func Transpose(events []note.Event, semitones float64) []note.Event {
	out := slices.Clone(events)
	for i := range out {
		out[i].Pitch += semitones
	}
	return out
}

// Stretch multiplies start times and durations by factor: 2 doubles the length of the line (augmentation),
// 0.5 halves it (diminution). It panics if factor is not positive
func Stretch(events []note.Event, factor float64) []note.Event {
	if !(factor > 0) || math.IsInf(factor, 0) {
		panic("line: stretch factor must be positive")
	}
	out := slices.Clone(events)
	for i := range out {
		out[i].Start *= factor
		out[i].Duration *= factor
	}
	return out
}

// Augment makes the line factor times slower (longer), see Stretch
func Augment(events []note.Event, factor float64) []note.Event { return Stretch(events, factor) }

// Diminish makes the line factor times faster (shorter), see Stretch
func Diminish(events []note.Event, factor float64) []note.Event {
	if !(factor > 0) || math.IsInf(factor, 0) {
		panic("line: diminish factor must be positive")
	}
	return Stretch(events, 1/factor)
}

// Invert mirrors the pitches around axis. With a scale the mirror is diatonic: scale degrees are mirrored
// around the degree nearest to the axis, so the result stays in the scale (pitches outside the scale are
// first taken to the nearest degree). With a nil scale it is chromatic: pitch becomes 2*axis-pitch
func Invert(events []note.Event, axis float64, s *scale.Scale) []note.Event {
	out := slices.Clone(events)
	for i := range out {
		if s == nil {
			out[i].Pitch = 2*axis - out[i].Pitch
		} else {
			out[i].Pitch = s.Degree(2*s.DegreeOf(axis) - s.DegreeOf(out[i].Pitch))
		}
	}
	return out
}

// Retrograde plays the line backwards in time: a note that ended last starts first, durations stay as
// they are. The line keeps its first start time and its length, so Retrograde twice gives the original
func Retrograde(events []note.Event) []note.Event {
	if len(events) == 0 {
		return nil
	}
	first, last := math.Inf(1), math.Inf(-1)
	for _, e := range events {
		first = min(first, e.Start)
		last = max(last, e.Start+e.Duration)
	}
	out := slices.Clone(events)
	for i := range out {
		out[i].Start = first + last - (out[i].Start + out[i].Duration)
	}
	sortByStart(out)
	return out
}

// Shift moves the line in time by steps (negative is earlier). A start before 0 is clamped to 0
func Shift(events []note.Event, steps float64) []note.Event {
	out := slices.Clone(events)
	for i := range out {
		out[i].Start = max(out[i].Start+steps, 0)
	}
	return out
}

// Merge combines lines into one, ordered by start (notes that start together keep their order)
func Merge(lines ...[]note.Event) []note.Event {
	var out []note.Event
	for _, l := range lines {
		out = append(out, l...)
	}
	sortByStart(out)
	return out
}

// Slice returns the notes that start in [from, to), moved so that from becomes time 0
func Slice(events []note.Event, from, to float64) []note.Event {
	var out []note.Event
	for _, e := range events {
		if e.Start >= from && e.Start < to {
			e.Start -= from
			out = append(out, e)
		}
	}
	return out
}

// Loop repeats the line n times, every pass length steps after the one before. It panics if length is not
// positive. n < 1 gives an empty line
func Loop(events []note.Event, n int, length float64) []note.Event {
	if !(length > 0) {
		panic("line: loop length must be positive")
	}
	var out []note.Event
	for k := range max(n, 0) {
		out = append(out, Shift(events, float64(k)*length)...)
	}
	return out
}

// Quantize moves every pitch to the nearest pitch of the scale
func Quantize(events []note.Event, s *scale.Scale) []note.Event {
	out := slices.Clone(events)
	for i := range out {
		out[i].Pitch = s.Quantize(out[i].Pitch)
	}
	return out
}

// Canon makes voices that all play the line: voice i enters i*delay steps later and i*interval
// semitones higher (a canon at the fifth after 2 bars of 16 steps is delay 32, interval 7). Voice 0 is
// the line itself. It panics if voices < 1
func Canon(events []note.Event, delay, interval float64, voices int) [][]note.Event {
	if voices < 1 {
		panic("line: a canon needs at least one voice")
	}
	out := make([][]note.Event, voices)
	for i := range out {
		out[i] = Transpose(Shift(events, float64(i)*delay), float64(i)*interval)
	}
	return out
}

// Thin keeps the notes for which the gate gives true, one gate value per note: a rhythmic thinning
// out of a line. It panics if gate is nil
func Thin(events []note.Event, gate genny.Generator[bool]) []note.Event {
	if gate == nil {
		panic("line: thin needs a gate")
	}
	var out []note.Event
	for _, e := range events {
		if gate.Generate() {
			out = append(out, e)
		}
	}
	return out
}

func sortByStart(events []note.Event) {
	slices.SortStableFunc(events, func(a, b note.Event) int { return cmp.Compare(a.Start, b.Start) })
}
