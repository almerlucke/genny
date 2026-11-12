// Package melody makes melodies over chord progressions. ChordTones puts chord tones on the strong
// beats, fills the steps in between with scale tones that move by step, and ends every chord on a chord tone
package melody

import (
	"math"
	"sort"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/chord"
	"github.com/almerlucke/genny/note"
	"github.com/almerlucke/genny/scale"
)

// Config is the setup of the melody
type Config struct {
	// Scale gives the passing tones (required)
	Scale *scale.Scale
	// BarSteps is the number of steps in a bar, one chord per bar (default 16)
	BarSteps int
	// BeatSteps is the number of steps in a beat (default 4); BarSteps should be a multiple of it
	BeatSteps int
	// MaxLeap is the largest leap in scale degrees (default 3); chord tones further away than that are only
	// used when there is nothing nearer
	MaxLeap int
	// Low and High are the range of the melody in semitones, at least an octave apart (default 60 and 84)
	Low, High float64
}

func (c Config) withDefaults() Config {
	if c.BarSteps == 0 {
		c.BarSteps = 16
	}
	if c.BeatSteps == 0 {
		c.BeatSteps = 4
	}
	if c.MaxLeap == 0 {
		c.MaxLeap = 3
	}
	if c.Low == 0 && c.High == 0 {
		c.Low, c.High = 60, 84
	}
	return c
}

func (c Config) validate() {
	if c.Scale == nil {
		panic("melody: a scale is needed")
	}
	if c.BarSteps < 1 || c.BeatSteps < 1 || c.MaxLeap < 1 {
		panic("melody: bar steps, beat steps and max leap must be positive")
	}
	if !(c.High-c.Low >= 12) {
		panic("melody: the range must be at least an octave")
	}
}

// Weight is the metric importance of a step in a bar: 3 for the first step, 2 for the other beats, 1 for
// half beats and 0 elsewhere. Chord tones are placed on steps with weight 2 or more
func Weight(step, barSteps, beatSteps int) int {
	s := step % barSteps
	switch {
	case s == 0:
		return 3
	case s%beatSteps == 0:
		return 2
	case beatSteps%2 == 0 && s%(beatSteps/2) == 0:
		return 1
	}
	return 0
}

// ChordTones makes a melody of one chord per bar from the chord generator, for a number of bars. On the
// first step of each beat the melody plays a tone of the chord (the first step of the bar most
// strongly: the velocity follows the weight), between them it walks over the scale mostly by step, and
// the last step of a bar is again a chord tone, so every chord is left on one of its own tones. A
// chord tone is at most MaxLeap scale degrees from the note before it, if there is one. It stops
// early if the generator ends, and an empty chord gives a bar of rest. It panics if the Config is
// invalid (no scale, a range of less than an octave, non positive sizes). Seed with genny.Seed
func ChordTones(chords genny.Generator[chord.Chord], bars int, cfg Config) []note.Event {
	cfg = cfg.withDefaults()
	cfg.validate()
	return build(cfg, bars, func() (chord.Chord, bool) {
		if !chords.Continuous() && chords.Done() {
			return nil, false
		}
		return chords.Generate(), true
	})
}

// ChordTonesList is like ChordTones with a list of chords, which is repeated when there are more bars
// than chords. An empty list gives no melody
func ChordTonesList(chords []chord.Chord, bars int, cfg Config) []note.Event {
	cfg = cfg.withDefaults()
	cfg.validate()
	if len(chords) == 0 {
		return nil
	}
	i := 0
	return build(cfg, bars, func() (chord.Chord, bool) {
		c := chords[i%len(chords)]
		i++
		return c, true
	})
}

