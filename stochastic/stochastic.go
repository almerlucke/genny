// Package stochastic makes music from chance in the manner of Iannis Xenakis: events whose onsets form a
// Poisson process (exponential gaps between them), clouds of events whose density changes over time and
// whose pitch, velocity and duration are drawn from ranges or distributions, and gates whose chance of a hit
// follows a density. All draws come from the random stream of package genny, so a seed repeats a piece
package stochastic

import (
	"math"
	"slices"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/float/dist"
	"github.com/almerlucke/genny/note"
	"github.com/almerlucke/genny/scale"
)

// maxEvents limits the number of events of one call, so an enormous rate cannot run forever
const maxEvents = 1 << 22

// Poisson returns the onset times of a Poisson process with the given rate (events per step) in [0, length):
// the gaps between events are exponentially distributed with mean 1/rate. The times are sorted. A rate or
// length of zero or less gives no events (at most 4 million events are returned). It panics if rate or
// length is NaN or infinite
func Poisson(rate, length float64) []float64 {
	if math.IsNaN(rate) || math.IsInf(rate, 0) || math.IsNaN(length) || math.IsInf(length, 0) {
		panic("stochastic: rate and length must be finite")
	}
	var out []float64
	if rate <= 0 || length <= 0 {
		return out
	}
	for t := expGap(rate); t < length && len(out) < maxEvents; t += expGap(rate) {
		out = append(out, t)
	}
	return out
}

// expGap draws an exponential gap with mean 1/rate (1-u keeps the logarithm finite)
func expGap(rate float64) float64 { return -math.Log(1-genny.Float64()) / rate }

// PoissonInt draws a Poisson distributed count with the given mean: how many events fall in a step when
// they come at random. A mean of zero or less (or NaN) gives 0, a mean above 1e6 is clamped to 1e6
func PoissonInt(mean float64) int {
	if !(mean > 0) {
		return 0
	}
	mean = min(mean, 1e6)
	// Knuth's method, in chunks of at most 30 because a Poisson count is the sum of the counts of
	// parts of its mean, and exp(-mean) would underflow for a large mean
	n := 0
	for mean > 0 {
		m := min(mean, 30)
		mean -= m
		l, p := math.Exp(-m), 1.0
		for {
			p *= genny.Float64()
			if p <= l {
				break
			}
			n++
		}
	}
	return n
}

// Config describes a Cloud
type Config struct {
	// Density is the mean number of events in a step, as a function of the step (0, 1, 2, ...).
	// Required, see FromGenerator to use a generator
	Density func(step int) float64
	// Pitch generates the pitch of an event, in semitones. Required, see Uniform for a range
	Pitch genny.Generator[float64]
	// Velocity generates the velocity in [0, 1] (default 0.8)
	Velocity genny.Generator[float64]
	// Duration generates the length of an event in steps (default 1)
	Duration genny.Generator[float64]
	// Scale, if set, quantizes the pitches to the scale
	Scale *scale.Scale
}

// Cloud is a stochastic cloud of events: in every step the number of events is Poisson distributed around
// the density of that step, their onsets are uniform within the step
type Cloud struct{ cfg Config }

// NewCloud creates a cloud. It panics if Density or Pitch is missing
func NewCloud(cfg Config) *Cloud {
	if cfg.Density == nil || cfg.Pitch == nil {
		panic("stochastic: a cloud needs a density and a pitch generator")
	}
	return &Cloud{cfg: cfg}
}

// Render plays the cloud for a number of steps and returns the events, ordered by start. Steps of zero or
// less give no events. Velocities are clamped to [0, 1] and durations are at least 0.001
func (c *Cloud) Render(steps int) []note.Event {
	var events []note.Event
	for step := range steps {
		n := PoissonInt(c.cfg.Density(step))
		starts := make([]float64, n)
		for i := range starts {
			starts[i] = float64(step) + genny.Float64()
		}
		slices.Sort(starts)
		for _, s := range starts {
			e := note.Event{Pitch: c.cfg.Pitch.Generate(), Velocity: 0.8, Start: s, Duration: 1}
			if c.cfg.Scale != nil {
				e.Pitch = c.cfg.Scale.Quantize(e.Pitch)
			}
			if c.cfg.Velocity != nil {
				e.Velocity = min(max(c.cfg.Velocity.Generate(), 0), 1)
			}
			if c.cfg.Duration != nil {
				e.Duration = max(c.cfg.Duration.Generate(), 1e-3)
			}
			events = append(events, e)
		}
	}
	return events
}

// FromGenerator makes a density function of a generator: it is asked once for every step, in order
func FromGenerator(g genny.Generator[float64]) func(step int) float64 {
	return func(int) float64 { return g.Generate() }
}

// Uniform generates values spread evenly in [lo, hi) (the bounds may be given in any order)
func Uniform(lo, hi float64) genny.Generator[float64] {
	if lo > hi {
		lo, hi = hi, lo
	}
	return &uniform{lo, hi}
}

type uniform struct{ lo, hi float64 }

func (u *uniform) Generate() float64 { return u.lo + (u.hi-u.lo)*genny.Float64() }
func (u *uniform) Continuous() bool  { return true }
func (u *uniform) Done() bool        { return false }
func (u *uniform) Reset()            {}

// Density is a gate whose chance of true in a step follows a density: a thinning or thickening rhythm
// of random hits, for rest.Gated. A density of 0 or less never hits, 1 or more always
type Density struct{ gen genny.Generator[float64] }

// NewDensity creates a gate from a generator of probabilities in [0, 1], one value per step
func NewDensity(gen genny.Generator[float64]) *Density { return &Density{gen: gen} }

func (d *Density) Generate() bool {
	p := d.gen.Generate()
	return genny.Float64() < p // NaN is never true
}

func (d *Density) Continuous() bool { return d.gen.Continuous() }
func (d *Density) Done() bool       { return d.gen.Done() }
func (d *Density) Reset()           { d.gen.Reset() }

// Free makes the events of one section of "free stochastic music": Poisson onsets with a density of events
// per step, exponentially distributed durations with mean meanDuration and Gaussian pitches around
// pitchMean with a spread of pitchSpread semitones, optionally quantized to a scale (nil for none). The
// velocity is uniform in [0.3, 0.9]. Xenakis derived all parameters of his stochastic pieces from such
// distributions. It panics if meanDuration is not positive or pitchSpread is negative
func Free(length, density, meanDuration, pitchMean, pitchSpread float64, s *scale.Scale) []note.Event {
	durations := dist.NewExponential(meanDuration)
	pitches := dist.NewGaussian(pitchMean, pitchSpread)
	starts := Poisson(density, length)
	events := make([]note.Event, len(starts))
	for i, t := range starts {
		p := pitches.Generate()
		if s != nil {
			p = s.Quantize(p)
		}
		events[i] = note.Event{
			Pitch: p, Velocity: 0.3 + 0.6*genny.Float64(), Start: t, Duration: max(durations.Generate(), 1e-3),
		}
	}
	return events
}
