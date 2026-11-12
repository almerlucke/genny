package samplehold

import (
	"math"
	"testing"

	"github.com/almerlucke/genny/float/shape"
	"github.com/almerlucke/genny/sequence"
)

var (
	_ shape.Shaper = (*SampleHold)(nil)
	_ shape.Shaper = (*TrackHold)(nil)
)

func trig(loop bool, v ...bool) *sequence.Sequence[bool] {
	if loop {
		return sequence.NewLoop(v...)
	}
	return sequence.New(v...)
}

func TestSampleHold(t *testing.T) {
	s := New(trig(true, true, false, false, true, false))
	in := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	want := []float64{1, 1, 1, 4, 4, 6, 6, 6, 9, 9}
	for i, x := range in {
		if got := s.Shape(x); got != want[i] {
			t.Errorf("step %d: got %v want %v", i, got, want[i])
		}
	}
}

func TestSampleHoldReset(t *testing.T) {
	s := New(trig(true, false, true))
	a := []float64{s.Shape(1), s.Shape(2), s.Shape(3)}
	if a[0] != 0 || a[1] != 2 || a[2] != 2 {
		t.Fatalf("got %v", a)
	}
	s.Reset()
	b := []float64{s.Shape(1), s.Shape(2), s.Shape(3)}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("not repeatable after Reset: %v vs %v", a, b)
		}
	}
	if s.Held() != 2 {
		t.Error("Held")
	}
}

func TestTrackHold(t *testing.T) {
	s := NewTrack(trig(true, true, true, false, false, true))
	in := []float64{1, 2, 3, 4, 5, 6, 7}
	want := []float64{1, 2, 2, 2, 5, 6, 7}
	for i, x := range in {
		if got := s.Shape(x); got != want[i] {
			t.Errorf("step %d: got %v want %v", i, got, want[i])
		}
	}
	s.Reset()
	if s.Shape(9) != 9 {
		t.Error("Reset must reset the trigger")
	}
}

func TestNonFiniteIgnored(t *testing.T) {
	s := New(trig(true, true))
	s.Shape(0.5)
	for _, x := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if got := s.Shape(x); got != 0.5 {
			t.Errorf("Shape(%v) = %v", x, got)
		}
	}
}

func TestPanics(t *testing.T) {
	for name, f := range map[string]func(){
		"New nil":      func() { New(nil) },
		"NewTrack nil": func() { NewTrack(nil) },
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: expected panic", name)
				}
			}()
			f()
		}()
	}
}
