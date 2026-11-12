package phase

import (
	"math"
	"testing"

	"github.com/almerlucke/genny/note"
)

func pattern() []note.Event {
	var out []note.Event
	for i, p := range []float64{64, 66, 71, 73} {
		out = append(out, note.Event{Pitch: p, Velocity: 0.8, Start: float64(i), Duration: 1})
	}
	return out
}

func TestLengthCycles(t *testing.T) {
	if Length(pattern()) != 4 || Length(nil) != 0 {
		t.Error("length")
	}
	if Cycles(pattern(), 0.5) != 8 || Cycles(pattern(), 1) != 4 || Cycles(nil, 1) != 0 || Cycles(pattern(), 0) != 0 {
		t.Error("cycles")
	}
}

func TestEvents(t *testing.T) {
	p := pattern()
	n := Cycles(p, 0.5)
	a, b := Events(p, n, 0.5)
	if len(a) != n*len(p) {
		t.Errorf("first voice has %d notes", len(a))
	}
	// the second voice starts cycle k at k*(l-drift); at cycle n it is back in unison with the first
	// voice's cycle n-1
	found := false
	for _, e := range b {
		if e.Start == float64(n-1)*4 && e.Pitch == 64 {
			found = true
		}
	}
	if !found {
		t.Error("no unison after n cycles")
	}
	// halfway it is half a cycle (2 steps) off: its note 64 starts 2 steps before a note of the first voice
	if b[len(p)*(n/2)].Start != float64(n/2)*3.5 {
		t.Error("drift per cycle")
	}
	total := float64(n) * 4
	for _, e := range b {
		if e.Start >= total {
			t.Error("second voice beyond the end", e)
		}
	}
	if len(b) <= len(a)-len(p) {
		t.Error("second voice should play more cycles")
	}
	// the input is not changed
	if p[1].Start != 1 {
		t.Error("mutated")
	}
	if x, y := Events(nil, 3, 1); x != nil || y != nil {
		t.Error("empty")
	}
}

func TestPanics(t *testing.T) {
	for name, f := range map[string]func(){
		"cycles 0":      func() { Events(pattern(), 0, 0.5) },
		"drift 0":       func() { Events(pattern(), 2, 0) },
		"drift too big": func() { Events(pattern(), 2, 4) },
		"drift nan":     func() { Events(pattern(), 2, math.NaN()) },
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
