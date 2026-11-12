package plot

import (
	"errors"
	"image"
	"image/color"
	"math"
)

// LineStyle is the colour and width of a line
type LineStyle struct {
	Color color.Color
	Width Length
}

func checkXYs(xys XYer) error {
	if xys == nil || xys.Len() == 0 {
		return errors.New("plot: no points")
	}
	for i := 0; i < xys.Len(); i++ {
		x, y := xys.XY(i)
		if math.IsNaN(x) || math.IsInf(x, 0) || math.IsNaN(y) || math.IsInf(y, 0) {
			return errors.New("plot: point is not finite")
		}
	}
	return nil
}

func rangeOf(lists ...XYer) (xmin, xmax, ymin, ymax float64, ok bool) {
	xmin, ymin, xmax, ymax = math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for _, xys := range lists {
		for i := 0; i < xys.Len(); i++ {
			x, y := xys.XY(i)
			xmin, xmax, ymin, ymax = math.Min(xmin, x), math.Max(xmax, x), math.Min(ymin, y), math.Max(ymax, y)
			ok = true
		}
	}
	return
}

// Line connects points with straight lines
type Line struct {
	XYer
	LineStyle
}

// NewLine returns a black line of 1 point through the points; it fails if there are none or if one
// is not finite
func NewLine(xys XYer) (*Line, error) {
	if err := checkXYs(xys); err != nil {
		return nil, err
	}
	return &Line{XYer: xys, LineStyle: LineStyle{Color: color.Black, Width: Points(1)}}, nil
}

func (l *Line) dataRange() (a, b, d, e float64, ok bool) { return rangeOf(l.XYer) }

func (l *Line) draw(c *canvas, m *mapping) {
	pts := make([]pt, l.Len())
	for i := range pts {
		x, y := l.XY(i)
		pts[i] = pt{m.x(x), m.y(y)}
	}
	c.polyline(pts, float64(l.Width)*m.scale, l.Color)
}

func (l *Line) thumbnail(c *canvas, r image.Rectangle) {
	y := float64(r.Min.Y+r.Max.Y) / 2
	c.segment(pt{float64(r.Min.X), y}, pt{float64(r.Max.X), y}, math.Max(float64(l.Width)*dpi/72*supersample, 2), l.Color)
}

// Glyph draws a point of a scatter plot
type Glyph interface {
	drawGlyph(c *canvas, x, y, radius float64, col color.Color)
}

// CircleGlyph is a filled circle
type CircleGlyph struct{}

func (CircleGlyph) drawGlyph(c *canvas, x, y, r float64, col color.Color) { c.disc(x, y, r, col) }

// BoxGlyph is a filled square
type BoxGlyph struct{}

func (BoxGlyph) drawGlyph(c *canvas, x, y, r float64, col color.Color) {
	c.fillRect(image.Rect(int(math.Round(x-r)), int(math.Round(y-r)), int(math.Round(x+r)), int(math.Round(y+r))), col)
}

// GlyphStyle is the shape, size and colour of the points of a scatter plot
type GlyphStyle struct {
	Color  color.Color
	Radius Length
	Shape  Glyph
}

// Scatter draws a glyph for every point
type Scatter struct {
	XYer
	GlyphStyle
}

// NewScatter returns black dots of 2.5 points radius at the points
func NewScatter(xys XYer) (*Scatter, error) {
	if err := checkXYs(xys); err != nil {
		return nil, err
	}
	return &Scatter{XYer: xys, GlyphStyle: GlyphStyle{Color: color.Black, Radius: Points(2.5), Shape: CircleGlyph{}}}, nil
}

func (s *Scatter) dataRange() (a, b, d, e float64, ok bool) { return rangeOf(s.XYer) }

func (s *Scatter) draw(c *canvas, m *mapping) {
	r := math.Max(float64(s.Radius)*m.scale, 0.75)
	for i := 0; i < s.Len(); i++ {
		x, y := s.XY(i)
		s.Shape.drawGlyph(c, m.x(x), m.y(y), r, s.Color)
	}
}

func (s *Scatter) thumbnail(c *canvas, r image.Rectangle) {
	s.Shape.drawGlyph(c, float64(r.Min.X+r.Max.X)/2, float64(r.Min.Y+r.Max.Y)/2, math.Max(float64(s.Radius)*dpi/72*supersample, 3), s.Color)
}

// Polygon draws filled polygons
type Polygon struct {
	XYs       []XYer
	Color     color.Color
	LineStyle LineStyle
}

// NewPolygon returns grey polygons, one for each list of points (at least three points each)
func NewPolygon(polys ...XYer) (*Polygon, error) {
	if len(polys) == 0 {
		return nil, errors.New("plot: no polygons")
	}
	for _, p := range polys {
		if err := checkXYs(p); err != nil {
			return nil, err
		}
		if p.Len() < 3 {
			return nil, errors.New("plot: a polygon needs three points")
		}
	}
	return &Polygon{XYs: polys, Color: color.Gray{128}, LineStyle: LineStyle{Color: color.Black, Width: Points(1)}}, nil
}

func (p *Polygon) dataRange() (a, b, d, e float64, ok bool) { return rangeOf(p.XYs...) }

func (p *Polygon) draw(c *canvas, m *mapping) {
	for _, xys := range p.XYs {
		pts := make([]pt, xys.Len())
		for i := range pts {
			x, y := xys.XY(i)
			pts[i] = pt{m.x(x), m.y(y)}
		}
		if p.Color != nil {
			c.fillPoly(pts, p.Color)
		}
		if p.LineStyle.Color != nil && p.LineStyle.Width > 0 {
			c.polyline(append(pts, pts[0]), float64(p.LineStyle.Width)*m.scale, p.LineStyle.Color)
		}
	}
}

func (p *Polygon) thumbnail(c *canvas, r image.Rectangle) {
	if p.Color != nil {
		c.fillRect(r, p.Color)
	}
}

// Grid draws light lines at the major ticks
type Grid struct{}

// NewGrid returns a grid
func NewGrid() *Grid { return &Grid{} }

func (*Grid) dataRange() (a, b, d, e float64, ok bool) { return 0, 0, 0, 0, false }

func (*Grid) draw(c *canvas, m *mapping) {
	w := float64(supersample)
	for _, t := range m.xticks {
		if t.Label != "" {
			c.segment(pt{m.x(t.Value), float64(m.area.Min.Y)}, pt{m.x(t.Value), float64(m.area.Max.Y)}, w, gridColor)
		}
	}
	for _, t := range m.yticks {
		if t.Label != "" {
			c.segment(pt{float64(m.area.Min.X), m.y(t.Value)}, pt{float64(m.area.Max.X), m.y(t.Value)}, w, gridColor)
		}
	}
}
