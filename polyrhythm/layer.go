// Package polyrhythm builds polyrhythms and metric modulations with exact arithmetic.
//
// Conventions. All times are exact fractions (Frac) of a beat, the unit of the tempo in beats per
// minute. A step of a note.Event is 1/StepsPerBeat of a beat; the only floating point conversion is
// the last one, when a time becomes a step. Nothing accumulates, so a 7:4 layer is as exact after a
// thousand bars as after one.
//
// A Layer is a row of equally spaced pulses: n pulses in a Span (3:2 is a layer of 3 pulses and a
// layer of 2 pulses in the same span). Nested tuplets, cross rhythms (a period that is a fraction
// of the bar) and additive meters (3+3+2) are layers or meters too. Cycle tells when layers realign
// and Grid gives the least common multiple grid. Modulation, Plan and Section handle metric
// modulation: a subdivision of the old tempo becomes the beat of the new one.
package polyrhythm

import (
	"strings"

	"github.com/almerlucke/genny/euclid"
	"github.com/almerlucke/genny/note"
	"github.com/almerlucke/genny/pulse"
)

// Voice is how the pulses of a layer sound
type Voice struct {
	Pitch float64
	// Lo and Hi are the velocities of an unaccented and a fully accented pulse (default 60 and 110)
	Lo, Hi float64
	// Gate is the length of a note as a fraction of the pulse period (default 0.5)
	Gate float64
}

func (v Voice) velocity(w float64) float64 {
	lo, hi := v.Lo, v.Hi
	if lo == 0 && hi == 0 {
		lo, hi = 60, 110
	}
	return lo + (hi-lo)*min(max(w, 0), 1)
}

func (v Voice) gate() float64 {
	if v.Gate <= 0 {
		return 0.5
	}
	return v.Gate
}

// Layer is Pulses equally spaced pulses in every Span of beats, the first at Offset beats. Its
// period is Span/Pulses. Pattern and Accent are cycled over the pulse index (pulse 0 is the pulse at
// Offset) and are optional.
type Layer struct {
	Name   string
	Pulses int
	Span   Frac
	Offset Frac
	// Pattern tells which pulses sound (nil: all). It cycles
	Pattern []bool
	// Accent is the weight 0..1 of every pulse (nil: all fully accented). It cycles
	Accent []float64
	Voice
}

// NewLayer returns a layer of pulses in span with a default voice. It panics if pulses is less than 1
// or span is not positive
func NewLayer(pulses int, span Frac) Layer {
	if pulses < 1 || span.Sign() <= 0 {
		panic("polyrhythm: a layer needs at least 1 pulse in a positive span")
	}
	return Layer{Pulses: pulses, Span: span.norm(), Voice: Voice{Pitch: 60, Lo: 60, Hi: 110, Gate: 0.5}}
}

// Every returns a layer with one pulse every period beats: a cross rhythm when the period is a
// fraction of the bar, such as 3/4 of a bar of 4 beats (3 beats)
func Every(period Frac) Layer { return NewLayer(1, period) }

// Tuplet returns a layer of n pulses in the time of m pulses of a layer with the given period:
// 3 in the time of 2 quarter notes is Tuplet(3, 2, Int(1))
func Tuplet(n, m int, period Frac) Layer { return NewLayer(n, period.MulInt(int64(m))) }

// Nest returns a layer of n pulses in the time of one pulse of l: 3:(2:5) is
// NewLayer(2, Int(5)).Nest(3), 2 pulses in 5 beats (period 5/2), each divided in 3 (period 5/6)
func (l Layer) Nest(n int) Layer {
	out := NewLayer(n, l.Period())
	out.Name, out.Voice, out.Offset = l.Name, l.Voice, l.Offset
	return out
}

// Period is the time between two pulses
func (l Layer) Period() Frac {
	if l.Pulses < 1 {
		return Frac{}
	}
	return l.Span.Div(Int(int64(l.Pulses)))
}

func (l Layer) valid() bool { return l.Pulses >= 1 && l.Span.Sign() > 0 }

// Onset is the time of pulse k (k may be negative)
func (l Layer) Onset(k int) Frac { return l.Offset.Add(l.Period().MulInt(int64(k))) }

func pmod(k, n int) int { return ((k % n) + n) % n }

// Sounds tells if pulse k sounds according to the Pattern
func (l Layer) Sounds(k int) bool {
	if len(l.Pattern) == 0 {
		return true
	}
	return l.Pattern[pmod(k, len(l.Pattern))]
}

// Weight is the accent weight of pulse k
func (l Layer) Weight(k int) float64 {
	if len(l.Accent) == 0 {
		return 1
	}
	return l.Accent[pmod(k, len(l.Accent))]
}

