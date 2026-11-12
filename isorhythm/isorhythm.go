// Package isorhythm makes isorhythmic lines, as in the motets of the Ars nova: a talea (a cycle of
// durations, a negative duration is a rest) and a color (a cycle of pitches) of different lengths run
// against each other. A talea of 5 and a color of 7 only repeat together after 35 notes
package isorhythm

import (
	"math"

	"github.com/almerlucke/genny/note"
	"github.com/almerlucke/genny/rest"
)

// Events lays the talea and the color against each other until the line is length steps long. The talea
// element i gives the duration of note i (negative: a rest of that length), the color element i modulo the
// length of the color gives its pitch. The color advances with every talea element, rests included,
// so the whole pattern repeats after Period elements. The last note is cut at length. Zero durations
// are skipped. It returns nil if the talea or the color is empty, or if the talea has no durations
func Events(talea []float64, color []float64, length float64) []note.Event {
	if len(talea) == 0 || len(color) == 0 || !(length > 0) || !hasDuration(talea) {
		return nil
	}
	var out []note.Event
	t := 0.0
	for i := 0; t < length; i++ {
		d := talea[i%len(talea)]
		if d == 0 {
			continue
		}
		dur := d
		if dur < 0 {
			dur = -dur
		} else {
			out = append(out, note.Event{
				Pitch:    color[i%len(color)],
				Velocity: 0.8,
				Start:    t,
				Duration: min(dur, length-t),
			})
		}
		t += dur
	}
	return out
}

func hasDuration(talea []float64) bool {
	for _, d := range talea {
		if d != 0 {
			return true
		}
	}
	return false
}

// Period is the number of talea elements after which talea and color start again together: the least
// common multiple of their lengths (35 for 5 and 7). It is 0 if one of them is empty
func Period(talea, color []float64) int {
	if len(talea) == 0 || len(color) == 0 {
		return 0
	}
	a, b := len(talea), len(color)
	for b != 0 {
		a, b = b, a%b
	}
	return len(talea) / a * len(color)
}

// Steps is a talea and a color for a note.Voice, which plays on a grid of steps: the durations are rounded
// to whole steps (at least 1). Pitch gives the pitch of a note on the step where it starts and a rest on the
// steps it sounds on; Length gives the length of the note that starts on the step (and 1 elsewhere). Use
// a Pitch and a Length of the same Steps in one voice: they advance once per step and stay together
type Steps struct {
	talea []float64
	color []float64
}

// New creates the grid form of a talea and a color. It panics if the talea or the color is empty or if
// the talea has no durations other than 0
func New(talea, color []float64) *Steps {
	if len(talea) == 0 || len(color) == 0 || !hasDuration(talea) {
		panic("isorhythm: a talea and a color are needed")
	}
	return &Steps{talea: append([]float64{}, talea...), color: append([]float64{}, color...)}
}

// cursor walks the grid
type cursor struct {
	s    *Steps
	pos  int // talea elements started
	left int // steps left of the current element
	cur  float64
	rest bool
	n    int
}

// next advances one step and reports the element that sounds, and whether it starts here
func (c *cursor) next() (pitch float64, rest, starts bool, length int) {
	if c.left == 0 {
		for {
			d := c.s.talea[c.pos%len(c.s.talea)]
			c.pos++
			if d == 0 {
				continue
			}
			c.rest = d < 0
			c.n = max(int(math.Abs(d)+0.5), 1)
			c.left = c.n
			c.cur = c.s.color[(c.pos-1)%len(c.s.color)]
			starts = true
			break
		}
	}
	c.left--
	return c.cur, c.rest || !starts, starts, c.n
}

func (c *cursor) reset() { c.pos, c.left = 0, 0 }

// Pitch returns a generator of pitches with rests, a genny.Generator[rest.Optional[float64]]
func (s *Steps) Pitch() *Pitch { return &Pitch{cursor{s: s}} }

// Length returns a generator of note lengths, a genny.Generator[float64]
func (s *Steps) Length() *Length { return &Length{cursor{s: s}} }

// Pitch generates the pitch of each step, see Steps
type Pitch struct{ c cursor }

func (p *Pitch) Generate() rest.Optional[float64] {
	pitch, r, _, _ := p.c.next()
	if r {
		return rest.None[float64]()
	}
	return rest.Some(pitch)
}
func (p *Pitch) Continuous() bool { return true }
func (p *Pitch) Done() bool       { return false }
func (p *Pitch) Reset()           { p.c.reset() }

// Length generates the length of the note that starts on each step, see Steps
type Length struct{ c cursor }

func (l *Length) Generate() float64 {
	_, _, starts, n := l.c.next()
	if starts {
		return float64(n)
	}
	return 1
}
func (l *Length) Continuous() bool { return true }
func (l *Length) Done() bool       { return false }
func (l *Length) Reset()           { l.c.reset() }
