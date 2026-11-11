package constrain_test

import (
	"slices"
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/constrain"
)

func TestScaleIntervals(t *testing.T) {
	cMajor := []int{60, 62, 64, 65, 67, 69, 71, 72, 74}
	tests := []struct {
		name      string
		first     int
		intervals []int
		allowed   []int
		opts      []constrain.Option
		want      []int
	}{
		{"cycle", 60, []int{2, 2, 1}, cMajor, nil, []int{60, 62, 64, 65, 67, 69, 71, 72, 74}},
		{"skips what is not allowed", 60, []int{1, 2}, cMajor, nil, []int{60, 62, 64, 65, 67, 69, 71, 72}},
		{"all allowed when empty", 10, []int{3, -1}, nil, nil, []int{10, 13, 12, 15, 14}},
		{"stays when nothing fits", 60, []int{1}, []int{60}, []constrain.Option{constrain.Stop(3)}, []int{60, 60, 60}},
		{"pitch classes", 60, []int{7}, []int{0, 7, 2, 9}, []constrain.Option{constrain.PitchClass()},
			[]int{60, 67, 74, 81}},
	}
	for _, tt := range tests {
		g := constrain.ScaleIntervals(tt.first, tt.intervals, tt.allowed, tt.opts...)
		got := take[int](g, len(tt.want))
		if !slices.Equal(got, tt.want) {
			t.Errorf("%s: %v want %v", tt.name, got, tt.want)
		}
	}
}

func TestScaleIntervalsStaysAllowed(t *testing.T) {
	allowed := []int{57, 59, 60, 62, 64, 65, 67, 69, 71, 72, 74, 76}
	g := constrain.ScaleIntervals(57, []int{2, 3, -1, 4, 1}, allowed)
	for i, p := range take[int](g, 200)[1:] {
		if !slices.Contains(allowed, p) {
			t.Fatalf("note %d: %d not allowed", i+1, p)
		}
	}
}

func TestScaleIntervalsEmptyAndReset(t *testing.T) {
	g := constrain.ScaleIntervals(60, nil, []int{60})
	if !g.Done() || g.Continuous() || g.Generate() != 0 {
		t.Error("no intervals: done at once")
	}
	g = constrain.ScaleIntervals(60, []int{1, 2}, nil)
	a := take[int](g, 5)
	g.Reset()
	if b := take[int](g, 5); !slices.Equal(a, b) || g.Done() || !g.Continuous() {
		t.Errorf("%v %v", a, b)
	}
}

func TestRandomIntervals(t *testing.T) {
	genny.Seed(5)
	g := constrain.RandomIntervals(48, 60, 1, 2, 5, 7)
	got := take[int](g, 500)
	if got[0] != 54 {
		t.Errorf("starts in the middle: %d", got[0])
	}
	moved := map[int]bool{}
	for i, p := range got {
		if p < 48 || p > 60 {
			t.Fatalf("note %d: %d out of range", i, p)
		}
		if i > 0 {
			d := p - got[i-1]
			if d < 0 {
				d = -d
			}
			moved[d] = true
			if d != 0 && d != 1 && d != 2 && d != 5 && d != 7 {
				t.Fatalf("note %d: interval %d", i, d)
			}
		}
	}
	if !moved[1] || !moved[7] {
		t.Errorf("did not use the intervals: %v", moved)
	}
}

func TestRandomIntervalsTightRange(t *testing.T) {
	genny.Seed(6)
	// no interval fits a range of one note: stays
	got := take[int](constrain.RandomIntervals(5, 5, 3), 10)
	if !slices.Equal(got, slices.Repeat([]int{5}, 10)) {
		t.Errorf("%v", got)
	}
	// reflecting: from the bottom of the range it can only go up
	g := constrain.RandomIntervals(0, 2, 2)
	for _, p := range take[int](g, 50) {
		if p != 0 && p != 2 && p != 1 {
			t.Fatalf("%d", p)
		}
	}
}
