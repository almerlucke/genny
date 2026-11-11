package chord

import (
	"math"
	"slices"
	"testing"
)

func near(a, b Chord, eps float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if math.Abs(a[i]-b[i]) > eps {
			return false
		}
	}
	return true
}

func TestHzMIDI(t *testing.T) {
	tests := []struct {
		hz, midi float64
	}{{440, 69}, {880, 81}, {220, 57}, {261.6255653005986, 60}}
	for _, tt := range tests {
		if got := HzToMIDI(tt.hz); math.Abs(got-tt.midi) > 1e-9 {
			t.Errorf("HzToMIDI(%v) = %v", tt.hz, got)
		}
		if got := MIDIToHz(tt.midi); math.Abs(got-tt.hz) > 1e-9 {
			t.Errorf("MIDIToHz(%v) = %v", tt.midi, got)
		}
	}
	if !math.IsInf(HzToMIDI(0), -1) || !math.IsInf(HzToMIDI(-5), -1) {
		t.Error("no pitch below 0 Hz")
	}
}

func TestHarmonic(t *testing.T) {
	tests := []struct {
		name string
		f    float64
		n    []int
		dist float64
		sub  bool
		want Chord
	}{
		{"octaves", 110, []int{1, 2, 4, 8}, 1, false, Chord{45, 57, 69, 81}},
		{"unsorted input", 110, []int{4, 1, 2}, 1, false, Chord{45, 57, 69}},
		{"skips bad numbers", 110, []int{0, -3, 2}, 1, false, Chord{57}},
		{"subharmonic", 440, []int{1, 2, 4}, 1, true, Chord{45, 57, 69}},
		{"stretched", 110, []int{1, 2}, 2, false, Chord{45, 45 + 24}},
		{"squeezed to unison", 110, []int{1, 2, 3}, 0, false, Chord{45}},
	}
	for _, tt := range tests {
		got := Harmonic(tt.f, tt.n, tt.dist, tt.sub)
		if !near(got, tt.want, 1e-9) {
			t.Errorf("%s: %v want %v", tt.name, got, tt.want)
		}
	}
	// the third harmonic is a pure fifth above the octave: not on the piano
	c := Harmonic(110, []int{3}, 1, false)
	if frac := c[0] - math.Floor(c[0]); math.Abs(frac-0.02) > 0.01 {
		t.Errorf("third harmonic %v", c)
	}
}

func TestHarmonicEdges(t *testing.T) {
	if Harmonic(0, []int{1}, 1, false) != nil || Harmonic(-1, []int{1}, 1, true) != nil {
		t.Error("no fundamental")
	}
	if Harmonic(100, nil, 1, false) != nil && len(Harmonic(100, nil, 1, false)) != 0 {
		t.Error("no harmonics")
	}
	if Harmonic(100, []int{1}, math.NaN(), false) != nil {
		t.Error("NaN distortion")
	}
}

func TestMultiply(t *testing.T) {
	tests := []struct {
		name string
		a, b Chord
		want Chord
	}{
		{"major on fifth", Chord{60, 67}, Chord{0, 4, 7}, Chord{60, 64, 67, 71, 74}},
		{"b not at zero", Chord{60}, Chord{50, 54}, Chord{60, 64}},
		{"single pitch", Chord{60, 64}, Chord{12}, Chord{60, 64}},
		{"empty a", nil, Chord{0, 4}, nil},
		{"empty b", Chord{60}, nil, nil},
	}
	for _, tt := range tests {
		if got := Multiply(tt.a, tt.b); !near(got, tt.want, 1e-9) {
			t.Errorf("%s: %v want %v", tt.name, got, tt.want)
		}
	}
}

func TestFMAndRingMod(t *testing.T) {
	// carrier 400, mod 100, 2 orders: 200 300 400 500 600
	hz := FMHz(400, 100, 2)
	if !slices.Equal(hz, []float64{200, 300, 400, 500, 600}) {
		t.Errorf("%v", hz)
	}
	// folds back, and 0 Hz has no pitch
	if hz := FMHz(100, 150, 1); !slices.Equal(hz, []float64{50, 100, 250}) {
		t.Errorf("%v", hz)
	}
	if c := FM(100, 100, 1); len(c) != 2 { // 0, 100, 200
		t.Errorf("%v", c)
	}
	if got := FM(440, 0, 3); !near(got, Chord{69}, 1e-9) {
		t.Errorf("%v", got)
	}
	if got := FMHz(440, 10, -2); !slices.Equal(got, []float64{440}) {
		t.Errorf("negative orders: %v", got)
	}
	// ring modulation of 440 and 660: 220 and 1100
	want := Chord{HzToMIDI(220), HzToMIDI(1100)}
	if got := RingMod(440, 660); !near(got, want, 1e-9) {
		t.Errorf("%v want %v", got, want)
	}
	if got := RingMod(440, 440); len(got) != 1 {
		t.Errorf("equal inputs: %v", got)
	}
}
