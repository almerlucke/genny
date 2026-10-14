package adsr

import (
	"math"
	"testing"
)

// run generates n samples, the last one
func run(a *ADSR, n int) (v float64) {
	for range n {
		v = a.Generate()
	}
	return
}

// a held note follows a new sustain level, gliding (no jump); its release still ends it
func TestSustainFollowsTheSetting(t *testing.T) {
	const sr = 44100.0
	for _, mode := range []ReleaseMode{NoteOff, Duration} {
		s := NewSetting(1, 5, 0.9, 10, 0, 50)
		a := New(s, mode, sr)
		a.TriggerFull(10000, 1, s, mode) // 10 s: held for the test
		if v := run(a, int(0.1*sr)); math.Abs(v-0.9) > 1e-9 {
			t.Fatalf("mode %d: sustaining at %v, want 0.9", mode, v)
		}
		s.DecayLevel = 0
		first := a.Generate()
		if first < 0.8 {
			t.Errorf("mode %d: a jump to %v, want a glide", mode, first)
		}
		if v := run(a, int(0.05*sr)); v > 1e-3 {
			t.Errorf("mode %d: 50 ms after the sustain went to 0 it is at %v", mode, v)
		}
		s.DecayLevel = 0.5
		if v := run(a, int(0.05*sr)); math.Abs(v-0.5) > 1e-3 {
			t.Errorf("mode %d: at %v, want the new sustain 0.5", mode, v)
		}
		a.Release()
		run(a, int(0.06*sr))
		if !a.Done() {
			t.Errorf("mode %d: not done after its release", mode)
		}
	}
}

// a note of a duration lets go after it, at the level it sustains at
func TestDurationReleases(t *testing.T) {
	const sr = 44100.0
	s := NewSetting(1, 5, 0.7, 10, 0, 20)
	a := New(s, Duration, sr)
	a.TriggerFull(100, 1, s, Duration)
	run(a, int(0.1*sr)+int(0.02*sr)+2)
	if !a.Done() {
		t.Error("a note of 100 ms with a release of 20 ms is not done after 120 ms")
	}
}
