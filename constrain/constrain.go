// Package constrain puts constraints on a generator by rejection: it draws a value, and draws again
// until the value is allowed. This is the way of Paul Berg's AC Toolbox, where a stream of
// random values is steered by rules about intervals, repetitions and duplicates. Every wrapper has a
// retry limit (Stop, 100 by default) so an impossible constraint cannot hang: when the limit is
// reached the last drawn value is used anyway. The package also has two melodic walks,
// ScaleIntervals and RandomIntervals, that build lines from lists of intervals
package constrain

import (
	"math"

	"github.com/almerlucke/genny"
)

// DefaultStop is the number of draws tried before a constraint gives up
const DefaultStop = 100

// Number is a type with an interval: the difference of two values
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~float32 | ~float64
}

// Option changes the behaviour of a constraint
type Option func(*options)

type options struct {
	stop       int
	pitchClass bool
}

// Stop sets the number of draws a constraint tries before it gives up and uses the last value (at
// least 1, default DefaultStop)
func Stop(n int) Option { return func(o *options) { o.stop = max(n, 1) } }

// PitchClass compares intervals and values mod 12: a fifth up and a fourth down are the same, and so
// are all octaves of a note
func PitchClass() Option { return func(o *options) { o.pitchClass = true } }

func newOptions(opts []Option) options {
	o := options{stop: DefaultStop}
	for _, f := range opts {
		f(&o)
	}
	return o
}

// filter draws from gen until accept likes the value, at most stop times
type filter[T any] struct {
	gen    genny.Generator[T]
	accept func(prev T, hasPrev bool, cand T) bool
	stop   int
	prev   T
	has    bool
	// onAccept is called with every value that is given out
	onAccept func(T)
	// onReset clears the state of the rule
	onReset func()
}

func (f *filter[T]) Generate() T {
	var cand T
	for range f.stop {
		cand = f.gen.Generate()
		if f.gen.Done() || f.accept(f.prev, f.has, cand) {
			break
		}
	}
	f.prev, f.has = cand, true
	if f.onAccept != nil {
		f.onAccept(cand)
	}
	return cand
}

func (f *filter[T]) Continuous() bool { return f.gen.Continuous() }
func (f *filter[T]) Done() bool       { return f.gen.Done() }

func (f *filter[T]) Reset() {
	f.gen.Reset()
	var zero T
	f.prev, f.has = zero, false
	if f.onReset != nil {
		f.onReset()
	}
}

// interval is the absolute distance of two values, mod 12 if asked, rounded to a whole interval
func interval[T Number](a, b T, pitchClass bool) int {
	d := math.Abs(float64(b) - float64(a))
	if pitchClass {
		d = math.Mod(d, 12)
	}
	return int(math.Round(d))
}

func set(values []int) map[int]bool {
	m := make(map[int]bool, len(values))
	for _, v := range values {
		m[v] = true
	}
	return m
}

// AllowInterval only lets values through that are an allowed interval (in semitones, absolute
// value: 2 is a step up or down) away from the previous value that came out. The first value is always
// accepted. With PitchClass the interval is taken mod 12, so 14 counts as 2. No allowed
// intervals means no value is ever allowed: every draw fails and the last of Stop draws is used
func AllowInterval[T Number](gen genny.Generator[T], allowed []int, opts ...Option) genny.Generator[T] {
	return intervalFilter(gen, allowed, true, opts)
}

// BlockInterval is the inverse of AllowInterval: values that are one of the listed intervals away
// from the previous one are drawn again. The first value is always accepted
func BlockInterval[T Number](gen genny.Generator[T], blocked []int, opts ...Option) genny.Generator[T] {
	return intervalFilter(gen, blocked, false, opts)
}

func intervalFilter[T Number](gen genny.Generator[T], list []int, want bool, opts []Option) genny.Generator[T] {
	o := newOptions(opts)
	in := set(list)
	return &filter[T]{
		gen:  gen,
		stop: o.stop,
		accept: func(prev T, has bool, cand T) bool {
			return !has || in[interval(prev, cand, o.pitchClass)] == want
		},
	}
}

// RemoveRepetitions draws again when the value equals the previous one
func RemoveRepetitions[T comparable](gen genny.Generator[T], opts ...Option) genny.Generator[T] {
	o := newOptions(opts)
	return &filter[T]{
		gen:    gen,
		stop:   o.stop,
		accept: func(prev T, has bool, cand T) bool { return !has || prev != cand },
	}
}

// Duplicates lets no value occur twice within a group of n values: a value already in the group is drawn
// again. After n values the group starts afresh. It panics if n < 1. Note that a source with fewer
// than n different values cannot fill a group, the limit then lets repeats through
func Duplicates[T comparable](gen genny.Generator[T], n int, opts ...Option) genny.Generator[T] {
	if n < 1 {
		panic("constrain: Duplicates needs a group size of at least 1")
	}
	return DuplicatesVar(gen, func() int { return n }, opts...)
}

// DuplicatesVar is like Duplicates with a group size that is asked at the start of every group, for
// instance from a generator's Generate method. A size below 1 counts as 1
func DuplicatesVar[T comparable](gen genny.Generator[T], size func() int, opts ...Option) genny.Generator[T] {
	o := newOptions(opts)
	seen := map[T]bool{}
	left := 0
	f := &filter[T]{gen: gen, stop: o.stop}
	f.accept = func(_ T, _ bool, cand T) bool {
		if left == 0 { // a new group
			clear(seen)
			left = max(size(), 1)
		}
		return !seen[cand]
	}
	f.onAccept = func(v T) {
		if left == 0 { // the first draw of a group was accepted by accept, which opened it
			clear(seen)
			left = max(size(), 1)
		}
		seen[v] = true
		left--
	}
	f.onReset = func() { clear(seen); left = 0 }
	return f
}

// Shotgun throws away a part of the values, at random: every value is dropped with probability p
// (0..1, clamped) and the next one is taken instead, so the stream gets thinner in content, not
// in length. After Stop drops in a row the last value is used. It can be used to thin out a pattern
// of values with an unpredictable hole pattern
func Shotgun[T any](gen genny.Generator[T], p float64, opts ...Option) genny.Generator[T] {
	o := newOptions(opts)
	if math.IsNaN(p) {
		p = 0
	}
	p = min(max(p, 0), 1)
	return &filter[T]{
		gen:    gen,
		stop:   o.stop,
		accept: func(T, bool, T) bool { return genny.Float64() >= p },
	}
}
