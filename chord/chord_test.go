package chord

import (
	"slices"
	"testing"

	"github.com/almerlucke/genny/scale"
)

func TestBuilders(t *testing.T) {
	c := New(60, Dominant7)
	if !slices.Equal(c, Chord{60, 64, 67, 70}) {
		t.Fatal(c)
	}
	if got := c.Invert(1); !slices.Equal(got, Chord{64, 67, 70, 72}) {
		t.Errorf("inv 1 %v", got)
	}
	if got := c.Invert(-1); !slices.Equal(got, Chord{58, 60, 64, 67}) {
		t.Errorf("inv -1 %v", got)
	}
	if !slices.Equal(c, Chord{60, 64, 67, 70}) {
		t.Error("invert changed the chord")
	}
	if got := New(60, Major).Open(); !slices.Equal(got, Chord{52, 60, 67}) {
		t.Errorf("open %v", got)
	}
}

func TestDiatonic(t *testing.T) {
	s := scale.New(60, scale.Major)
	for deg, want := range map[int]Chord{0: {60, 64, 67}, 1: {62, 65, 69}, 4: {67, 71, 74}, 6: {71, 74, 77}} {
		if got := Diatonic(s, deg, 3); !slices.Equal(got, want) {
			t.Errorf("degree %d: %v, want %v", deg, got, want)
		}
	}
	if got := Diatonic(s, 4, 4); !slices.Equal(got, Chord{67, 71, 74, 77}) {
		t.Errorf("V7 %v", got)
	}
}

func TestVoiceLeader(t *testing.T) {
	s := scale.New(60, scale.Major)
	v := NewVoiceLeader(48, 84)
	var total float64
	var prev []float64
	for _, deg := range []int{0, 3, 4, 0} {
		out := v.Transform(Diatonic(s, deg, 3))
		if prev != nil {
			for i := range out {
				d := out[i] - prev[i]
				if d < 0 {
					d = -d
				}
				total += d
			}
		}
		prev = out
		if out[0] < 48 || out[len(out)-1] > 84 {
			t.Fatalf("out of range %v", out)
		}
	}
	// root position chords I IV V I jump 5+5+2+7... close voicings must move far less
	if total > 14 {
		t.Errorf("voices moved %v semitones", total)
	}
	v.Reset()
	if got := v.Transform([]float64{64, 60, 67}); !slices.Equal(got, []float64{60, 64, 67}) {
		t.Errorf("first chord %v", got)
	}
}
