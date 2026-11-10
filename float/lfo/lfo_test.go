package lfo

import (
	"math"
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/sequence"
)

const fs = 1000.0

func cycle(l *LFO) (out []float64) {
	for range int(fs) { // 1 Hz
		out = append(out, l.Generate())
	}
	return
}

func TestShapes(t *testing.T) {
	sine := cycle(New(Sine, 1, fs))
	if math.Abs(sine[0]) > 1e-9 || math.Abs(sine[250]-1) > 1e-9 || math.Abs(sine[750]+1) > 1e-9 {
		t.Error("sine")
	}
	tri := cycle(New(Triangle, 1, fs))
	if tri[0] != -1 || math.Abs(tri[500]-1) > 1e-9 || math.Abs(tri[250]) > 1e-9 {
		t.Errorf("triangle %v %v %v", tri[0], tri[250], tri[500])
	}
	saw := cycle(New(Saw, 1, fs))
	if saw[0] != -1 || saw[500] > 0.01 || saw[999] < 0.99 {
		t.Error("saw")
	}
	sq := New(Square, 1, fs)
	sq.SetPulseWidth(0.25)
	hi := 0
	for _, v := range cycle(sq) {
		if v == 1 {
			hi++
		} else if v != -1 {
			t.Fatal(v)
		}
	}
	if hi != 250 {
		t.Errorf("pulse width: %d high", hi)
	}
}

func TestUnipolarRange(t *testing.T) {
	genny.Seed(1)
	for _, sh := range []Shape{Sine, Triangle, Saw, Square, SampleHold, SmoothRandom} {
		l := New(sh, 3, fs).SetUnipolar(true)
		for range 5000 {
			if v := l.Generate(); v < 0 || v > 1 {
				t.Fatalf("shape %d: %v", sh, v)
			}
		}
	}
}

func TestSampleHoldHoldsPerCycle(t *testing.T) {
	genny.Seed(2)
	l := New(SampleHold, 1, fs)
	values := map[float64]bool{}
	for range 5 {
		c := cycle(l)
		for _, v := range c {
			if v != c[0] {
				t.Fatal("value changed within a cycle")
			}
		}
		values[c[0]] = true
	}
	if len(values) < 4 {
		t.Error("no new values per cycle")
	}
}

func TestSmoothRandomHasNoJumps(t *testing.T) {
	genny.Seed(2)
	l := New(SmoothRandom, 2, fs)
	prev := l.Generate()
	for range 10000 {
		v := l.Generate()
		if math.Abs(v-prev) > 0.02 {
			t.Fatalf("jump %v", math.Abs(v-prev))
		}
		prev = v
	}
}

func TestSlew(t *testing.T) {
	s := NewSlew(sequence.NewLoop(0.0, 0.0, 1.0, 1.0, 1.0, 1.0, 0.0, 0.0, 0.0, 0.0), 0.5, 0.25)
	want := []float64{0, 0, 0.5, 1, 1, 1, 0.75, 0.5, 0.25, 0}
	for i, w := range want {
		if got := s.Generate(); math.Abs(got-w) > 1e-12 {
			t.Fatalf("step %d: %v, want %v", i, got, w)
		}
	}
}
