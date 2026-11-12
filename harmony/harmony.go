// Package harmony generates functional chord progressions and realises them in four voices.
//
// A progression is a stream of Chord values (genny.Generator[Chord]): chords with a key, a Roman
// numeral and a function (tonic, subdominant or dominant). Moves between the functions follow the
// grammar T -> S -> D -> T with weights you can change, phrases end in a cadence (authentic, half,
// plagal or deceptive), and chords can be dressed up with sevenths, secondary dominants (V/ii, V7/V)
// and a few chords borrowed from the parallel minor. Modulate gives the chords to move into another
// key through a pivot chord.
//
// Realize turns a list of chords into soprano, alto, tenor and bass notes (SATB) with the rules of
// four part writing: no parallel fifths or octaves, no crossing or overlapping voices, leading
// tones and sevenths resolve, small leaps, a doubled root in triads and a leading tone that is
// never doubled. Check is a separate implementation of the same rules that lists the violations of
// any voicing, written by hand or generated.
//
// Not modelled: chord inversions in the grammar (the voicing picks the bass note), the cadential
// 6/4, suspensions and other non chord tones, hidden (direct) fifths and octaves, augmented
// intervals in a voice, modal mixture beyond iv and bVI, tonicisation longer than one chord, key
// signatures and enharmonic spelling (pitches are MIDI numbers), and meter or rhythm: every chord
// is one event.
package harmony

import (
	"fmt"
	"strings"

	"github.com/almerlucke/genny/chord"
)

// Function is the role of a chord in a key
type Function int

const (
	Tonic Function = iota
	Subdominant
	Dominant
)

func (f Function) String() string {
	switch f {
	case Tonic:
		return "T"
	case Subdominant:
		return "S"
	case Dominant:
		return "D"
	}
	return "?"
}

// Key is a major or minor key. Tonic is a MIDI note number or pitch class (only the pitch class
// counts). The minor key uses the harmonic minor for its dominant chords: a major V and a vii
// diminished on the raised seventh
type Key struct {
	Tonic int
	Minor bool
}

var noteNames = []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}

func pcName(pc int) string { return noteNames[mod12(pc)] }

func mod12(x int) int { return ((x % 12) + 12) % 12 }

func (k Key) String() string {
	if k.Minor {
		return pcName(k.Tonic) + " minor"
	}
	return pcName(k.Tonic) + " major"
}

// Dominant is the key a fifth up with the same mode
func (k Key) Dominant() Key { return Key{mod12(k.Tonic + 7), k.Minor} }

// Relative is the relative minor of a major key, the relative major of a minor key
func (k Key) Relative() Key {
	if k.Minor {
		return Key{mod12(k.Tonic + 3), false}
	}
	return Key{mod12(k.Tonic + 9), true}
}

// Chord is a chord in a key. Root is in semitones above the tonic of Key and Intervals are the
// semitones above the root (a triad has three, a seventh chord four)
type Chord struct {
	Key       Key
	Numeral   string // "V7", "ii", "V/vi", "vi=ii" for a pivot chord
	Function  Function
	Root      int
	Intervals []int
	Of        string // the numeral a secondary dominant points at, empty otherwise
	Borrowed  bool   // from the parallel key
}

// PitchClasses returns the pitch classes of the chord (0 is C) in chord order: root, third, fifth
// and for a seventh chord the seventh
func (c Chord) PitchClasses() []int {
	out := make([]int, len(c.Intervals))
	for i, iv := range c.Intervals {
		out[i] = mod12(c.Key.Tonic + c.Root + iv)
	}
	return out
}

// IsSeventh reports whether the chord has a seventh
func (c Chord) IsSeventh() bool { return len(c.Intervals) == 4 }

// Seventh returns the pitch class of the seventh, if there is one
func (c Chord) Seventh() (int, bool) {
	if !c.IsSeventh() {
		return 0, false
	}
	return c.PitchClasses()[3], true
}

// LeadingTone returns the pitch class of the leading tone of a dominant chord: the third of a
// major dominant (V, V7, V/x) or the root of a diminished vii chord. It is the note that must rise
// by a semitone
func (c Chord) LeadingTone() (int, bool) {
	if c.Function != Dominant || len(c.Intervals) < 3 {
		return 0, false
	}
	pcs := c.PitchClasses()
	if c.Intervals[1] == 4 {
		return pcs[1], true
	}
	if c.Intervals[1] == 3 && c.Intervals[2] == 6 {
		return pcs[0], true
	}
	return 0, false
}

// Name is the chord as a lead sheet symbol: G, Em, Bdim, D7, Am7
func (c Chord) Name() string {
	root := pcName(c.Key.Tonic + c.Root)
	switch strings.Join(intStrings(c.Intervals), ",") {
	case "0,4,7":
		return root
	case "0,3,7":
		return root + "m"
	case "0,3,6":
		return root + "dim"
	case "0,4,7,10":
		return root + "7"
	case "0,3,7,10":
		return root + "m7"
	case "0,3,6,10":
		return root + "m7b5"
	case "0,3,6,9":
		return root + "dim7"
	case "0,4,7,11":
		return root + "maj7"
	}
	return root + "?"
}

func intStrings(a []int) []string {
	out := make([]string, len(a))
	for i, x := range a {
		out[i] = fmt.Sprint(x)
	}
	return out
}

func (c Chord) String() string { return c.Numeral + " (" + c.Name() + ")" }

func ints(f []float64) []int {
	out := make([]int, len(f))
	for i, x := range f {
		out[i] = int(x)
	}
	return out
}

type def struct {
	root int
	ivs  []int
	fn   Function
	seg  int // semitones of the seventh above the root
}

