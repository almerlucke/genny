package scale

import "math"

// Messiaen's modes of limited transposition, as steps in semitones above the root. Each is built from
// a group of intervals that repeats within the octave, so after a few transpositions it maps onto
// itself: the pitch set of the mode transposed up by that many semitones is the same set. Transpositions
// tells how many different versions there are
var (
	// Mode1 is the whole tone scale (2 transpositions)
	Mode1 = []float64{0, 2, 4, 6, 8, 10}
	// Mode2 alternates semitone and tone, the octatonic scale (3 transpositions)
	Mode2 = []float64{0, 1, 3, 4, 6, 7, 9, 10}
	// Mode3 repeats tone, semitone, semitone (4 transpositions)
	Mode3 = []float64{0, 2, 3, 4, 6, 7, 8, 10, 11}
	// Mode4 repeats semitone, semitone, minor third, semitone (6 transpositions)
	Mode4 = []float64{0, 1, 2, 5, 6, 7, 8, 11}
	// Mode5 repeats semitone, major third, semitone (6 transpositions)
	Mode5 = []float64{0, 1, 5, 6, 7, 11}
	// Mode6 repeats tone, tone, semitone, semitone (6 transpositions)
	Mode6 = []float64{0, 2, 4, 5, 6, 8, 10, 11}
	// Mode7 repeats semitone, semitone, semitone, tone, semitone (6 transpositions)
	Mode7 = []float64{0, 1, 2, 3, 5, 6, 7, 8, 9, 11}
)

// Transpositions counts the different transpositions of a set of steps within the 12 semitone octave: the
// smallest number of semitones (1..12) that maps the set onto itself, so 12 for an ordinary scale such
// as Major and 2 for the whole tone scale. Steps count modulo 12 (fractions are rounded to
// the nearest semitone) and an empty set returns 1
func Transpositions(steps []float64) int {
	if len(steps) == 0 {
		return 1
	}
	var set [12]bool
	for _, s := range steps {
		if math.IsNaN(s) || math.IsInf(s, 0) {
			continue
		}
		set[((int(math.Round(s))%12)+12)%12] = true
	}
	for k := 1; k < 12; k++ {
		same := true
		for i, in := range set {
			if in != set[(i+k)%12] {
				same = false
				break
			}
		}
		if same {
			return k
		}
	}
	return 12
}
