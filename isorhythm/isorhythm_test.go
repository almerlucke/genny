package isorhythm

import (
	"slices"
	"testing"

	"github.com/almerlucke/genny/note"
)

func TestEvents(t *testing.T) {
	tests := []struct {
		name          string
		talea, color  []float64
		length        float64
		wantN         int
		wantFirstRest float64 // start of the second note
	}{
		{"simple", []float64{1}, []float64{60}, 4, 4, 1},
		{"rest", []float64{1, -1}, []float64{60}, 4, 2, 2},
		{"empty talea", nil, []float64{60}, 4, 0, 0},
		{"empty color", []float64{1}, nil, 4, 0, 0},
		{"zero length", []float64{1}, []float64{60}, 0, 0, 0},
		{"all zero talea", []float64{0, 0}, []float64{60}, 4, 0, 0},
		{"cut at length", []float64{3}, []float64{60}, 4, 2, 3},
	}
	for _, tt := range tests {
		got := Events(tt.talea, tt.color, tt.length)
		if len(got) != tt.wantN {
			t.Errorf("%s: %d notes, want %d", tt.name, len(got), tt.wantN)
			continue
		}
		if tt.wantN > 1 && got[1].Start != tt.wantFirstRest {
			t.Errorf("%s: second note at %v", tt.name, got[1].Start)
		}
		for _, e := range got {
			if e.Start+e.Duration > tt.length+1e-9 {
				t.Errorf("%s: note ends after length: %v", tt.name, e)
			}
		}
	}
}

// talea 5 against color 7 repeat after 35
func TestPeriodMath(t *testing.T) {
	talea := []float64{1, 2, 1, -1, 2} // 5 elements, 7 steps long
	color := []float64{60, 62, 64, 65, 67, 69, 71}
	if Period(talea, color) != 35 {
		t.Fatal(Period(talea, color))
	}
	var sum float64
	for _, d := range talea {
		sum += max(d, -d)
	}
	cycle := sum * 7 // 35 elements = 7 talea cycles
	ev := Events(talea, color, 2*cycle)
	half := Events(talea, color, cycle)
	// the second period repeats the first, shifted by cycle
	var second []note.Event
	for _, e := range ev {
		if e.Start >= cycle {
			e.Start -= cycle
			second = append(second, e)
		}
	}
	if !slices.Equal(half, second) {
		t.Error("the pattern does not repeat after 35 elements")
	}
	// and not earlier: after 5 elements (one talea) the colors differ
	first := Events(talea, color, sum)
	next := Events(talea, color, 2*sum)[len(first):]
	if first[0].Pitch == next[0].Pitch {
		t.Error("color repeated with the talea")
	}
	if Period(nil, color) != 0 || Period(talea, nil) != 0 {
		t.Error("empty period")
	}
}

func TestSteps(t *testing.T) {
	for _, bad := range [][2][]float64{{nil, {60}}, {{1}, nil}, {{0}, {60}}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("expected panic", bad)
				}
			}()
			New(bad[0], bad[1])
		}()
	}
	s := New([]float64{2, -2, 1}, []float64{60, 62})
	p, l := s.Pitch(), s.Length()
	var got []float64
	var lens []float64
	for range 8 {
		o := p.Generate()
		if o.Played() {
			got = append(got, o.Value)
		} else {
			got = append(got, -1)
		}
		lens = append(lens, l.Generate())
	}
	if want := []float64{60, -1, -1, -1, 60, 62, -1, -1}; !slices.Equal(got, want) {
		t.Errorf("pitches %v want %v", got, want)
	}
	if want := []float64{2, 1, 2, 1, 1, 2, 1, 2}; !slices.Equal(lens, want) {
		t.Errorf("lengths %v want %v", lens, want)
	}
	p.Reset()
	if p.Generate().Value != 60 {
		t.Error("reset")
	}
	if !p.Continuous() || p.Done() {
		t.Error("continuous")
	}
}