var (
	majorDefs = [7]def{
		{0, ints(chord.Major), Tonic, 11}, {2, ints(chord.Minor), Subdominant, 10}, {4, ints(chord.Minor), Tonic, 10},
		{5, ints(chord.Major), Subdominant, 11}, {7, ints(chord.Major), Dominant, 10}, {9, ints(chord.Minor), Tonic, 10},
		{11, ints(chord.Diminished), Dominant, 10},
	}
	minorDefs = [7]def{
		{0, ints(chord.Minor), Tonic, 10}, {2, ints(chord.Diminished), Subdominant, 10}, {3, ints(chord.Major), Tonic, 11},
		{5, ints(chord.Minor), Subdominant, 10}, {7, ints(chord.Major), Dominant, 10}, {8, ints(chord.Major), Tonic, 11},
		{11, ints(chord.Diminished), Dominant, 9},
	}
	romans = []string{"I", "II", "III", "IV", "V", "VI", "VII"}
)

func defs(k Key) [7]def {
	if k.Minor {
		return minorDefs
	}
	return majorDefs
}

func numeralFor(deg int, ivs []int) string {
	r := romans[deg]
	switch {
	case ivs[1] == 3:
		r = strings.ToLower(r)
		if ivs[2] == 6 {
			if len(ivs) == 4 && ivs[3] == 10 {
				return r + "ø7"
			}
			if len(ivs) == 4 {
				return r + "°7"
			}
			return r + "°"
		}
	}
	if len(ivs) == 4 {
		if ivs[3] == 11 {
			return r + "maj7"
		}
		return r + "7"
	}
	return r
}

func withSeventh(ivs []int, sev int) []int {
	return append(append([]int(nil), ivs...), sev)
}

// Diatonic is the chord on a scale degree (0 is I, 4 is V) of the key, a triad or with the seventh
// of the scale on top. The degree is taken modulo 7
func Diatonic(k Key, degree int, seventh bool) Chord {
	k.Tonic = mod12(k.Tonic)
	d := ((degree % 7) + 7) % 7
	df := defs(k)[d]
	ivs := append([]int(nil), df.ivs...)
	if seventh {
		ivs = withSeventh(ivs, df.seg)
		if k.Minor && d == 1 {
			ivs[3] = 10
		}
	}
	return Chord{Key: k, Numeral: numeralFor(d, ivs), Function: df.fn, Root: df.root, Intervals: ivs}
}

// degreeOf finds the diatonic degree of a chord by root and quality, or -1
func degreeOf(c Chord) int {
	for d, df := range defs(c.Key) {
		if df.root == c.Root && len(c.Intervals) >= 3 && df.ivs[1] == c.Intervals[1] && df.ivs[2] == c.Intervals[2] {
			return d
		}
	}
	return -1
}

// SecondaryDominant is the dominant of a scale degree: V/ii, V7/V. Only degrees whose chord is a
// major or minor triad can be tonicised (not the tonic itself, not the diminished chords, and in a
// major key not IV, whose dominant is just the tonic chord); ok is false for the others
func SecondaryDominant(k Key, degree int, seventh bool) (Chord, bool) {
	k.Tonic = mod12(k.Tonic)
	d := ((degree % 7) + 7) % 7
	df := defs(k)[d]
	if d == 0 || df.ivs[2] == 6 || (d == 3 && !k.Minor) {
		return Chord{}, false
	}
	target := Diatonic(k, d, false)
	ivs := ints(chord.Major)
	num := "V"
	if seventh {
		ivs = ints(chord.Dominant7)
		num = "V7"
	}
	return Chord{
		Key: k, Numeral: num + "/" + target.Numeral, Function: Dominant,
		Root: (df.root + 7) % 12, Intervals: ivs, Of: target.Numeral,
	}, true
}

// Borrowed is a chord from the parallel minor in a major key: degree 3 gives iv, degree 5 gives
// bVI. Other degrees, or a minor key, give ok false
func Borrowed(k Key, degree int) (Chord, bool) {
	k.Tonic = mod12(k.Tonic)
	if k.Minor {
		return Chord{}, false
	}
	switch degree {
	case 3:
		return Chord{Key: k, Numeral: "iv", Function: Subdominant, Root: 5, Intervals: ints(chord.Minor), Borrowed: true}, true
	case 5:
		return Chord{Key: k, Numeral: "bVI", Function: Subdominant, Root: 8, Intervals: ints(chord.Major), Borrowed: true}, true
	}
	return Chord{}, false
}

// Modulate gives the chords that lead from one key into another: a pivot chord, a triad of the
// old key that is a ii or IV of the new one (labelled like "vi=ii"), then the dominant seventh of
// the new key. If there is no pivot the dominant seventh comes alone. A chord on the tonic of the
// new key should follow. The chords are in the new key
func Modulate(from, to Key) []Chord {
	from.Tonic, to.Tonic = mod12(from.Tonic), mod12(to.Tonic)
	var out []Chord
	for _, d := range []int{1, 3} {
		p := Diatonic(to, d, false)
		if p.Intervals[2] == 6 {
			continue
		}
		pa := mod12(to.Tonic + p.Root)
		if e := degreeIn(from, pa, p.Intervals); e >= 0 {
			p.Numeral = Diatonic(from, e, false).Numeral + "=" + p.Numeral
			out = append(out, p)
			break
		}
	}
	return append(out, Diatonic(to, 4, true))
}

func degreeIn(k Key, pitchClass int, ivs []int) int {
	for d, df := range defs(k) {
		if mod12(k.Tonic+df.root) == pitchClass && df.ivs[1] == ivs[1] && df.ivs[2] == ivs[2] {
			return d
		}
	}
	return -1
}
