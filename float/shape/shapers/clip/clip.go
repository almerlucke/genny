// Package clip provides Clip, a shaper that limits its input with a hard clip between a low and a high
// value or a soft (arctan) clip.
package clip

import "math"

const arctanMult = 2.0 / math.Pi

type Mode int

const (
	Hard Mode = iota
	Soft
)

type Clip struct {
	norm float64
	a    float64
	low  float64
	high float64
	mode Mode
}

func New(m Mode, args ...float64) *Clip {
	c := &Clip{
		mode: m,
	}

	switch m {
	case Hard:
		if len(args) < 2 {
			panic("clip: Hard needs a low and a high bound")
		}
		c.low = args[0]
		c.high = args[1]
	case Soft:
		if len(args) < 1 {
			panic("clip: Soft needs a gain")
		}
		c.norm = 1.0 / math.Atan(args[0])
		c.a = args[0]
	}

	return c
}

func (c *Clip) Shape(x float64) float64 {
	switch c.mode {
	case Hard:
		if x > c.high {
			x = c.high
		}
		if x < c.low {
			x = c.low
		}
	case Soft:
		x = c.norm * math.Atan(c.a*x)
	}

	return x
}
