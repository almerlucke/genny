// Package lfo has a low frequency oscillator with periodic and random shapes, and a slew limiter
package lfo

import (
	"math"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/float/phasor"
)

// Shape of an LFO
type Shape int

const (
	Sine Shape = iota
	Triangle
	// Saw rises from -1 to 1
	Saw
	// Square is 1 for the first PulseWidth of the cycle and -1 after
	Square
	// SampleHold holds a new random value for every cycle
	SampleHold
	// SmoothRandom glides between a new random value every cycle
	SmoothRandom
)

// LFO generates values in [-1, 1] (or [0, 1] if unipolar)
type LFO struct {
	shape      Shape
	phasor     *phasor.Phasor
	pulseWidth float64
	unipolar   bool
	prevPhase  float64
	from, to   float64
}

// New creates an LFO of frequency fc at sample (or step) rate fs
func New(shape Shape, fc, fs float64) *LFO {
	l := &LFO{shape: shape, phasor: phasor.New(fc, fs, 0), pulseWidth: 0.5}
	l.Reset()
	return l
}

// SetFrequency changes the frequency
func (l *LFO) SetFrequency(fc, fs float64) { l.phasor.SetFrequency(fc, fs) }

// SetShape changes the shape
func (l *LFO) SetShape(shape Shape) { l.shape = shape }

// SetPulseWidth sets the width of the Square shape in [0, 1]
func (l *LFO) SetPulseWidth(pw float64) { l.pulseWidth = pw }

// SetUnipolar makes the output range [0, 1] instead of [-1, 1]
func (l *LFO) SetUnipolar(u bool) *LFO {
	l.unipolar = u
	return l
}

func randBipolar() float64 { return genny.Float64()*2 - 1 }

func (l *LFO) Generate() float64 {
	p := l.phasor.Generate()
	if p < l.prevPhase {
		l.from, l.to = l.to, randBipolar()
	}
	l.prevPhase = p

	var v float64
	switch l.shape {
	case Sine:
		v = math.Sin(2 * math.Pi * p)
	case Triangle:
		v = 1 - 4*math.Abs(p-0.5)
	case Saw:
		v = 2*p - 1
	case Square:
		if p < l.pulseWidth {
			v = 1
		} else {
			v = -1
		}
	case SampleHold:
		v = l.from
	case SmoothRandom:
		v = l.from + (l.to-l.from)*p*p*(3-2*p)
	}

	if l.unipolar {
		v = (v + 1) / 2
	}
	return v
}

func (l *LFO) Continuous() bool { return true }

func (l *LFO) Done() bool { return false }

func (l *LFO) Reset() {
	l.phasor.Reset()
	l.prevPhase = -1
	l.from, l.to = randBipolar(), randBipolar()
}

// Slew limits how fast a generator can change: at most rise per step up and fall per step down.
// Use it to smooth steps from SampleHold or any other stepped source
type Slew struct {
	gen        genny.Generator[float64]
	rise, fall float64
	current    float64
	started    bool
}

// NewSlew creates a slew limiter, rise and fall are the maximum change per step (> 0)
func NewSlew(gen genny.Generator[float64], rise, fall float64) *Slew {
	return &Slew{gen: gen, rise: rise, fall: fall}
}

func (s *Slew) Generate() float64 {
	target := s.gen.Generate()
	if !s.started {
		s.started = true
		s.current = target
		return target
	}
	if d := target - s.current; d > s.rise {
		s.current += s.rise
	} else if d < -s.fall {
		s.current -= s.fall
	} else {
		s.current = target
	}
	return s.current
}

func (s *Slew) Continuous() bool { return s.gen.Continuous() }

func (s *Slew) Done() bool { return s.gen.Done() }

func (s *Slew) Reset() {
	s.gen.Reset()
	s.started = false
}
