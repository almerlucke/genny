// Package samplehold provides sample and hold and track and hold shapers driven by a trigger
// generator. Every call to Shape asks the trigger for one bool.
//
// SampleHold: on a true trigger the held value becomes the input, otherwise it stays, and the held
// value is the output. TrackHold: while the trigger is true the output follows the input (track),
// when it is false the last tracked value is held. Both start holding 0. Non finite input is
// ignored.
package samplehold

import (
	"math"

	"github.com/almerlucke/genny"
)

// SampleHold samples the input on a trigger
type SampleHold struct {
	trig genny.Generator[bool]
	held float64
}

// New creates a sample and hold with a trigger. It panics if trig is nil
func New(trig genny.Generator[bool]) *SampleHold {
	if trig == nil {
		panic("samplehold: New: trigger is nil")
	}
	return &SampleHold{trig: trig}
}

// Shape returns the held value, after sampling x if the trigger fires
func (s *SampleHold) Shape(x float64) float64 {
	if s.trig.Generate() && !math.IsNaN(x) && !math.IsInf(x, 0) {
		s.held = x
	}
	return s.held
}

// Held is the current held value
func (s *SampleHold) Held() float64 { return s.held }

// Reset sets the held value to 0 and resets the trigger
func (s *SampleHold) Reset() {
	s.held = 0
	s.trig.Reset()
}

// TrackHold follows the input while the trigger is true and holds when it is false
type TrackHold struct {
	trig genny.Generator[bool]
	held float64
}

// NewTrack creates a track and hold with a trigger. It panics if trig is nil
func NewTrack(trig genny.Generator[bool]) *TrackHold {
	if trig == nil {
		panic("samplehold: NewTrack: trigger is nil")
	}
	return &TrackHold{trig: trig}
}

// Shape returns x while tracking, else the last tracked value
func (t *TrackHold) Shape(x float64) float64 {
	if t.trig.Generate() && !math.IsNaN(x) && !math.IsInf(x, 0) {
		t.held = x
	}
	return t.held
}

// Reset sets the held value to 0 and resets the trigger
func (t *TrackHold) Reset() {
	t.held = 0
	t.trig.Reset()
}
