package euclid

import (
	"slices"
	"testing"
)

func str(p []bool) string {
	b := make([]byte, len(p))
	for i, v := range p {
		b[i] = '.'
		if v {
			b[i] = 'x'
		}
	}
	return string(b)
}

func TestKnownRhythms(t *testing.T) {
	for _, c := range []struct {
		steps, pulses int
		want          string
	}{
		{8, 3, "x..x..x."}, // tresillo
		{8, 5, "x.xx.xx."},
		{4, 4, "xxxx"},
		{5, 0, "....."},
		{16, 4, "x...x...x...x..."},
		{7, 3, "x.x.x.."},
		{3, 9, "xxx"}, // pulses clamp
	} {
		if got := str(Pattern(c.steps, c.pulses, 0)); got != c.want {
			t.Errorf("E(%d,%d) = %s, want %s", c.pulses, c.steps, got, c.want)
		}
	}
}

func TestPulseCountAndRotation(t *testing.T) {
	for steps := 1; steps <= 24; steps++ {
		for pulses := 0; pulses <= steps; pulses++ {
			p := Pattern(steps, pulses, 0)
			n := 0
			for _, v := range p {
				if v {
					n++
				}
			}
			if len(p) != steps || n != pulses {
				t.Fatalf("E(%d,%d): %d steps %d pulses", pulses, steps, len(p), n)
			}
		}
	}
	if got := str(Pattern(8, 3, 1)); got != "..x..x.x" {
		t.Errorf("rotation 1 = %s", got)
	}
	if !slices.Equal(Pattern(8, 3, 8), Pattern(8, 3, 0)) || !slices.Equal(Pattern(8, 3, -1), Pattern(8, 3, 7)) {
		t.Error("rotation does not wrap")
	}
}

func TestGeneratorLoops(t *testing.T) {
	e := New(4, 1, 0)
	var got []bool
	for range 8 {
		got = append(got, e.Generate())
	}
	if str(got) != "x...x..." || !e.Continuous() {
		t.Errorf("got %s", str(got))
	}
}
