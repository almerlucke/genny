package phasor

import (
	"reflect"
	"testing"
)

func take(p *Phasor, n int) []float64 {
	out := make([]float64, n)
	for i := range out {
		out[i] = p.Generate()
	}
	return out
}

func TestPhasor(t *testing.T) {
	p := New(1, 4, 0)
	if !p.Continuous() || p.Done() {
		t.Fatal("continuous, never done")
	}
	if got := take(p, 6); !reflect.DeepEqual(got, []float64{0, 0.25, 0.5, 0.75, 0, 0.25}) {
		t.Fatalf("got %v", got)
	}
	p.Reset()
	if got := p.Generate(); got != 0 {
		t.Fatalf("reset goes to the start phase, got %v", got)
	}
}

func TestPhasorStartPhaseAndSetters(t *testing.T) {
	p := New(1, 4, 0.5)
	if got := take(p, 3); !reflect.DeepEqual(got, []float64{0.5, 0.75, 0}) {
		t.Fatalf("got %v", got)
	}
	p.Reset()
	p.SetFrequency(2, 4) // 0.5 per value
	if got := take(p, 3); !reflect.DeepEqual(got, []float64{0.5, 0, 0.5}) {
		t.Fatalf("got %v", got)
	}
	p.SetPhase(0.25)
	if got := p.Generate(); got != 0.25 {
		t.Fatalf("got %v", got)
	}
	p.Reset()
	if got := p.Generate(); got != 0.25 {
		t.Fatalf("SetPhase also sets the reset phase, got %v", got)
	}
}

func TestPhasorNegativeFrequencyWraps(t *testing.T) {
	p := New(-1, 4, 0)
	if got := take(p, 5); !reflect.DeepEqual(got, []float64{0, 0.75, 0.5, 0.25, 0}) {
		t.Fatalf("got %v", got)
	}
}

func TestPhaseOffset(t *testing.T) {
	p := New(1, 4, 0)
	p.PhaseOffset(0.75)
	p.PhaseOffset(0.5) // 1.25 wraps to 0.25
	if got := p.Generate(); got != 0.25 {
		t.Fatalf("got %v", got)
	}
	p.PhaseOffset(-0.5) // 0.5 - 0.5 = 0
	if got := p.Generate(); got != 0 {
		t.Fatalf("got %v", got)
	}
	p.PhaseOffset(-0.5) // 0.25 - 0.5 wraps to 0.75
	if got := p.Generate(); got != 0.75 {
		t.Fatalf("negative offset wraps, got %v", got)
	}
}
