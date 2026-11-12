package melody

import (
	"math"
	"slices"
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/chord"
	"github.com/almerlucke/genny/note"
	"github.com/almerlucke/genny/scale"
	"github.com/almerlucke/genny/sequence"
)

var (
	cMajor = scale.New(60, scale.Major)
	prog   = []chord.Chord{
		chord.Diatonic(cMajor, 0, 3), chord.Diatonic(cMajor, 3, 3),
		chord.Diatonic(cMajor, 4, 3), chord.Diatonic(cMajor, 0, 3),
	}
)

func inChord(c chord.Chord, p float64) bool {
	for _, q := range c {
		if math.Mod(math.Mod(q-p, 12)+12, 12) == 0 {
			return true
		}
	}
	return false
}

func TestChordTones(t *testing.T) {
	for seed := uint64(1); seed <= 200; seed++ {
		genny.Seed(seed)
		cfg := Config{Scale: cMajor}
		ev := ChordTonesList(prog, 8, cfg)
		if len(ev) != 8*16 {
			t.Fatalf("%d events", len(ev))
		}
		for i, e := range ev {
			bar, step := i/16, i%16
			c := prog[bar%len(prog)]
			if (Weight(step, 16, 4) >= 2 || step == 15) && !inChord(c, e.Pitch) {
				t.Errorf("seed %d bar %d step %d: %v is not a chord tone of %v", seed, bar, step, e.Pitch, c)
			}
			if e.Pitch < 60 || e.Pitch > 84 {
				t.Errorf("out of range: %v", e.Pitch)
			}
			if e.Start != float64(i) || e.Duration != 1 {
				t.Errorf("timing %v", e)
			}
			if i > 0 {
				d := cMajor.DegreeOf(e.Pitch) - cMajor.DegreeOf(ev[i-1].Pitch)
				if d < 0 {
					d = -d
				}
				// leaps are limited, except into a chord tone when nothing is nearer
				if d > 3 {
					t.Errorf("seed %d step %d: leap of %d degrees", seed, i, d)
				}
			}
		}
		// weak steps are scale tones
		for i, e := range ev {
			if Weight(i%16, 16, 4) == 0 && i%16 != 15 && cMajor.Quantize(e.Pitch) != e.Pitch {
				t.Errorf("passing tone %v not in the scale", e.Pitch)
			}
		}
	}
}

func TestGeneratorForm(t *testing.T) {
	genny.Seed(4)
	a := ChordTones(sequence.NewLoop(prog...), 6, Config{Scale: cMajor, BarSteps: 8})
	genny.Seed(4)
	b := ChordTonesList(prog, 6, Config{Scale: cMajor, BarSteps: 8})
	if !slices.Equal(a, b) || len(a) != 48 {
		t.Error("generator and list form differ", len(a), len(b))
	}
	// a generator that ends stops the melody
	short := sequence.New(prog[0], prog[1])
	if got := ChordTones(short, 10, Config{Scale: cMajor, BarSteps: 8}); len(got) != 16 {
		t.Errorf("ended generator gave %d events", len(got))
	}
}

func TestEdgeCases(t *testing.T) {
	cfg := Config{Scale: cMajor}
	if ChordTonesList(nil, 4, cfg) != nil {
		t.Error("no chords")
	}
	if len(ChordTonesList(prog, 0, cfg)) != 0 || len(ChordTonesList(prog, -3, cfg)) != 0 {
		t.Error("no bars")
	}
	// an empty chord is a bar of rest
	got := ChordTonesList([]chord.Chord{{}, prog[0]}, 2, cfg)
	if len(got) != 16 || got[0].Start != 16 {
		t.Error("empty chord", len(got))
	}
	// odd meter and sizes
	got = ChordTonesList(prog, 3, Config{Scale: cMajor, BarSteps: 7, BeatSteps: 3, MaxLeap: 1})
	if len(got) != 21 {
		t.Error(len(got))
	}
	for _, bad := range []Config{
		{},
		{Scale: cMajor, BarSteps: -1},
		{Scale: cMajor, Low: 60, High: 66},
		{Scale: cMajor, MaxLeap: -2},
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("expected panic for %+v", bad)
				}
			}()
			ChordTonesList(prog, 1, bad)
		}()
	}
}

func TestWeight(t *testing.T) {
	var got []int
	for s := range 8 {
		got = append(got, Weight(s, 8, 4))
	}
	if want := []int{3, 0, 1, 0, 2, 0, 1, 0}; !slices.Equal(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
	// velocity follows the weight
	genny.Seed(1)
	ev := ChordTonesList(prog, 1, Config{Scale: cMajor})
	if ev[0].Velocity <= ev[4].Velocity || ev[4].Velocity <= ev[1].Velocity {
		t.Error("velocity", ev[0].Velocity, ev[4].Velocity, ev[1].Velocity)
	}
}

var _ = []note.Event(nil)
