// Package tendency has tendency masks, as used by Gottfried Michael Koenig: a band between a lower and an
// upper bound that moves over time. Every value that is drawn lies inside the band, so a band that
// widens or narrows, or glides up and down, shapes a stream of random values (pitches, durations,
// velocities) without fixing them. A mask is a list of segments, each with a start and an end band
package tendency

import (
	"math"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/float/dist"
)

// Segment is a part of a mask. Over N values the lower bound moves linearly from A1 to Z1 and the upper
// bound from A2 to Z2. The first value has the band A1..A2 and the last Z1..Z2. A segment with N = 1 has
// just the band A1..A2, one with N = 0 is skipped. If a lower bound is above its upper bound they swap
type Segment struct {
	N      int
	A1, A2 float64
	Z1, Z2 float64
}

// Mask is a chain of segments. It only describes the bands: Value and Choice draw from it
type Mask struct {
	segs  []Segment
	total int
	loop  bool
}

// NewMask chains segments into a mask. A looping mask starts again after its last segment, otherwise
// the last band stays. It panics if a segment has a negative N
func NewMask(loop bool, segments ...Segment) *Mask {
	m := &Mask{loop: loop}
	for _, s := range segments {
		if s.N < 0 {
			panic("tendency: segment length must not be negative")
		}
		if s.N > 0 {
			m.segs = append(m.segs, s)
			m.total += s.N
		}
	}
	return m
}

// Len is the number of values of one pass through the mask
func (m *Mask) Len() int { return m.total }

// Loops is true if the mask starts again after the last segment
func (m *Mask) Loops() bool { return m.loop }

// At returns the band (lo <= hi) at position i. Past the end a looping mask wraps around and
// another keeps its last band, a position below 0 is 0. An empty mask has the band 0..0
func (m *Mask) At(i int) (lo, hi float64) {
	if m.total == 0 {
		return 0, 0
	}
	if i < 0 {
		i = 0
	}
	if i >= m.total {
		if m.loop {
			i %= m.total
		} else {
			i = m.total - 1
		}
	}
	for _, s := range m.segs {
		if i < s.N {
			t := 0.0
			if s.N > 1 {
				t = float64(i) / float64(s.N-1)
			}
			lo = s.A1 + (s.Z1-s.A1)*t
			hi = s.A2 + (s.Z2-s.A2)*t
			if lo > hi {
				lo, hi = hi, lo
			}
			return lo, hi
		}
		i -= s.N
	}
	return 0, 0
}

// cursor walks through a mask and draws a unit value in [0, 1] for every step
type cursor struct {
	mask *Mask
	pos  int
	draw func() float64
}

func (c *cursor) done() bool {
	return c.mask.total == 0 || (!c.mask.loop && c.pos >= c.mask.total)
}

// next returns the band of the current step and moves on
func (c *cursor) next() (lo, hi float64) {
	lo, hi = c.mask.At(c.pos)
	c.pos++
	return
}

func uniform() float64 { return genny.Float64() }

// Value generates values between the bounds of a mask: each Generate draws one value from the current
// band and moves to the next. When a mask that does not loop is finished, Done is true and Generate
// keeps returning the last value. An empty mask is done at once and generates 0
type Value struct {
	cursor
	last float64
}

// NewValue draws uniformly between the bounds. It panics if mask is nil
func NewValue(mask *Mask) *Value {
	return newValue(mask, uniform)
}

// NewBetaValue draws between the bounds with a beta distribution (see dist.SampleBeta): a < b favours
// the lower bound, a > b the upper bound, both below 1 the two bounds, both above 1 the middle of the
// band. It panics if mask is nil, or a or b is not positive
func NewBetaValue(mask *Mask, a, b float64) *Value {
	checkBeta(a, b)
	return newValue(mask, func() float64 { return dist.SampleBeta(a, b) })
}

func newValue(mask *Mask, draw func() float64) *Value {
	if mask == nil {
		panic("tendency: mask must not be nil")
	}
	return &Value{cursor: cursor{mask: mask, draw: draw}}
}

func checkBeta(a, b float64) {
	if !(a > 0) || !(b > 0) {
		panic("tendency: beta parameters must be positive")
	}
}

func (v *Value) Generate() float64 {
	if v.done() {
		return v.last
	}
	lo, hi := v.next()
	v.last = lo + (hi-lo)*v.draw()
	return v.last
}

func (v *Value) Continuous() bool { return v.mask.loop }

func (v *Value) Done() bool { return v.done() }

func (v *Value) Reset() {
	v.pos = 0
	v.last = 0
}

// Choice picks elements of a list. The bounds of the mask are percentages (0 to 100) of the list: a
// band of 0..50 picks from the first half and 25..75 from the middle. Percentages outside 0..100
// are clamped. When a mask that does not loop is finished, Done is true and Generate keeps returning the
// last pick. An empty mask or an empty list is done at once and generates the zero value
type Choice[T any] struct {
	cursor
	items []T
	last  T
}

// NewChoice picks uniformly between the bounds, from a copy of items. It panics if mask is nil
func NewChoice[T any](mask *Mask, items []T) *Choice[T] {
	return newChoice(mask, items, uniform)
}

