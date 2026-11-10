// Package note turns the output of generators into note events: a Voice combines a generator of
// pitches (with rests) with generators of velocity, length, timing offset and ratchets, and renders
// them step by step. The events can be written to a MIDI file with package midi
package note

import (
	"cmp"
	"slices"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/rest"
)

// Event is a note. Time is in steps (for instance sixteenths), pitch in semitones (a MIDI note number,
// 60 is middle C) and velocity in [0, 1]
type Event struct {
	Pitch    float64
	Velocity float64
	Start    float64
	Duration float64
}

// Always turns a generator of values into one that never rests, to use a plain pitch generator in a Voice
func Always[T any](gen genny.Generator[T]) genny.Generator[rest.Optional[T]] {
	return &always[T]{gen}
}

type always[T any] struct{ gen genny.Generator[T] }

func (a *always[T]) Generate() rest.Optional[T] { return rest.Some(a.gen.Generate()) }
func (a *always[T]) Continuous() bool           { return a.gen.Continuous() }
func (a *always[T]) Done() bool                 { return a.gen.Done() }
func (a *always[T]) Reset()                     { a.gen.Reset() }

// Voice is one line of notes. Set Pitch for single notes or Chord for chords (exactly one of them), the
// other generators are optional. Every step all generators advance once, also on a rest, so they
// stay together
type Voice struct {
	// Pitch generates the pitch of each step, or a rest
	Pitch genny.Generator[rest.Optional[float64]]
	// Chord generates the pitches that sound together in each step, or a rest
	Chord genny.Generator[rest.Optional[[]float64]]
	// Velocity in [0, 1] (default 0.8)
	Velocity genny.Generator[float64]
	// Length of the note in steps (default 1)
	Length genny.Generator[float64]
	// Offset moves the start of the step, in steps (default 0): swing and humanize. The first
	// notes cannot start before time 0
	Offset genny.Generator[float64]
	// Ratchet gives the offsets of the repeats within the step, see package ratchet (default one
	// hit). The length is divided over the repeats
	Ratchet genny.Generator[[]float64]
}

// Render plays the voice for a number of steps and returns the events, ordered by start. It panics if
// not exactly one of Pitch and Chord is set
func (v Voice) Render(steps int) []Event {
	if (v.Pitch == nil) == (v.Chord == nil) {
		panic("note: a voice needs exactly one of Pitch and Chord")
	}

	var events []Event
	for step := range steps {
		var pitches []float64
		played := false
		if v.Pitch != nil {
			if p := v.Pitch.Generate(); p.Played() {
				pitches, played = []float64{p.Value}, true
			}
		} else if c := v.Chord.Generate(); c.Played() {
			pitches, played = c.Value, true
		}

		velocity, length, offset, hits := 0.8, 1.0, 0.0, []float64{0}
		if v.Velocity != nil {
			velocity = v.Velocity.Generate()
		}
		if v.Length != nil {
			length = v.Length.Generate()
		}
		if v.Offset != nil {
			offset = v.Offset.Generate()
		}
		if v.Ratchet != nil {
			if h := v.Ratchet.Generate(); len(h) > 0 {
				hits = h
			}
		}

		if !played {
			continue
		}
		for _, h := range hits {
			for _, p := range pitches {
				events = append(events, Event{
					Pitch:    p,
					Velocity: min(max(velocity, 0), 1),
					Start:    max(float64(step)+offset+h, 0),
					Duration: max(length/float64(len(hits)), 1e-3),
				})
			}
		}
	}

	slices.SortStableFunc(events, func(a, b Event) int { return cmp.Compare(a.Start, b.Start) })
	return events
}
