// Package rwalk provides RWalk, a random walk generator that stays within a range by clipping,
// bouncing or jumping at the edges.
package rwalk

import (
	"github.com/almerlucke/genny"
)

// Mode sets what happens when the walk leaves its range
type Mode int

const (
	Clip Mode = iota
	Bounce
	Jump
)

// RWalk
type RWalk struct {
	origin  float64
	current float64
	min     float64
	max     float64
	maxPct  float64
	mode    Mode
}

// New creates a random walk from origin within [min, max], each step moves at most maxPct
func New(origin, min, max, maxPct float64, mode Mode) *RWalk {
	if min > max {
		min, max = max, min
	}

	return &RWalk{
		origin:  origin,
		current: origin,
		min:     min,
		max:     max,
		maxPct:  maxPct,
		mode:    mode,
	}
}

func (rw *RWalk) Reset() {
	rw.current = rw.origin
}

func (rw *RWalk) Continuous() bool {
	return true
}

func (rw *RWalk) Done() bool {
	return false
}

func (rw *RWalk) Generate() float64 {
	rw.current += (genny.Float64()*2 - 1) * rw.maxPct

	if rw.max <= rw.min {
		// no room to walk (and bouncing or jumping would never end)
		rw.current = rw.min
		return rw.current
	}

	for {
		if rw.current < rw.min {
			if rw.mode == Clip {
				rw.current = rw.min
			} else if rw.mode == Bounce {
				rw.current = rw.min + (rw.min - rw.current)
				continue
			} else if rw.mode == Jump {
				rw.current = rw.max - (rw.min - rw.current)
				continue
			}
		}

		if rw.current > rw.max {
			if rw.mode == Clip {
				rw.current = rw.max
			} else if rw.mode == Bounce {
				rw.current = rw.max - (rw.current - rw.max)
				continue
			} else if rw.mode == Jump {
				rw.current = rw.min + (rw.current - rw.max)
				continue
			}
		}

		break
	}

	return rw.current
}