func build(cfg Config, bars int, next func() (chord.Chord, bool)) []note.Event {
	var out []note.Event
	prev := math.NaN()
	for bar := range max(bars, 0) {
		c, ok := next()
		if !ok {
			break
		}
		options := cfg.options(c)
		if len(options) == 0 {
			continue
		}

		// the steps that get a chord tone
		var targets []int
		for s := range cfg.BarSteps {
			if Weight(s, cfg.BarSteps, cfg.BeatSteps) >= 2 || s == cfg.BarSteps-1 {
				targets = append(targets, s)
			}
		}

		pitches := make([]float64, cfg.BarSteps)
		last := -1 // step of the last pitch placed
		for _, t := range targets {
			p := cfg.pickTone(options, prev)
			if last >= 0 {
				cfg.fill(pitches, last, t, p)
			}
			pitches[t] = p
			prev, last = p, t
		}

		for s, p := range pitches {
			w := Weight(s, cfg.BarSteps, cfg.BeatSteps)
			out = append(out, note.Event{
				Pitch:    p,
				Velocity: []float64{0.6, 0.7, 0.85, 0.95}[w],
				Start:    float64(bar*cfg.BarSteps + s),
				Duration: 1,
			})
		}
	}
	return out
}

// options are the pitches of the chord within the range, ascending
func (c Config) options(ch chord.Chord) []float64 {
	var out []float64
	seen := map[int]bool{}
	for _, p := range ch {
		pc := int(math.Round(p))
		pc = ((pc % 12) + 12) % 12
		if seen[pc] {
			continue
		}
		seen[pc] = true
		for q := pc; float64(q) <= c.High; q += 12 {
			if float64(q) >= c.Low {
				out = append(out, float64(q))
			}
		}
	}
	sort.Float64s(out)
	return out
}

// pickTone chooses a chord tone near the previous note: nearer is more likely, a repeat less so
func (c Config) pickTone(options []float64, prev float64) float64 {
	if math.IsNaN(prev) {
		mid := (c.Low + c.High) / 2
		return options[nearest(options, mid)]
	}
	dp := c.Scale.DegreeOf(prev)
	var cand []float64
	var weights []float64
	total := 0.0
	for _, o := range options {
		d := c.Scale.DegreeOf(o) - dp
		if d < 0 {
			d = -d
		}
		if d > c.MaxLeap {
			continue
		}
		w := 1 / float64(1+d)
		if d == 0 {
			w = 0.25
		}
		cand, weights = append(cand, o), append(weights, w)
		total += w
	}
	if len(cand) == 0 {
		return options[nearest(options, prev)]
	}
	r := genny.Float64() * total
	for i, w := range weights {
		if r < w {
			return cand[i]
		}
		r -= w
	}
	return cand[len(cand)-1]
}

func nearest(sorted []float64, x float64) int {
	best := 0
	for i, o := range sorted {
		if math.Abs(o-x) < math.Abs(sorted[best]-x) {
			best = i
		}
	}
	return best
}

// fill sets the steps between from and to (exclusive): scale tones that move by step towards the pitch at to
func (c Config) fill(pitches []float64, from, to int, target float64) {
	cur := c.Scale.DegreeOf(pitches[from])
	dt := c.Scale.DegreeOf(target)
	lo, hi := c.Scale.DegreeOf(c.Low), c.Scale.DegreeOf(c.High)
	for i := from + 1; i < to; i++ {
		left := to - i // moves still to make including the one into the target
		toward := 1
		if dt < cur {
			toward = -1
		}
		step := toward
		switch r := genny.Float64(); {
		case dt == cur:
			step = 1 - 2*genny.Intn(2)
		case r < 0.25:
			step = -toward // a turn away from the target
		case r < 0.35 && c.MaxLeap >= 2:
			step = 2 * toward
		}
		next := cur + step
		// stay in range and able to reach the target
		if next < lo || next > hi {
			next = cur - step
		}
		reach := left // mostly stepwise: every move left may cover one degree
		if d := dt - next; d > reach || -d > reach {
			next = cur + toward
		}
		cur = next
		pitches[i] = c.Scale.Degree(cur)
	}
}
