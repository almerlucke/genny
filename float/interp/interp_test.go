package interp_test

import (
	"math"
	"testing"

	"github.com/almerlucke/genny/float/interp"
)

// frames is a FrameGenerator over fixed frames, it keeps returning the last frame when exhausted
type frames struct {
	data  [][]float64
	index int
	dims  int
}

func (f *frames) Generate() []float64 {
	i := f.index
	if i >= len(f.data) {
		i = len(f.data) - 1
	} else {
		f.index++
	}
	return f.data[i]
}
func (f *frames) Dimensions() int  { return f.dims }
func (f *frames) Continuous() bool { return false }
func (f *frames) Done() bool       { return f.index >= len(f.data) }
func (f *frames) Reset()           { f.index = 0 }

func ramp(n int) *frames {
	f := &frames{dims: 1}
	for i := 0; i < n; i++ {
		f.data = append(f.data, []float64{float64(i)})
	}
	return f
}

func near(a, b float64) bool { return math.Abs(a-b) < 1e-12 }

func TestLinear(t *testing.T) {
	ip := interp.New(ramp(4), interp.Linear, 0.5)
	if ip.Dimensions() != 1 || ip.Continuous() {
		t.Fatal("dimensions / continuous")
	}
	want := []float64{0, 0.5, 1, 1.5, 2, 2.5}
	for i, w := range want {
		if ip.Done() {
			t.Fatalf("done too early at %d", i)
		}
		if got := ip.Generate()[0]; !near(got, w) {
			t.Fatalf("value %d: got %v want %v", i, got, w)
		}
	}
	if !ip.Done() {
		t.Fatal("done once the source is exhausted")
	}
	ip.Reset()
	if ip.Done() || !near(ip.Generate()[0], 0) || !near(ip.Generate()[0], 0.5) {
		t.Fatal("reset restarts")
	}
}

func TestStepOfOneReadsTheSource(t *testing.T) {
	for _, m := range []interp.Method{interp.Linear, interp.Cubic} {
		ip := interp.New(ramp(8), m, 1.0)
		for i := 0; i < 6; i++ {
			if got := ip.Generate()[0]; !near(got, float64(i)) {
				t.Fatalf("method %d value %d: got %v", m, i, got)
			}
		}
	}
}

func TestCubicFollowsALine(t *testing.T) {
	// the interpolation of a straight line must be that line (the first segment is skipped:
	// it has no real neighbour on the left)
	ip := interp.New(ramp(10), interp.Cubic, 0.25)
	for i := 0; i < 20; i++ {
		got := ip.Generate()[0]
		if i < 4 {
			continue
		}
		want := float64(i) * 0.25
		if !near(got, want) {
			t.Fatalf("value %d: got %v want %v", i, got, want)
		}
	}
}

func TestMultipleDimensionsAndSetDelta(t *testing.T) {
	f := &frames{dims: 2, data: [][]float64{{0, 10}, {2, 20}, {4, 30}, {6, 40}}}
	ip := interp.New(f, interp.Linear, 0.5)
	if got := ip.Generate(); !near(got[0], 0) || !near(got[1], 10) {
		t.Fatalf("got %v", got)
	}
	if got := ip.Generate(); !near(got[0], 1) || !near(got[1], 15) {
		t.Fatalf("got %v", got)
	}
	ip.SetDelta(0.25)
	if got := ip.Generate(); !near(got[0], 2) || !near(got[1], 20) {
		t.Fatalf("got %v", got)
	}
	if got := ip.Generate(); !near(got[0], 2.5) || !near(got[1], 22.5) {
		t.Fatalf("got %v", got)
	}
}
