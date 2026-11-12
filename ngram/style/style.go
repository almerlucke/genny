// Package style learns the style of a melody (events from package note, for instance read from a MIDI
// file with midi.Read) and plays new melodies in that style. It is a small layer on package ngram: a
// melody becomes a sequence of (pitch, duration) tokens, the chain learns what follows what, and the
// generator hands out new notes.
//
// It is a separate package so that ngram stays a plain chain over any comparable value, without
// knowing about notes, and so that the melody extraction and rendering have a home of their own.
//
// By default pitches and durations are learned as they are. With Intervals the pitch is learned as the
// step from the previous note, and with Ratios the duration as the factor of the previous duration, so the style
// does not depend on the key or the tempo of the example and can be played from any pitch (Transpose) at any speed
// (Scale)
package style

import (
	"cmp"
	"math"
	"slices"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/ngram"
	"github.com/almerlucke/genny/note"
)

// Note is a pitch in semitones and a duration in steps. The duration is the time up to the next note
type Note struct {
	Pitch    float64
	Duration float64
}

// token is what the chain learns: the pitch or interval, and the duration or ratio, rounded so equal
// moves are equal values
type token struct{ P, D float64 }

func round(x float64, decimals float64) float64 {
	m := math.Pow(10, decimals)
	return math.Round(x*m) / m
}

// Melody turns events into a single line: the notes in order of start, of simultaneous notes the highest
// one (the top of a chord), and as duration the time up to the start of the next note, so a rest after
// a note lengthens it. The last note keeps its own duration. Notes that start at (nearly) the same time
// count as one
func Melody(events []note.Event) []Note {
	es := slices.Clone(events)
	slices.SortStableFunc(es, func(a, b note.Event) int {
		return cmp.Or(cmp.Compare(a.Start, b.Start), -cmp.Compare(a.Pitch, b.Pitch))
	})

	var top []note.Event
	for _, e := range es {
		if len(top) > 0 && e.Start-top[len(top)-1].Start < 1e-6 {
			continue
		}
		top = append(top, e)
	}

	out := make([]Note, len(top))
	for i, e := range top {
		d := e.Duration
		if i+1 < len(top) {
			d = top[i+1].Start - e.Start
		}
		out[i] = Note{Pitch: e.Pitch, Duration: d}
	}
	return out
}

// Events renders a phrase as events one after the other, from step start, with a velocity in [0, 1]
func Events(phrase []Note, start, velocity float64) []note.Event {
	out := make([]note.Event, len(phrase))
	t := start
	for i, n := range phrase {
		out[i] = note.Event{Pitch: n.Pitch, Velocity: velocity, Start: t, Duration: n.Duration}
		t += n.Duration
	}
	return out
}

// Length is the total duration of a phrase in steps
func Length(phrase []Note) float64 {
	t := 0.0
	for _, n := range phrase {
		t += n.Duration
	}
	return t
}

// Option changes how a style is learned
type Option func(*Style)

// Intervals learns pitches as the interval from the previous note, so the style is the same in any key
func Intervals() Option { return func(s *Style) { s.intervals = true } }

// Ratios learns durations as the ratio to the previous duration, so the style is the same at any speed
func Ratios() Option { return func(s *Style) { s.ratios = true } }

// Style plays notes in the style of an example. It is a genny.Generator of Note
type Style struct {
	chain          *ngram.Chain[token]
	intervals      bool
	ratios         bool
	startPitch     float64 // the pitch the generator starts from (intervals), or the first token's offset
	startDur       float64
	lo, hi         float64 // pitch range of the example, a walk of intervals is folded back into it
	minDur, maxDur float64
	pitch, dur     float64 // where the intervals and ratios are relative to
	shift, scale   float64 // transposition and duration factor
	tokens         []token
	Velocity       float64 // mean velocity of the example, for Events
}

// Learn builds a style from the melody in events (see Melody). It looks back order notes (1 is the simplest,
// 2 or 3 follow the example more closely). Like ngram.Learn it panics if order is less than 1 or the
// melody has no more than order notes
func Learn(events []note.Event, order int, opts ...Option) *Style {
	mel := Melody(events)
	if order < 1 || len(mel) <= order {
		panic("style: need order >= 1 and a melody longer than the order")
	}
	s := &Style{scale: 1, lo: math.Inf(1), hi: math.Inf(-1), minDur: math.Inf(1)}
	for _, o := range opts {
		o(s)
	}

	for _, e := range events {
		s.Velocity += e.Velocity / float64(len(events))
	}
	if s.Velocity <= 0 {
		s.Velocity = 0.8
	}

	prev := Note{Pitch: mel[0].Pitch, Duration: mel[0].Duration}
	for _, n := range mel {
		t := token{P: n.Pitch, D: n.Duration}
		if s.intervals {
			t.P = n.Pitch - prev.Pitch
		}
		if s.ratios {
			t.D = n.Duration / prev.Duration
		}
		t.P, t.D = round(t.P, 3), round(t.D, 3)
		s.tokens = append(s.tokens, t)
		prev = n

		s.lo, s.hi = min(s.lo, n.Pitch), max(s.hi, n.Pitch)
		s.minDur, s.maxDur = min(s.minDur, n.Duration), max(s.maxDur, n.Duration)
	}

	// the generated notes follow the first `order` notes, which only set the context
	last := mel[order-1]
	s.startPitch, s.startDur = last.Pitch, last.Duration
	s.chain = ngram.Learn(s.tokens, order)
	s.Reset()
	return s
}

// Transpose moves all generated pitches by a number of semitones, on top of earlier calls. It returns
// the style to chain calls
func (s *Style) Transpose(semitones float64) *Style {
	s.shift += semitones
	return s
}

// Scale multiplies all generated durations by a factor (2 plays at half the speed), on top of earlier
// calls. It panics if factor is not positive
func (s *Style) Scale(factor float64) *Style {
	if !(factor > 0) {
		panic("style: duration factor must be positive")
	}
	s.scale *= factor
	return s
}

// Generate returns the next note. Pitches stay in the range of the example, durations between its
// shortest and longest
func (s *Style) Generate() Note {
	t := s.chain.Generate()

	p, d := t.P, t.D
	if s.intervals {
		p = s.pitch + t.P
		// a walk of intervals can wander off: fold it back into the range of the example by octaves
		if p > s.hi {
			p -= 12 * math.Ceil((p-s.hi)/12)
		}
		if p < s.lo {
			p += 12 * math.Ceil((s.lo-p)/12)
		}
		s.pitch = p
	}
	if s.ratios {
		d = min(max(s.dur*t.D, s.minDur), s.maxDur)
		s.dur = d
	}
	return Note{Pitch: p + s.shift, Duration: d * s.scale}
}

func (s *Style) Continuous() bool { return true }

func (s *Style) Done() bool { return false }

// Reset starts again from the context the example starts with
func (s *Style) Reset() {
	s.chain.Reset()
	s.pitch, s.dur = s.startPitch, s.startDur
}

// Phrase returns the next n notes
func (s *Style) Phrase(n int) []Note {
	out := make([]Note, 0, max(n, 0))
	for range n {
		out = append(out, s.Generate())
	}
	return out
}

// Fill returns notes up to a total duration of length steps, the last one shortened to fit
func (s *Style) Fill(length float64) []Note {
	var out []Note
	t := 0.0
	for t < length-1e-9 {
		n := s.Generate()
		n.Duration = min(n.Duration, length-t)
		if !(n.Duration > 0) {
			break // cannot happen for a melody of positive durations
		}
		out = append(out, n)
		t += n.Duration
	}
	return out
}

var _ genny.Generator[Note] = (*Style)(nil)
