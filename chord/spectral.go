package chord

import (
	"math"
	"slices"
)

// A4 is the pitch of the MIDI note 69 in Hz, the reference of HzToMIDI and MIDIToHz
const A4 = 440.0

// HzToMIDI converts a frequency in Hz to a (fractional) MIDI pitch, 69 is A4 at 440 Hz. A frequency
// of 0 or less has no pitch and gives -Inf
func HzToMIDI(hz float64) float64 {
	if hz <= 0 {
		return math.Inf(-1)
	}
	return 69 + 12*math.Log2(hz/A4)
}

// MIDIToHz converts a (fractional) MIDI pitch to a frequency in Hz
func MIDIToHz(pitch float64) float64 { return A4 * math.Pow(2, (pitch-69)/12) }

// sortedPitches turns frequencies into a sorted chord without doubles, frequencies of 0 or less
// and non finite ones are left out
func sortedPitches(hz []float64) Chord {
	var c Chord
	for _, f := range hz {
		if f > 0 && !math.IsInf(f, 0) && !math.IsNaN(f) {
			c = append(c, HzToMIDI(f))
		}
	}
	return unique(c)
}

// unique sorts a chord and removes pitches closer than a thousandth of a semitone
func unique(c Chord) Chord {
	slices.Sort(c)
	return slices.CompactFunc(c, func(a, b float64) bool { return math.Abs(a-b) < 1e-3 })
}

// Harmonic builds a spectral chord from a fundamental: the partial n sounds at fundamental *
// n^distortion. A distortion of 1 gives the harmonic series, above 1 stretches the partials apart and
// below 1 squeezes them together (0 gives unisons). With subharmonic the fundamental is divided
// instead (fundamental / n^distortion), a series that goes down, as in Grisey's Partiels. The
// harmonics are partial numbers (1 is the fundamental), numbers below 1 are skipped. The pitches are
// fractional MIDI notes in ascending order, microtonal as the spectrum asks. It returns nil for a
// fundamental of 0 Hz or less
func Harmonic(fundamentalHz float64, harmonics []int, distortion float64, subharmonic bool) Chord {
	if fundamentalHz <= 0 || math.IsNaN(fundamentalHz) || math.IsNaN(distortion) {
		return nil
	}
	hz := make([]float64, 0, len(harmonics))
	for _, n := range harmonics {
		if n < 1 {
			continue
		}
		ratio := math.Pow(float64(n), distortion)
		if subharmonic {
			hz = append(hz, fundamentalHz/ratio)
		} else {
			hz = append(hz, fundamentalHz*ratio)
		}
	}
	return sortedPitches(hz)
}

// Multiply is chord multiplication: for every pitch of a it builds a chord on that pitch from the
// intervals of b (the distances of the pitches of b to its lowest one), and joins them. The result
// is sorted and has no doubles. Multiplying by a chord of one pitch gives a, and an empty chord
// gives an empty result
func Multiply(a, b Chord) Chord {
	if len(a) == 0 || len(b) == 0 {
		return nil
	}
	low := slices.Min(b)
	out := make(Chord, 0, len(a)*len(b))
	for _, p := range a {
		for _, q := range b {
			out = append(out, p+q-low)
		}
	}
	return unique(out)
}

// FMHz gives the frequencies of simple frequency modulation: the carrier and the sidebands
// carrier +- k*mod for k up to orders. Sidebands below 0 Hz fold back (a negative frequency sounds
// as its absolute value). The result is ascending without doubles
func FMHz(carrierHz, modHz float64, orders int) []float64 {
	orders = max(orders, 0)
	hz := []float64{math.Abs(carrierHz)}
	for k := 1; k <= orders; k++ {
		hz = append(hz, math.Abs(carrierHz+float64(k)*modHz), math.Abs(carrierHz-float64(k)*modHz))
	}
	slices.Sort(hz)
	return slices.CompactFunc(hz, func(a, b float64) bool { return math.Abs(a-b) < 1e-9 })
}

// FM is the spectrum of FMHz as a chord of fractional MIDI pitches. Components of 0 Hz have no pitch
// and are left out
func FM(carrierHz, modHz float64, orders int) Chord {
	return sortedPitches(FMHz(carrierHz, modHz, orders))
}

// RingMod is ring modulation of two frequencies: the sum and the difference tone, which replace
// the two inputs. The difference is the absolute value, equal inputs give only the sum
func RingMod(aHz, bHz float64) Chord {
	return sortedPitches([]float64{aHz + bHz, math.Abs(aHz - bHz)})
}