// WithEuclid returns a copy that sounds the pulses of an Euclidean rhythm of pulses in steps,
// rotated, cycling over the pulse index
func (l Layer) WithEuclid(steps, pulses, rotation int) Layer {
	l.Pattern = euclid.Pattern(steps, pulses, rotation)
	return l
}

// WithEuclidAccent returns a copy that accents the pulses of an Euclidean rhythm; the others get
// the weight soft
func (l Layer) WithEuclidAccent(steps, pulses, rotation int, soft float64) Layer {
	p := euclid.Pattern(steps, pulses, rotation)
	l.Accent = make([]float64, len(p))
	for i, on := range p {
		l.Accent[i] = soft
		if on {
			l.Accent[i] = 1
		}
	}
	return l
}

// WithBarlow returns a copy with the accents of the metric hierarchy of the divisions (see
// pulse.Hierarchy): Barlow's indispensability scaled to the weights soft..1. With a level above 0 the
// pulses less important than level do not sound (pulse.Thin)
func (l Layer) WithBarlow(soft float64, level int, divisions ...int) Layer {
	ind := pulse.Hierarchy(divisions...)
	l.Accent = pulse.Velocity(ind, soft, 1)
	if level > 0 {
		l.Pattern = pulse.Thin(ind, level)
	}
	return l
}

// Events returns the notes of the pulses that start in [from, to) beats, with the starts and
// durations in steps (stepsPerBeat steps to a beat). An invalid layer gives none
func (l Layer) Events(from, to Frac, stepsPerBeat float64) []note.Event {
	if !l.valid() || to.Cmp(from) <= 0 {
		return nil
	}
	p := l.Period()
	k := int(from.Sub(l.Offset).Div(p).Ceil())
	var out []note.Event
	for ; l.Onset(k).Cmp(to) < 0; k++ {
		if !l.Sounds(k) {
			continue
		}
		out = append(out, note.Event{
			Pitch:    l.Pitch,
			Velocity: l.velocity(l.Weight(k)),
			Start:    l.Onset(k).Float() * stepsPerBeat,
			Duration: p.Float() * l.gate() * stepsPerBeat,
		})
	}
	return out
}

// Cycle is the time after which all layers are at the same place in their pattern again: the
// least common multiple of the periods (and, with patterns, of the pattern lengths in pulses). It is
// 0 without valid layers. Cycle(3:2, 5:4 in a bar of 4) is the bar. 7:4 against a period of
// 7/2 beats is 28 beats
func Cycle(layers ...Layer) Frac {
	var c Frac
	for _, l := range layers {
		if !l.valid() {
			continue
		}
		p := l.Period()
		for _, n := range []int{len(l.Pattern), len(l.Accent)} {
			if n > 1 {
				p = p.MulInt(int64(n))
			}
		}
		if c.IsZero() {
			c = p
		} else {
			c = LCM(c, p)
		}
	}
	return c
}

// periodCycle is the cycle of the pulse rows alone, ignoring patterns
func periodCycle(layers []Layer) Frac {
	var c Frac
	for _, l := range layers {
		if !l.valid() {
			continue
		}
		if c.IsZero() {
			c = l.Period()
		} else {
			c = LCM(c, l.Period())
		}
	}
	return c
}

// onsetAt tells if t is a pulse of the layer, and which one (modulo the cycle)
func (l Layer) onsetAt(t Frac) (int, bool) {
	q := t.Sub(l.Offset).Div(l.Period())
	if !q.IsInt() {
		return 0, false
	}
	return int(q.Floor()), true
}

// Realign returns the times in [0, cycle) at which every layer has a pulse at once (ignoring the
// Patterns), where cycle is the least common multiple of the periods. For 3:2 in a bar there is one
// time, 0. Layers with offsets that never meet give none
func Realign(layers ...Layer) []Frac {
	c := periodCycle(layers)
	if c.IsZero() {
		return nil
	}
	first := -1
	for i, l := range layers {
		if l.valid() {
			first = i
			break
		}
	}
	l0 := layers[first]
	var out []Frac
	n := int(c.Sub(l0.Offset).Div(l0.Period()).Ceil())
	for k := int(Frac{}.Sub(l0.Offset).Div(l0.Period()).Ceil()); k < n; k++ {
		t := l0.Onset(k)
		if t.Sign() < 0 || t.Cmp(c) >= 0 {
			continue
		}
		ok := true
		for _, l := range layers {
			if !l.valid() {
				continue
			}
			if _, hit := l.onsetAt(t); !hit {
				ok = false
				break
			}
		}
		if ok {
			out = append(out, t)
		}
	}
	return out
}

