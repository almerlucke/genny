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
