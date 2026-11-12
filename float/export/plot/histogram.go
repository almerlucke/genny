package plot

import (
	"errors"
	"image"
	"image/color"
	"math"
	"strconv"
)

var barOutline = LineStyle{Color: color.White, Width: Points(0.5)}

// Histogram draws the distribution of values as bars
type Histogram struct {
	Values []float64
	Bins   int
	// Density scales the bars so that their area is 1 (a probability density) instead of counts
	Density bool
	// Min and Max fix the range of the bins when Min < Max; values outside are not counted. By
	// default the range is that of the values
	Min, Max float64
	Color    color.Color
	// LineStyle is the outline of the bars, white by default so adjacent bars are told apart
	LineStyle LineStyle
}

// NewHistogram returns a histogram of the finite values with a number of equally wide bins. It fails
// if bins is less than 1 or there is no finite value. Set Density, Min and Max on the result
func NewHistogram(values []float64, bins int) (*Histogram, error) {
	if bins < 1 {
		return nil, errors.New("plot: a histogram needs at least one bin")
	}
	for _, v := range values {
		if !math.IsNaN(v) && !math.IsInf(v, 0) {
			return &Histogram{Values: values, Bins: bins, Color: Color(0), LineStyle: barOutline}, nil
		}
	}
	return nil, errors.New("plot: a histogram needs at least one finite value")
}

// Counts returns the bin edges (Bins+1 of them) and the number of values in every bin. A value on
// the upper edge of the range counts in the last bin. Without a fixed range the counts add up to
// the number of finite values
func (h *Histogram) Counts() (edges []float64, counts []int) {
	lo, hi := h.Min, h.Max
	if !(lo < hi) || math.IsInf(lo, 0) || math.IsInf(hi, 0) {
		lo, hi = math.Inf(1), math.Inf(-1)
		for _, v := range h.Values {
			if !math.IsNaN(v) && !math.IsInf(v, 0) {
				lo, hi = math.Min(lo, v), math.Max(hi, v)
			}
		}
		if lo > hi {
			lo, hi = 0, 1
		}
		if lo == hi {
			lo, hi = lo-0.5, hi+0.5
		}
	}
	bins := max(h.Bins, 1)
	edges = make([]float64, bins+1)
	for i := range edges {
		edges[i] = lo + (hi-lo)*float64(i)/float64(bins)
	}
	counts = make([]int, bins)
	for _, v := range h.Values {
		if math.IsNaN(v) || v < lo || v > hi {
			continue
		}
		counts[min(int((v-lo)/(hi-lo)*float64(bins)), bins-1)]++
	}
	return edges, counts
}

// heights are the bar heights: counts or densities
func (h *Histogram) heights() (edges, heights []float64) {
	edges, counts := h.Counts()
	total := 0
	for _, c := range counts {
		total += c
	}
	heights = make([]float64, len(counts))
	for i, c := range counts {
		heights[i] = float64(c)
		if h.Density && total > 0 {
			heights[i] /= float64(total) * (edges[i+1] - edges[i])
		}
	}
	return edges, heights
}

func (h *Histogram) dataRange() (a, b, d, e float64, ok bool) {
	edges, hs := h.heights()
	top := 0.0
	for _, v := range hs {
		top = math.Max(top, v)
	}
	return edges[0], edges[len(edges)-1], 0, top, true
}

func (h *Histogram) zeroBased() bool { return true }

func (h *Histogram) draw(c drawer, m *mapping) {
	edges, hs := h.heights()
	drawBars(c, m, edges[:len(hs)], edges[1:], hs, h.Color, h.LineStyle)
}

func (h *Histogram) thumbnail(c drawer, r image.Rectangle) { c.fillRect(r, h.Color) }

// drawBars draws bars from x0[i] to x1[i] and from zero to height[i]
func drawBars(c drawer, m *mapping, x0, x1, height []float64, fill color.Color, line LineStyle) {
	zero := m.y(0)
	if m.ylog {
		zero = float64(m.area.Max.Y)
	}
	for i, v := range height {
		pts := []pt{{m.x(x0[i]), zero}, {m.x(x1[i]), zero}, {m.x(x1[i]), m.y(v)}, {m.x(x0[i]), m.y(v)}}
		if fill != nil {
			c.fillPoly(pts, fill)
		}
		if line.Color != nil && line.Width > 0 {
			c.polyline(append(pts, pts[0]), float64(line.Width)*m.scale, line.Color)
		}
	}
}

// Bars draws categorical bars: one bar for every label, from zero to its value. It is also a
// TickMarker that puts the labels under the bars, and the plot uses it for the x axis unless that
// has a marker of its own
type Bars struct {
	Labels []string
	Values []float64
	// Width of a bar, 0 < Width <= 1 (the distance between bars is 1)
	Width     float64
	Color     color.Color
	LineStyle LineStyle
}

// NewBars returns bars of values with a label under each. It fails if there are no values, if the
// number of labels differs from the number of values (no labels at all is allowed) or if a value is
// not finite
func NewBars(labels []string, values []float64) (*Bars, error) {
	if len(values) == 0 {
		return nil, errors.New("plot: no bars")
	}
	if len(labels) != 0 && len(labels) != len(values) {
		return nil, errors.New("plot: the number of labels differs from the number of values")
	}
	for _, v := range values {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, errors.New("plot: bar value is not finite")
		}
	}
	return &Bars{Labels: labels, Values: values, Width: 0.8, Color: Color(0), LineStyle: barOutline}, nil
}

// Ticks puts a labelled tick under every bar (bar i is at x = i); without labels the numbers are used
func (b *Bars) Ticks(min, max float64) []Tick {
	var ticks []Tick
	for i := range b.Values {
		label := strconv.Itoa(i)
		if len(b.Labels) > 0 {
			label = b.Labels[i]
		}
		ticks = append(ticks, Tick{Value: float64(i), Label: label})
	}
	return ticks
}

func (b *Bars) dataRange() (a, c, d, e float64, ok bool) {
	d, e = 0, 0
	for _, v := range b.Values {
		d, e = math.Min(d, v), math.Max(e, v)
	}
	return -0.5, float64(len(b.Values)) - 0.5, d, e, true
}

func (b *Bars) zeroBased() bool { return true }

func (b *Bars) draw(c drawer, m *mapping) {
	w := math.Min(math.Max(b.Width, 0.01), 1) / 2
	x0, x1 := make([]float64, len(b.Values)), make([]float64, len(b.Values))
	for i := range b.Values {
		x0[i], x1[i] = float64(i)-w, float64(i)+w
	}
	drawBars(c, m, x0, x1, b.Values, b.Color, b.LineStyle)
}

func (b *Bars) thumbnail(c drawer, r image.Rectangle) { c.fillRect(r, b.Color) }