// Grid is the least common multiple grid of some layers: the finest raster on which all pulses of
// one cycle fall exactly
type Grid struct {
	// Unit is the time of one grid cell in beats; 1/Unit is the grid in steps per beat
	Unit Frac
	// Steps is the number of cells in one cycle
	Steps int
	// Cells[i][s] is true if layer i sounds at s*Unit
	Cells [][]bool
}

// maxGrid keeps the grid to a size that fits in memory
const maxGrid = 1 << 22

// NewGrid rasters the layers over one Cycle. The unit is the greatest common divisor of all
// periods and offsets, so 3:2 in a bar of 6 beats (periods 2 and 3) is a grid of 1 beat, 6 cells. A
// grid that would be larger than 4M cells (or of no valid layer) is empty
func NewGrid(layers ...Layer) Grid {
	c := Cycle(layers...)
	if c.IsZero() {
		return Grid{}
	}
	var u Frac
	for _, l := range layers {
		if !l.valid() {
			continue
		}
		for _, f := range []Frac{l.Period(), l.Offset} {
			if f.IsZero() {
				continue
			}
			if u.IsZero() {
				u = f
			} else {
				u = GCD(u, f)
			}
		}
	}
	n := c.Div(u)
	if !n.IsInt() || n.Num > maxGrid {
		return Grid{}
	}
	g := Grid{Unit: u, Steps: int(n.Num), Cells: make([][]bool, len(layers))}
	for i, l := range layers {
		g.Cells[i] = make([]bool, g.Steps)
		if !l.valid() {
			continue
		}
		for s := range g.Steps {
			if k, hit := l.onsetAt(u.MulInt(int64(s))); hit && l.Sounds(k) {
				g.Cells[i][s] = true
			}
		}
	}
	return g
}

// String draws the grid as rows of x and .
func (g Grid) String() string {
	var b strings.Builder
	for i, row := range g.Cells {
		if i > 0 {
			b.WriteByte('\n')
		}
		for _, on := range row {
			if on {
				b.WriteByte('x')
			} else {
				b.WriteByte('.')
			}
		}
	}
	return b.String()
}

// Meter is an additive meter: groups of units, such as 3+3+2 eighth notes (Unit 1/2 beat)
type Meter struct {
	Unit   Frac
	Groups []int
}

// Additive returns a meter of the groups, each group a number of units of the given length in beats.
// 3+3+2 eighths is Additive(F(1, 2), 3, 3, 2). It panics for no groups, a group below 1 or a
// non-positive unit
func Additive(unit Frac, groups ...int) Meter {
	if len(groups) == 0 || unit.Sign() <= 0 {
		panic("polyrhythm: an additive meter needs groups and a positive unit")
	}
	for _, g := range groups {
		if g < 1 {
			panic("polyrhythm: a group must be at least 1 unit")
		}
	}
	return Meter{Unit: unit.norm(), Groups: append([]int(nil), groups...)}
}

// Units is the number of units in a bar
func (m Meter) Units() int {
	n := 0
	for _, g := range m.Groups {
		n += g
	}
	return n
}

// Length is the length of a bar in beats
func (m Meter) Length() Frac { return m.Unit.MulInt(int64(m.Units())) }

// Onsets are the times of the units within a bar
func (m Meter) Onsets() []Frac {
	out := make([]Frac, m.Units())
	for i := range out {
		out[i] = m.Unit.MulInt(int64(i))
	}
	return out
}

// Accents are the weights of the units: 1 for the start of the bar, 0.7 for the start of another
// group, 0.3 inside a group
func (m Meter) Accents() []float64 {
	out := make([]float64, 0, m.Units())
	for gi, g := range m.Groups {
		for i := range g {
			switch {
			case i == 0 && gi == 0:
				out = append(out, 1)
			case i == 0:
				out = append(out, 0.7)
			default:
				out = append(out, 0.3)
			}
		}
	}
	return out
}

// TimeSignature is the meter as a time signature when the unit is a power of two note value
// (unit 1/2 beat is an eighth: 8). ok is false otherwise
func (m Meter) TimeSignature() (num, den int, ok bool) {
	d := Int(4).Div(m.Unit)
	if !d.IsInt() || d.Num&(d.Num-1) != 0 || d.Num > 128 {
		return 0, 0, false
	}
	return m.Units(), int(d.Num), true
}

// Layer is the meter as a layer: one pulse per unit with the group accents, cycling every bar.
// Cross rhythms with other layers come from the bar length against theirs
func (m Meter) Layer() Layer {
	l := NewLayer(m.Units(), m.Length())
	l.Accent = m.Accents()
	return l
}
