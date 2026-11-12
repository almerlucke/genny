package scale

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

// Scl is a tuning read from a Scala .scl file: a description and the pitches above the 1/1 as cents.
// The last pitch is the period (the interval at which the scale repeats, 1200 for an octave).
// Cents is never empty, ascends strictly and is positive when it comes from ParseScl
type Scl struct {
	// Description is the first line of the file
	Description string
	// Cents holds the pitch of every line after the count: the degrees above the root and, last, the period
	Cents []float64
}

// ParseScl reads a Scala .scl file. The first line that is not a comment ('!') is the description (it may
// be empty), the next is the number of notes and then follow that many pitches, one per line, each a
// value in cents (it contains a dot: 701.955) or a ratio (3/2, or an integer such as 2 for 2/1); text after
// the number is ignored. The pitches must ascend strictly and above 0 (the unison 1/1 is implied), which
// real tunings do. Anything else, malformed input included, gives an error, ParseScl never panics
func ParseScl(r io.Reader) (Scl, error) {
	var scl Scl
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)

	// next returns the next line that is not a comment
	next := func() (string, bool) {
		for sc.Scan() {
			line := strings.TrimRight(sc.Text(), "\r")
			if !strings.HasPrefix(strings.TrimSpace(line), "!") {
				return line, true
			}
		}
		return "", false
	}

	desc, ok := next()
	if !ok {
		if err := sc.Err(); err != nil {
			return scl, err
		}
		return scl, errors.New("scale: empty scl file")
	}
	scl.Description = strings.TrimSpace(desc)

	var countLine string
	for {
		line, ok := next()
		if !ok {
			return scl, errors.New("scale: scl file has no note count")
		}
		if countLine = strings.TrimSpace(line); countLine != "" {
			break
		}
	}
	fields := strings.Fields(countLine)
	n, err := strconv.Atoi(fields[0])
	if err != nil || n < 1 || n > 4096 {
		return scl, fmt.Errorf("scale: scl note count %q must be an integer from 1 to 4096", fields[0])
	}

	for len(scl.Cents) < n {
		line, ok := next()
		if !ok {
			if err := sc.Err(); err != nil {
				return scl, err
			}
			return scl, fmt.Errorf("scale: scl file has %d of %d pitches", len(scl.Cents), n)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		c, err := parsePitch(strings.Fields(line)[0])
		if err != nil {
			return scl, err
		}
		if last := lastOr(scl.Cents, 0); !(c > last) {
			return scl, fmt.Errorf("scale: scl pitch %q does not ascend above %.3f cents", line, last)
		}
		scl.Cents = append(scl.Cents, c)
	}
	if err := sc.Err(); err != nil {
		return Scl{}, err
	}
	return scl, nil
}

func lastOr(s []float64, def float64) float64 {
	if len(s) == 0 {
		return def
	}
	return s[len(s)-1]
}

// parsePitch reads one pitch as cents ("701.955") or as a ratio ("3/2", "2")
func parsePitch(f string) (float64, error) {
	if strings.Contains(f, ".") {
		c, err := strconv.ParseFloat(f, 64)
		if err != nil || math.IsNaN(c) || math.IsInf(c, 0) {
			return 0, fmt.Errorf("scale: scl pitch %q is not a number of cents", f)
		}
		return c, nil
	}
	num, den := f, "1"
	if i := strings.IndexByte(f, '/'); i >= 0 {
		num, den = f[:i], f[i+1:]
	}
	a, err1 := strconv.ParseUint(num, 10, 53)
	b, err2 := strconv.ParseUint(den, 10, 53)
	if err1 != nil || err2 != nil || a == 0 || b == 0 {
		return 0, fmt.Errorf("scale: scl pitch %q is not a ratio", f)
	}
	return 1200 * math.Log2(float64(a)/float64(b)), nil
}

// Period is the interval at which the tuning repeats, in cents
func (s Scl) Period() float64 { return lastOr(s.Cents, 1200) }

// Steps returns the degrees in semitones above the root: 0 and the pitches before the last one, divided
// by 100
func (s Scl) Steps() []float64 {
	steps := []float64{0}
	for i := 0; i+1 < len(s.Cents); i++ {
		steps = append(steps, s.Cents[i]/100)
	}
	return steps
}

// Scale creates a scale from the tuning with the given root pitch (in semitones): 100 cents are a semitone and
// the period is the last pitch, so Bohlen-Pierce repeats every tritave. The pitches are fractional, a
// microtonal scale needs an output that can play them, see midi.Track.PitchBend. It panics if the Scl
// is empty or does not ascend, which cannot happen for one from ParseScl
func (s Scl) Scale(root float64) *Scale {
	return NewPeriod(root, s.Steps(), s.Period()/100)
}

// Text of some tunings in Scala format
const (
	// SclTET12 is 12 tone equal temperament
	SclTET12 = `! 12tet.scl
12 tone equal temperament
12
!
100.0
200.0
300.0
400.0
500.0
600.0
700.0
800.0
900.0
1000.0
1100.0
2/1
`
	// SclJust5 is the 5-limit just intonation chromatic scale of Ptolemy
	SclJust5 = `! ptolemy.scl
Ptolemy's intense diatonic scale extended to a 5-limit chromatic scale
12
!
16/15
9/8
6/5
5/4
4/3
45/32
3/2
8/5
5/3
9/5
15/8
2/1
`
	// SclMeantone is quarter-comma meantone, 12 notes (Eb to G#)
	SclMeantone = `! meanquar.scl
1/4-comma meantone scale. Pietro Aaron's temperament (1523)
12
!
76.04900
193.15686
310.26471
386.31371
503.42157
579.47057
696.57843
772.62743
889.73529
1006.84314
1082.89214
2/1
`
	// SclPythagorean is the Pythagorean chromatic scale built from pure fifths
	SclPythagorean = `! pyth_12.scl
Pythagorean chromatic scale
12
!
256/243
9/8
32/27
81/64
4/3
729/512
3/2
128/81
27/16
16/9
243/128
2/1
`
	// SclBohlenPierce is the Bohlen-Pierce scale: 13 steps to the tritave (3/1) instead of the octave
	SclBohlenPierce = `! bohlen-pierce.scl
Bohlen-Pierce scale, just version
13
!
27/25
25/21
9/7
7/5
75/49
5/3
9/5
49/25
15/7
7/3
63/25
25/9
3/1
`
)
