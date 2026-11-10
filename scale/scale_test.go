package scale

import (
	"slices"
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/float/shape"
	"github.com/almerlucke/genny/sequence"
)

func TestDegree(t *testing.T) {
	s := New(60, Major)
	for deg, want := range map[int]float64{0: 60, 1: 62, 6: 71, 7: 72, 9: 76, -1: 59, -7: 48, -8: 47} {
		if got := s.Degree(deg); got != want {
			t.Errorf("degree %d = %v, want %v", deg, got, want)
		}
	}
}

func TestQuantize(t *testing.T) {
	s := New(60, MajorPentatonic) // 60 62 64 67 69 72
	for in, want := range map[float64]float64{
		60: 60, 60.4: 60, 61: 60, 61.2: 62, 65: 64, 65.6: 67, 65.5: 64, 70.9: 72, 71.9: 72, 58: 57, 12: 12, 100: 100,
	} {
		if got := s.Quantize(in); got != want {
			t.Errorf("quantize %v = %v, want %v", in, got, want)
		}
	}
	// every output is a scale pitch
	for x := -30.0; x < 130; x += 0.37 {
		q := s.Quantize(x)
		if s.Degree(s.DegreeOf(q)) != q {
			t.Fatalf("%v -> %v is not on the scale", x, q)
		}
	}
}

func TestShaperAndDegrees(t *testing.T) {
	s := New(0, Minor)
	g := shape.NewSingle(sequence.NewLoop(1.0, 4.0, 6.0), s)
	got := []float64{g.Generate(), g.Generate(), g.Generate()}
	if !slices.Equal(got, []float64{0, 3, 5}) { // ties go down
		t.Errorf("shaped %v", got)
	}
	d := NewDegrees(s, sequence.New(0, 2, 7))
	if got := []float64{d.Generate(), d.Generate(), d.Generate()}; !slices.Equal(got, []float64{0, 3, 12}) {
		t.Errorf("degrees %v", got)
	}
}

func TestWalkStaysInRangeAndLeaps(t *testing.T) {
	genny.Seed(3)
	s := New(48, Major)
	w := NewWalk(s, 7, 0, 14, 1, 0.2, 4)
	leaps := 0
	prev := w.Degree()
	for range 5000 {
		p := w.Generate()
		d := w.Degree()
		if d < 0 || d > 14 || p != s.Degree(d) {
			t.Fatalf("degree %d pitch %v", d, p)
		}
		if diff := d - prev; diff > 1 || diff < -1 {
			leaps++
		}
		prev = d
	}
	if leaps == 0 || leaps > 2500 {
		t.Errorf("%d leaps", leaps)
	}
	w.Reset()
	if w.Degree() != 7 {
		t.Error("reset")
	}
}
