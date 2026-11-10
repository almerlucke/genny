// Package groove makes timing offsets (swing, humanize) and velocity jitter. Offsets are in
// fractions of a step, the caller adds them to the step time
package groove

import (
	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/transform"
)

// Swing delays every other step. Generate returns the offset of the next step: 0 for even steps,
// amount for odd steps (1/3 is a triplet feel), plus optional humanize jitter
type Swing struct {
	amount   float64
	humanize float64
	step     int
}

// NewSwing creates a swing generator, amount is the delay of odd steps in fractions of a step
func NewSwing(amount float64) *Swing { return &Swing{amount: amount} }

// SetAmount changes the swing amount
func (s *Swing) SetAmount(amount float64) { s.amount = amount }

// Humanize adds a random offset in [-h, h] to every step
func (s *Swing) Humanize(h float64) *Swing {
	s.humanize = h
	return s
}

func (s *Swing) Generate() float64 {
	off := 0.0
	if s.step%2 == 1 {
		off = s.amount
	}
	s.step++
	if s.humanize != 0 {
		off += (genny.Float64()*2 - 1) * s.humanize
	}
	return off
}

func (s *Swing) Continuous() bool { return true }

func (s *Swing) Done() bool { return false }

func (s *Swing) Reset() { s.step = 0 }

// Humanize generates random offsets in [-amount, amount]
type Humanize struct {
	amount float64
}

// NewHumanize creates a humanize generator
func NewHumanize(amount float64) *Humanize { return &Humanize{amount: amount} }

// SetAmount changes the amount
func (h *Humanize) SetAmount(amount float64) { h.amount = amount }

func (h *Humanize) Generate() float64 { return (genny.Float64()*2 - 1) * h.amount }

func (h *Humanize) Continuous() bool { return true }

func (h *Humanize) Done() bool { return false }

func (h *Humanize) Reset() {}

// Jitter adds a random value in [-amount, amount] to every value of gen, for instance to vary velocities
func Jitter(gen genny.Generator[float64], amount float64) *transform.Transform[float64] {
	return transform.New(gen, func(v float64) float64 {
		return v + (genny.Float64()*2-1)*amount
	})
}

// Accent generates velocities (accents) that follow the beat groups of a meter: the first step of the bar is
// the strongest, the first step of every other group is medium and the rest are weak
type Accent struct {
	groups             []int
	down, group, other float64
	current, offset    int // the beat group and the step in it
}

// NewAccent creates an accent generator for a bar with groups of steps: 7/8 as (2, 2, 3), 3/4 as 4 groups of
// (4, 4, 4) sixteenths, 4/4 in eighths as (2, 2, 2, 2). Levels default to 1, 0.8 and 0.55, see Levels.
// It panics if there are no groups or a group is shorter than 1
func NewAccent(groups ...int) *Accent {
	if len(groups) == 0 {
		panic("groove: an accent needs beat groups")
	}
	for _, g := range groups {
		if g < 1 {
			panic("groove: a beat group must be at least 1 step")
		}
	}
	return &Accent{groups: append([]int{}, groups...), down: 1, group: 0.8, other: 0.55}
}

// Levels sets the velocity of the first step of the bar, of the first step of the other groups, and of all other steps
func (a *Accent) Levels(downbeat, group, other float64) *Accent {
	a.down, a.group, a.other = downbeat, group, other
	return a
}

func (a *Accent) Generate() float64 {
	v := a.other
	if a.offset == 0 {
		v = a.group
		if a.current == 0 {
			v = a.down
		}
	}

	a.offset++
	if a.offset >= a.groups[a.current] {
		a.offset = 0
		a.current = (a.current + 1) % len(a.groups)
	}
	return v
}

func (a *Accent) Continuous() bool { return true }

func (a *Accent) Done() bool { return false }

func (a *Accent) Reset() { a.current, a.offset = 0, 0 }
