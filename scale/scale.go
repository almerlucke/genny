// Package scale holds scales, quantizes values to a scale and makes melodic walks over scale degrees.
// Pitches are in semitones (for instance MIDI notes), a scale repeats every Period (12)
package scale

import (
	"math"

	"github.com/almerlucke/genny"
)

// Steps of some common scales, in semitones above the root
var (
	Major           = []float64{0, 2, 4, 5, 7, 9, 11}
	Minor           = []float64{0, 2, 3, 5, 7, 8, 10}
	HarmonicMinor   = []float64{0, 2, 3, 5, 7, 8, 11}
	MelodicMinor    = []float64{0, 2, 3, 5, 7, 9, 11}
	Dorian          = []float64{0, 2, 3, 5, 7, 9, 10}
	Phrygian        = []float64{0, 1, 3, 5, 7, 8, 10}
	Lydian          = []float64{0, 2, 4, 6, 7, 9, 11}
	Mixolydian      = []float64{0, 2, 4, 5, 7, 9, 10}
	Locrian         = []float64{0, 1, 3, 5, 6, 8, 10}
	MajorPentatonic = []float64{0, 2, 4, 7, 9}
	MinorPentatonic = []float64{0, 3, 5, 7, 10}
	Blues           = []float64{0, 3, 5, 6, 7, 10}
	WholeTone       = []float64{0, 2, 4, 6, 8, 10}
	Diminished      = []float64{0, 2, 3, 5, 6, 8, 9, 11}
	Chromatic       = []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}
)

// Scale is a root and ascending steps within one period
type Scale struct {
	root   float64
	steps  []float64
	period float64
}

// New creates a scale with an octave of 12 semitones. Steps must ascend and lie in [0, 12)
func New(root float64, steps []float64) *Scale {
	return NewPeriod(root, steps, 12)
}

// NewPeriod creates a scale that repeats every period, for non-octave or microtonal scales
func NewPeriod(root float64, steps []float64, period float64) *Scale {
	if len(steps) == 0 {
		steps = []float64{0}
	}
	return &Scale{root: root, steps: append([]float64{}, steps...), period: period}
}

// Root of the scale
func (s *Scale) Root() float64 { return s.root }

// Len is the number of degrees in a period
func (s *Scale) Len() int { return len(s.steps) }

// Degree returns the pitch of a scale degree, 0 is the root, Len() is the root an octave up and
// negative degrees go down
func (s *Scale) Degree(i int) float64 {
	n := len(s.steps)
	oct := floorDiv(i, n)
	return s.root + float64(oct)*s.period + s.steps[i-oct*n]
}

// DegreeOf returns the degree nearest to a pitch
func (s *Scale) DegreeOf(x float64) int {
	rel := x - s.root
	oct := math.Floor(rel / s.period)
	r := rel - oct*s.period

	best, bestDist := 0, math.Inf(1)
	for i, st := range s.steps {
		if d := math.Abs(r - st); d < bestDist {
			best, bestDist = i, d
		}
	}
	// the root of the next period can be closer than the last step
	if math.Abs(s.period-r) < bestDist {
		best = len(s.steps)
	}
	return int(oct)*len(s.steps) + best
}

// Quantize returns the pitch of the scale nearest to x, a tie goes down
func (s *Scale) Quantize(x float64) float64 { return s.Degree(s.DegreeOf(x)) }

// Shape makes a Scale a shape.Shaper, so any float generator can be quantized to the scale
// with shape.NewSingle
func (s *Scale) Shape(x float64) float64 { return s.Quantize(x) }

func floorDiv(a, b int) int {
	q := a / b
	if a%b != 0 && (a < 0) != (b < 0) {
		q--
	}
	return q
}

// Degrees turns a generator of degrees into a generator of pitches
type Degrees struct {
	scale *Scale
	gen   genny.Generator[int]
}

// NewDegrees creates a generator of pitches from degrees
func NewDegrees(s *Scale, gen genny.Generator[int]) *Degrees { return &Degrees{scale: s, gen: gen} }

func (d *Degrees) Generate() float64 { return d.scale.Degree(d.gen.Generate()) }

func (d *Degrees) Continuous() bool { return d.gen.Continuous() }

func (d *Degrees) Done() bool { return d.gen.Done() }

func (d *Degrees) Reset() { d.gen.Reset() }

// Walk is a melody that moves over scale degrees: mostly by step, sometimes by a leap
type Walk struct {
	scale    *Scale
	start    int
	current  int
	lo, hi   int
	maxStep  int
	leapProb float64
	maxLeap  int
}

// NewWalk creates a walk starting at degree start and staying within degrees [lo, hi] (bouncing
// off the edges). Every step moves 1..maxStep degrees up or down, with probability leapProb it is a
// leap of maxStep+1..maxLeap degrees instead
func NewWalk(s *Scale, start, lo, hi, maxStep int, leapProb float64, maxLeap int) *Walk {
	maxStep = max(maxStep, 1)
	return &Walk{
		scale: s, start: start, current: start, lo: lo, hi: hi,
		maxStep: maxStep, leapProb: leapProb, maxLeap: max(maxLeap, maxStep),
	}
}

// Degree is the current degree
func (w *Walk) Degree() int { return w.current }

func (w *Walk) Generate() float64 {
	size := 1 + genny.Intn(w.maxStep)
	if w.maxLeap > w.maxStep && genny.Float64() < w.leapProb {
		size = w.maxStep + 1 + genny.Intn(w.maxLeap-w.maxStep)
	}
	if genny.Float64() < 0.5 {
		size = -size
	}

	d := w.current + size
	if d > w.hi {
		d = 2*w.hi - d
	}
	if d < w.lo {
		d = 2*w.lo - d
	}
	w.current = min(max(d, w.lo), w.hi)

	return w.scale.Degree(w.current)
}

func (w *Walk) Continuous() bool { return true }

func (w *Walk) Done() bool { return false }

func (w *Walk) Reset() { w.current = w.start }
