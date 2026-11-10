package scale

import "math"

// Semitones converts frequency ratios to semitones (12 log2 ratio): 3/2 gives 7.0196, the just fifth
func Semitones(ratios ...float64) []float64 {
	out := make([]float64, len(ratios))
	for i, r := range ratios {
		out[i] = 12 * math.Log2(r)
	}
	return out
}

// Just intonation: the intervals are pure frequency ratios, so the thirds and fifths are slightly
// different from the equal tempered scales above (the major third 5/4 is 3.86 semitones, not 4)
var (
	JustMajor = Semitones(1, 9.0/8, 5.0/4, 4.0/3, 3.0/2, 5.0/3, 15.0/8)
	JustMinor = Semitones(1, 9.0/8, 6.0/5, 4.0/3, 3.0/2, 8.0/5, 9.0/5)
	// Pythagorean major is built from stacked fifths
	PythagoreanMajor = Semitones(1, 9.0/8, 81.0/64, 4.0/3, 3.0/2, 27.0/16, 243.0/128)
)

// FromRatios creates a scale from frequency ratios within one octave (the first is 1), for any just or
// other tuning from a list of ratios. It panics under the same conditions as New
func FromRatios(root float64, ratios ...float64) *Scale { return New(root, Semitones(ratios...)) }

// Equal creates a scale that divides the octave in equal steps: 12 is the normal scale, 24 gives
// quarter tones, 19 and 31 are popular too. The pitches are in (fractional) semitones, so a
// microtonal scale needs an output that can play them, see midi.Track.PitchBend
func Equal(root float64, divisions int) *Scale {
	if divisions < 1 {
		panic("scale: divisions must be at least 1")
	}
	steps := make([]float64, divisions)
	for i := range steps {
		steps[i] = 12 * float64(i) / float64(divisions)
	}
	return New(root, steps)
}
