package scale

import (
	"math"
	"slices"
	"testing"
)

func near(a, b float64) bool { return math.Abs(a-b) < 1e-3 }

func TestSemitonesAndJust(t *testing.T) {
	if got := Semitones(2, 3.0/2); !near(got[0], 12) || !near(got[1], 7.0196) {
		t.Errorf("%v", got)
	}
	want := []float64{0, 2.0391, 3.8631, 4.9804, 7.0196, 8.8436, 10.8826}
	for i, w := range want {
		if !near(JustMajor[i], w) {
			t.Errorf("just major step %d: %v, want %v", i, JustMajor[i], w)
		}
	}
	if len(JustMinor) != 7 || !near(JustMinor[2], 3.1564) || len(PythagoreanMajor) != 7 || !near(PythagoreanMajor[2], 4.0782) {
		t.Errorf("%v %v", JustMinor, PythagoreanMajor)
	}
	s := FromRatios(60, 1, 5.0/4, 3.0/2)
	if !near(s.Degree(1), 63.8631) || !near(s.Degree(3), 72) {
		t.Errorf("%v %v", s.Degree(1), s.Degree(3))
	}
}

func TestEqual(t *testing.T) {
	q := Equal(60, 24) // quarter tones
	if q.Len() != 24 || q.Degree(1) != 60.5 || q.Degree(24) != 72 || q.Quantize(60.3) != 60.5 {
		t.Errorf("%v %v", q.Degree(1), q.Quantize(60.3))
	}
	for i, st := range Equal(0, 12).steps {
		if st != Chromatic[i] {
			t.Fatal("12 equal steps are the chromatic scale")
		}
	}
	if !slices.Equal(Equal(0, 1).steps, []float64{0}) {
		t.Error("one division")
	}
}

func TestInvalidScales(t *testing.T) {
	for name, f := range map[string]func(){
		"no steps":      func() { New(0, nil) },
		"descending":    func() { New(0, []float64{0, 4, 2}) },
		"repeated":      func() { New(0, []float64{0, 2, 2}) },
		"beyond period": func() { New(0, []float64{0, 12}) },
		"negative step": func() { New(0, []float64{-1, 2}) },
		"zero period":   func() { NewPeriod(0, []float64{0}, 0) },
		"equal 0":       func() { Equal(0, 0) },
		"walk lo > hi":  func() { NewWalk(New(0, Major), 3, 5, 1, 1, 0, 1) },
		"walk start":    func() { NewWalk(New(0, Major), 9, 0, 7, 1, 0, 1) },
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s should panic", name)
				}
			}()
			f()
		}()
	}
}
