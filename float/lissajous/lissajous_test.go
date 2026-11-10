package lissajous

import (
	"math"
	"testing"
)

func TestCurves(t *testing.T) {
	// equal frequencies without phase difference: a diagonal line
	l := New(1, 1, 0, 16)
	for range 16 {
		v := l.Generate()
		if v[0] != v[1] {
			t.Fatalf("%v", v)
		}
	}
	// a quarter phase difference: a circle
	c := New(1, 1, math.Pi/2, 32)
	for range 32 {
		v := c.Generate()
		if math.Abs(v[0]*v[0]+v[1]*v[1]-1) > 1e-12 {
			t.Fatalf("not on the unit circle: %v", v)
		}
	}
}

func TestPeriodAndFrame(t *testing.T) {
	l := New(3, 2, 0.3, 24)
	first := append([]float64{}, l.Generate()...)
	for range 23 {
		l.Generate()
	}
	again := l.Generate()
	if first[0] != again[0] || first[1] != again[1] {
		t.Error("not periodic")
	}
	if l.Dimensions() != 2 || !l.Continuous() || l.Done() {
		t.Error("flags")
	}
	l.Reset()
	if v := l.Generate(); v[0] != first[0] {
		t.Error("reset")
	}
}

func TestInvalid(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("0 steps should panic")
		}
	}()
	New(1, 1, 0, 0)
}