// NewBetaChoice picks with a beta distribution between the bounds, see NewBetaValue. It panics if mask
// is nil, or a or b is not positive
func NewBetaChoice[T any](mask *Mask, items []T, a, b float64) *Choice[T] {
	checkBeta(a, b)
	return newChoice(mask, items, func() float64 { return dist.SampleBeta(a, b) })
}

func newChoice[T any](mask *Mask, items []T, draw func() float64) *Choice[T] {
	if mask == nil {
		panic("tendency: mask must not be nil")
	}
	return &Choice[T]{cursor: cursor{mask: mask, draw: draw}, items: append([]T(nil), items...)}
}

func (c *Choice[T]) isDone() bool { return len(c.items) == 0 || c.done() }

func (c *Choice[T]) Generate() T {
	if c.isDone() {
		return c.last
	}
	lo, hi := c.next()
	lo, hi = min(max(lo, 0), 100), min(max(hi, 0), 100)
	p := (lo + (hi-lo)*c.draw()) / 100
	i := min(max(int(math.Floor(p*float64(len(c.items)))), 0), len(c.items)-1)
	c.last = c.items[i]
	return c.last
}

func (c *Choice[T]) Continuous() bool { return c.mask.loop && len(c.items) > 0 }

func (c *Choice[T]) Done() bool { return c.isDone() }

func (c *Choice[T]) Reset() {
	var zero T
	c.pos = 0
	c.last = zero
}

// TwoMasks draws values from an outer mask but keeps them out of an inner mask: a value that lies inside
// the inner band of the same step is rejected and drawn again, up to a number of retries. If every
// retry fails the last outer value is used, so it never blocks. With the inner band inside the outer one, the
// values avoid a hole that moves over time. It is done when the outer mask is done
type TwoMasks struct {
	outer, inner *Mask
	retries      int
	pos          int
	last         float64
}

// NewTwoMasks creates the generator. A negative retries is 0. It panics if a mask is nil
func NewTwoMasks(outer, inner *Mask, retries int) *TwoMasks {
	if outer == nil || inner == nil {
		panic("tendency: masks must not be nil")
	}
	return &TwoMasks{outer: outer, inner: inner, retries: max(retries, 0)}
}

func (t *TwoMasks) isDone() bool {
	return t.outer.total == 0 || (!t.outer.loop && t.pos >= t.outer.total)
}

func (t *TwoMasks) Generate() float64 {
	if t.isDone() {
		return t.last
	}
	lo, hi := t.outer.At(t.pos)
	ilo, ihi := 0.0, -1.0 // no inner band: nothing is inside
	if t.inner.total > 0 {
		ilo, ihi = t.inner.At(t.pos)
	}
	t.pos++
	v := lo + (hi-lo)*genny.Float64()
	for i := 0; i < t.retries && v >= ilo && v <= ihi; i++ {
		v = lo + (hi-lo)*genny.Float64()
	}
	t.last = v
	return v
}

func (t *TwoMasks) Continuous() bool { return t.outer.loop }

func (t *TwoMasks) Done() bool { return t.isDone() }

func (t *TwoMasks) Reset() {
	t.pos = 0
	t.last = 0
}

// Anchor makes n values that follow a shape but start at first and end at last, within lo and hi. The shape is
// stretched to n values (linear interpolation) and scaled to fill lo..hi, then bent so it begins and ends at the
// anchors, and clamped. An empty or flat shape gives a straight line from first to last. A swapped lo
// and hi are put in order. Fewer than 1 value gives nil, and 1 gives first
func Anchor(shape []float64, n int, first, last, lo, hi float64) []float64 {
	if n < 1 {
		return nil
	}
	if lo > hi {
		lo, hi = hi, lo
	}
	first, last = min(max(first, lo), hi), min(max(last, lo), hi)
	if n == 1 {
		return []float64{first}
	}

	unit := make([]float64, n) // the shape stretched and scaled to 0..1
	smin, smax := math.Inf(1), math.Inf(-1)
	for _, v := range shape {
		smin, smax = min(smin, v), max(smax, v)
	}
	for i := range unit {
		if len(shape) > 0 && smax > smin {
			pos := float64(i) / float64(n-1) * float64(len(shape)-1)
			j := min(int(pos), len(shape)-1)
			v := shape[j]
			if j+1 < len(shape) {
				v += (shape[j+1] - v) * (pos - float64(j))
			}
			unit[i] = (v - smin) / (smax - smin)
		}
	}

	out := make([]float64, n)
	flat := len(shape) == 0 || smax <= smin
	d0, d1 := 0.0, 0.0
	if !flat {
		d0 = first - (lo + unit[0]*(hi-lo))
		d1 = last - (lo + unit[n-1]*(hi-lo))
	}
	for i := range out {
		t := float64(i) / float64(n-1)
		if flat {
			out[i] = first + (last-first)*t
			continue
		}
		v := lo + unit[i]*(hi-lo) + d0*(1-t) + d1*t
		out[i] = min(max(v, lo), hi)
	}
	return out
}

// Spray scatters values: every value moves by a random amount within plus or minus spread
// (drawn from the genny random stream). It returns a new slice. A negative spread is 0
func Spray(values []float64, spread float64) []float64 {
	spread = max(spread, 0)
	out := make([]float64, len(values))
	for i, v := range values {
		out[i] = v + spread*(2*genny.Float64()-1)
	}
	return out
}
