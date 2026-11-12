package stochastic_test

import (
	"math"
	"slices"
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/constant"
	"github.com/almerlucke/genny/scale"
	"github.com/almerlucke/genny/stochastic"
)

func TestPoisson(t *testing.T) {
	genny.Seed(1)
	ts := stochastic.Poisson(2, 1000)
	if !slices.IsSorted(ts) {
		t.Error("onsets are not sorted")
	}
	for _, v := range ts {
		if v < 0 || v >= 1000 {
			t.Fatalf("onset %v outside [0, 1000)", v)
		}
	}
	if n := len(ts); n < 1800 || n > 2200 {
		t.Errorf("%d events for rate 2 over 1000 steps, want about 2000", n)
	}
}

func TestPoissonEdgeCases(t *testing.T) {
	for _, c := range []struct{ rate, length float64 }{{0, 10}, {-1, 10}, {5, 0}, {5, -3}} {
		if got := stochastic.Poisson(c.rate, c.length); len(got) != 0 {
			t.Errorf("Poisson(%v, %v) = %v, want none", c.rate, c.length, got)
		}
	}
	for _, c := range []struct{ rate, length float64 }{{math.NaN(), 1}, {1, math.Inf(1)}, {math.Inf(1), 1}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("Poisson(%v, %v) did not panic", c.rate, c.length)
				}
			}()
			stochastic.Poisson(c.rate, c.length)
		}()
	}
}

func TestPoissonInt(t *testing.T) {
	genny.Seed(2)
	for _, mean := range []float64{0.5, 4, 25, 100, 2000} {
		const n = 4000
		sum, sq := 0.0, 0.0
		for range n {
			v := float64(stochastic.PoissonInt(mean))
			sum += v
			sq += v * v
		}
		m := sum / n
		variance := sq/n - m*m
		if math.Abs(m-mean) > 0.1*mean+0.1 {
			t.Errorf("mean %v: got mean %v", mean, m)
		}
		if math.Abs(variance-mean) > 0.25*mean+0.2 {
			t.Errorf("mean %v: got variance %v", mean, variance)
		}
	}
	for _, mean := range []float64{0, -2, math.NaN()} {
		if got := stochastic.PoissonInt(mean); got != 0 {
			t.Errorf("PoissonInt(%v) = %d, want 0", mean, got)
		}
	}
}

func TestCloudDensityFollowsCurve(t *testing.T) {
	genny.Seed(3)
	c := stochastic.NewCloud(stochastic.Config{
		Density: func(step int) float64 {
			if step < 100 {
				return 0.5
			}
			return 5
		},
		Pitch: stochastic.Uniform(60, 72),
	})
	events := c.Render(200)
	first, second := 0, 0
	for i, e := range events {
		if i > 0 && e.Start < events[i-1].Start {
			t.Fatal("events not ordered")
		}
		if e.Pitch < 60 || e.Pitch >= 72 || e.Velocity != 0.8 || e.Duration != 1 {
			t.Fatalf("unexpected event %+v", e)
		}
		if e.Start < 100 {
			first++
		} else {
			second++
		}
	}
	if second < 4*first {
		t.Errorf("first half %d events, second half %d: the density did not rise", first, second)
	}
}

func TestCloudOptions(t *testing.T) {
	genny.Seed(4)
	key := scale.New(60, scale.WholeTone)
	c := stochastic.NewCloud(stochastic.Config{
		Density:  func(int) float64 { return 3 },
		Pitch:    stochastic.Uniform(72, 60), // reversed bounds
		Velocity: constant.New(2.0),
		Duration: constant.New(-1.0),
		Scale:    key,
	})
	for _, e := range c.Render(50) {
		if key.Quantize(e.Pitch) != e.Pitch {
			t.Fatalf("pitch %v not in scale", e.Pitch)
		}
		if e.Velocity != 1 || e.Duration != 1e-3 {
			t.Fatalf("velocity and duration not clamped: %+v", e)
		}
	}
	if got := c.Render(0); len(got) != 0 {
		t.Errorf("0 steps gave %d events", len(got))
	}
}

func TestNewCloudPanics(t *testing.T) {
	for name, cfg := range map[string]stochastic.Config{
		"no density": {Pitch: stochastic.Uniform(0, 1)},
		"no pitch":   {Density: func(int) float64 { return 1 }},
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: no panic", name)
				}
			}()
			stochastic.NewCloud(cfg)
		}()
	}
}

func TestDensityGate(t *testing.T) {
	genny.Seed(5)
	for _, c := range []struct {
		p        float64
		min, max int
	}{{0, 0, 0}, {-1, 0, 0}, {1, 1000, 1000}, {2, 1000, 1000}, {0.3, 240, 360}, {math.NaN(), 0, 0}} {
		g := stochastic.NewDensity(constant.New(c.p))
		n := 0
		for range 1000 {
			if g.Generate() {
				n++
			}
		}
		if n < c.min || n > c.max {
			t.Errorf("p=%v: %d hits of 1000, want %d..%d", c.p, n, c.min, c.max)
		}
	}
}

func TestFree(t *testing.T) {
	genny.Seed(6)
	key := scale.New(0, scale.Major)
	events := stochastic.Free(500, 1, 2, 60, 5, key)
	if len(events) < 400 || len(events) > 600 {
		t.Errorf("%d events, want about 500", len(events))
	}
	for _, e := range events {
		if key.Quantize(e.Pitch) != e.Pitch || e.Velocity < 0.3 || e.Velocity > 0.9 || e.Duration <= 0 || e.Start >= 500 {
			t.Fatalf("bad event %+v", e)
		}
	}
}
