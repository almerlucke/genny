// Package mirror provides Mirror, a shaper that reflects its input back into a range.
package mirror

import "math"

type Mirror struct {
	Bottom float64
	Top    float64
}

func New(bottom float64, top float64) *Mirror {
	return &Mirror{
		Bottom: bottom,
		Top:    top,
	}
}

func (mirror *Mirror) Shape(x float64) float64 {
	bottom, top := mirror.Bottom, mirror.Top
	if bottom > top {
		bottom, top = top, bottom
	}

	if x >= bottom && x <= top {
		return x
	}

	width := top - bottom
	if width == 0 {
		return bottom
	}

	// fold x back and forth between bottom and top
	p := math.Mod(x-bottom, 2*width)
	if p < 0 {
		p += 2 * width
	}
	if p > width {
		p = 2*width - p
	}

	return bottom + p
}
