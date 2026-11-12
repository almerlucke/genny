// Package plot draws line plots, scatter plots, filled polygons and piano rolls to PNG files, and
// plots the output of a FrameGenerator. It uses nothing but the standard library: the drawing, the
// anti-aliasing and the bitmap font are in the package itself, so it brings no dependencies. Its
// types follow the shape of gonum's plot package (a Plot with axes, a Legend, plotters added with
// Add) in a much smaller form.
package plot

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/almerlucke/genny/float"
)

// Length is a length in points (1/72 inch)
type Length float64

const (
	Point      Length = 1
	Inch       Length = 72
	Centimeter        = Inch / 2.54
)

// Points returns n points
func Points(n float64) Length { return Length(n) * Point }

// dpi is the resolution of the PNG files
const dpi = 120

// XY is a point
type XY struct{ X, Y float64 }

// XYer is a list of points
type XYer interface {
	Len() int
	XY(i int) (x, y float64)
}

// XYs is a slice of points
type XYs []XY

func (s XYs) Len() int                { return len(s) }
func (s XYs) XY(i int) (x, y float64) { return s[i].X, s[i].Y }

// Tick is a mark on an axis, with an optional label
type Tick struct {
	Value float64
	Label string
}

// TickMarker chooses the ticks of an axis between min and max
type TickMarker interface {
	Ticks(min, max float64) []Tick
}

// Axis is the x or the y axis. Min and Max are taken from the data unless both are set
type Axis struct {
	Label struct{ Text string }
	Min   float64
	Max   float64
	Tick  struct{ Marker TickMarker }
}

// Legend lists the plotters that were added to it
type Legend struct {
	Top, Left bool
	entries   []legendEntry
}

type legendEntry struct {
	name  string
	thumb Thumbnailer
}

// Thumbnailer is something that can draw a small sample of itself for the legend
type Thumbnailer interface {
	thumbnail(c *canvas, r image.Rectangle)
}

// Add adds an entry to the legend
func (l *Legend) Add(name string, t Thumbnailer) { l.entries = append(l.entries, legendEntry{name, t}) }

// Plotter is something that can be drawn in a plot
type Plotter interface {
	dataRange() (xmin, xmax, ymin, ymax float64, ok bool)
	draw(c *canvas, m *mapping)
}

// mapping converts data to canvas coordinates
type mapping struct {
	area           image.Rectangle
	xmin, xmax     float64
	ymin, ymax     float64
	xticks, yticks []Tick
	scale          float64 // canvas pixels per point
}

func (m *mapping) x(v float64) float64 {
	return float64(m.area.Min.X) + (v-m.xmin)/(m.xmax-m.xmin)*float64(m.area.Dx())
}

func (m *mapping) y(v float64) float64 {
	return float64(m.area.Max.Y) - (v-m.ymin)/(m.ymax-m.ymin)*float64(m.area.Dy())
}

// Plot is a drawing with a title, two axes, a legend and plotters
type Plot struct {
	Title  struct{ Text string }
	X, Y   Axis
	Legend Legend

	plotters []Plotter
}

// New returns an empty plot
func New() *Plot {
	p := &Plot{}
	p.X.Min, p.X.Max = math.Inf(1), math.Inf(-1)
	p.Y.Min, p.Y.Max = math.Inf(1), math.Inf(-1)
	return p
}

// Add adds plotters, drawn in the order they are added
func (p *Plot) Add(ps ...Plotter) { p.plotters = append(p.plotters, ps...) }

// resolve gives the range of an axis: its Min and Max when set, otherwise the data with some margin
func resolve(a Axis, lo, hi float64, haveData bool) (float64, float64) {
	if a.Min < a.Max && !math.IsInf(a.Min, 0) && !math.IsInf(a.Max, 0) {
		return a.Min, a.Max
	}
	if !haveData {
		return 0, 1
	}
	if lo == hi {
		lo, hi = lo-1, hi+1
	}
	pad := (hi - lo) * 0.04
	return lo - pad, hi + pad
}

// Save draws the plot at a size in a PNG file
func (p *Plot) Save(w, h Length, path string) error {
	if path == "" {
		return errors.New("plot: Save needs a path")
	}
	if !strings.HasSuffix(strings.ToLower(path), ".png") {
		return fmt.Errorf("plot: %s: only PNG files are supported", path)
	}
	if w <= 0 || h <= 0 {
		return errors.New("plot: size must be positive")
	}
	img := p.render(int(float64(w)*dpi/72), int(float64(h)*dpi/72))
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

var (
	axisColor  = color.RGBA{60, 60, 60, 255}
	textColor  = color.RGBA{20, 20, 20, 255}
	gridColor  = color.RGBA{225, 225, 225, 255}
	boxColor   = color.NRGBA{255, 255, 255, 225}
	borderGray = color.RGBA{170, 170, 170, 255}
)

func (p *Plot) render(pw, ph int) *image.RGBA {
	if pw < 40 {
		pw = 40
	}
	if ph < 40 {
		ph = 40
	}
	W, H := pw*supersample, ph*supersample
	c := newCanvas(W, H)
	const tu, titleU, pad, tickLen = 3, 4, 12, 8

	// ranges and ticks
	xlo, xhi, ylo, yhi := math.Inf(1), math.Inf(-1), math.Inf(1), math.Inf(-1)
	have := false
	for _, pl := range p.plotters {
		if a, b, d, e, ok := pl.dataRange(); ok {
			xlo, xhi, ylo, yhi = math.Min(xlo, a), math.Max(xhi, b), math.Min(ylo, d), math.Max(yhi, e)
			have = true
		}
	}
	m := &mapping{scale: dpi / 72.0 * supersample}
	m.xmin, m.xmax = resolve(p.X, xlo, xhi, have)
	m.ymin, m.ymax = resolve(p.Y, ylo, yhi, have)
	xm, ym := p.X.Tick.Marker, p.Y.Tick.Marker
	if xm == nil {
		xm = niceTicks{target: max(3, pw/90)}
	}
	if ym == nil {
		ym = niceTicks{target: max(3, ph/55)}
	}
	m.xticks, m.yticks = xm.Ticks(m.xmin, m.xmax), ym.Ticks(m.ymin, m.ymax)

	// layout
	left := pad + tickLen + 6
	maxLabel := 0
	for _, t := range m.yticks {
		maxLabel = max(maxLabel, textWidth(t.Label, tu))
	}
	left += maxLabel
	if p.Y.Label.Text != "" {
		left += textHeight(tu) + pad
	}
	bottom := pad + tickLen + 6 + textHeight(tu)
	if p.X.Label.Text != "" {
		bottom += textHeight(tu) + pad
	}
	top := pad
	if p.Title.Text != "" {
		top += textHeight(titleU) + pad
	}
	right := pad
	if n := len(m.xticks); n > 0 {
		right += textWidth(m.xticks[n-1].Label, tu) / 2
	}
	m.area = image.Rect(left, top, W-right, H-bottom)
	if m.area.Dx() < 10 || m.area.Dy() < 10 {
		return c.downsample()
	}

	// grid-like plotters and data, clipped to the area
	full := c.clip
	c.clip = m.area
	for _, pl := range p.plotters {
		pl.draw(c, m)
	}
	c.clip = full

	// axes
	lw := float64(supersample)
	c.segment(pt{float64(m.area.Min.X), float64(m.area.Max.Y)}, pt{float64(m.area.Max.X), float64(m.area.Max.Y)}, lw, axisColor)
	c.segment(pt{float64(m.area.Min.X), float64(m.area.Min.Y)}, pt{float64(m.area.Min.X), float64(m.area.Max.Y)}, lw, axisColor)
	for _, t := range m.xticks {
		x := m.x(t.Value)
		if x < float64(m.area.Min.X)-1 || x > float64(m.area.Max.X)+1 {
			continue
		}
		l := float64(tickLen)
		if t.Label == "" {
			l /= 2
		}
		c.segment(pt{x, float64(m.area.Max.Y)}, pt{x, float64(m.area.Max.Y) + l}, lw, axisColor)
		if t.Label != "" {
			c.text(t.Label, int(x)-textWidth(t.Label, tu)/2, m.area.Max.Y+tickLen+6, tu, textColor, false)
		}
	}
	for _, t := range m.yticks {
		y := m.y(t.Value)
		if y < float64(m.area.Min.Y)-1 || y > float64(m.area.Max.Y)+1 {
			continue
		}
		l := float64(tickLen)
		if t.Label == "" {
			l /= 2
		}
		c.segment(pt{float64(m.area.Min.X), y}, pt{float64(m.area.Min.X) - l, y}, lw, axisColor)
		if t.Label != "" {
			c.text(t.Label, m.area.Min.X-tickLen-6-textWidth(t.Label, tu), int(y)-textHeight(tu)/2, tu, textColor, false)
		}
	}
	if s := p.X.Label.Text; s != "" {
		c.text(s, m.area.Min.X+(m.area.Dx()-textWidth(s, tu))/2, H-pad-textHeight(tu), tu, textColor, false)
	}
	if s := p.Y.Label.Text; s != "" {
		c.text(s, pad, m.area.Min.Y+(m.area.Dy()+textWidth(s, tu))/2, tu, textColor, true)
	}
	if s := p.Title.Text; s != "" {
		c.text(s, (W-textWidth(s, titleU))/2, pad, titleU, textColor, false)
	}
	p.drawLegend(c, m, tu)
	return c.downsample()
}

func (p *Plot) drawLegend(c *canvas, m *mapping, u int) {
	var entries []legendEntry
	for _, e := range p.Legend.entries {
		if e.name != "" {
			entries = append(entries, e)
		}
	}
	if len(entries) == 0 {
		return
	}
	const swatch, gap, pad = 30, 8, 8
	w := 0
	for _, e := range entries {
		w = max(w, textWidth(e.name, u))
	}
	rowH := textHeight(u) + 6
	boxW, boxH := pad*2+swatch+gap+w, pad*2+rowH*len(entries)-6
	x0 := m.area.Max.X - boxW - 10
	if p.Legend.Left {
		x0 = m.area.Min.X + 10
	}
	y0 := m.area.Max.Y - boxH - 10
	if p.Legend.Top {
		y0 = m.area.Min.Y + 10
	}
	box := image.Rect(x0, y0, x0+boxW, y0+boxH)
	saved := c.clip
	c.clip = saved
	c.fillRect(box, boxColor)
	for _, r := range []image.Rectangle{
		{box.Min, image.Pt(box.Max.X, box.Min.Y+1)}, {image.Pt(box.Min.X, box.Max.Y-1), box.Max},
		{box.Min, image.Pt(box.Min.X+1, box.Max.Y)}, {image.Pt(box.Max.X-1, box.Min.Y), box.Max},
	} {
		c.fillRect(r, borderGray)
	}
	for i, e := range entries {
		y := y0 + pad + i*rowH
		e.thumb.thumbnail(c, image.Rect(x0+pad, y, x0+pad+swatch, y+textHeight(u)))
		c.text(e.name, x0+pad+swatch+gap, y, u, textColor, false)
	}
}

// niceTicks puts about target ticks at round numbers
type niceTicks struct{ target int }

func (n niceTicks) Ticks(min, max float64) []Tick {
	if !(max > min) {
		return nil
	}
	raw := (max - min) / float64(n.target)
	mag := math.Pow(10, math.Floor(math.Log10(raw)))
	f := raw / mag
	var step float64
	switch {
	case f < 1.5:
		step = mag
	case f < 3.5:
		step = 2 * mag
	case f < 7.5:
		step = 5 * mag
	default:
		step = 10 * mag
	}
	var ticks []Tick
	for v := math.Ceil(min/step-1e-9) * step; v <= max+step*1e-9; v += step {
		v = math.Round(v/step) * step
		ticks = append(ticks, Tick{Value: v, Label: formatTick(v, step)})
	}
	return ticks
}

func formatTick(v, step float64) string {
	if math.Abs(v) < step*1e-9 {
		return "0"
	}
	if a := math.Abs(v); a >= 1e6 || a < 1e-4 {
		return strconv.FormatFloat(v, 'g', 4, 64)
	}
	dec := 0
	if step < 1 {
		dec = int(math.Ceil(-math.Log10(step) - 1e-9))
	}
	return strconv.FormatFloat(v, 'f', dec, 64)
}

// Palette is the list of line colours
var Palette = []color.RGBA{
	{31, 119, 180, 255}, {214, 39, 40, 255}, {44, 160, 44, 255}, {255, 127, 14, 255},
	{148, 103, 189, 255}, {23, 190, 207, 255}, {140, 86, 75, 255}, {227, 119, 194, 255},
}

// Color returns the i-th colour of the palette (it wraps around)
func Color(i int) color.Color { return Palette[((i%len(Palette))+len(Palette))%len(Palette)] }

// FramesX renders n frames of a generator as lines (one per dimension) with a title, axis labels and
// a legend (when lineNames has names) in a PNG
func FramesX(g float.FrameGenerator, n int, title, xLabel, yLabel string, lineNames []string, w, h Length, file string) error {
	p := New()
	p.Title.Text, p.X.Label.Text, p.Y.Label.Text = title, xLabel, yLabel
	p.Legend.Top = true
	dims := g.Dimensions()
	streams := make([]XYs, dims)
	for i := range streams {
		streams[i] = make(XYs, n)
	}
	for i := 0; i < n; i++ {
		for d, v := range g.Generate() {
			if d < dims {
				streams[d][i] = XY{float64(i), v}
			}
		}
	}
	for d, s := range streams {
		l, err := NewLine(s)
		if err != nil {
			return err
		}
		l.Color = Color(d)
		p.Add(l)
		if d < len(lineNames) {
			p.Legend.Add(lineNames[d], l)
		}
	}
	return p.Save(w, h, file)
}

// Frames renders n frames of a generator as lines in a PNG
func Frames(g float.FrameGenerator, n int, w, h Length, file string) error {
	return FramesX(g, n, "", "", "", nil, w, h, file)
}

// AddLines adds lines to the plot, in the palette colours and with legend entries. The arguments
// are pairs of a name (string) and points (XYer), or points alone
func AddLines(p *Plot, vs ...any) error {
	name, n := "", 0
	for _, v := range vs {
		switch v := v.(type) {
		case string:
			name = v
		case XYer:
			l, err := NewLine(v)
			if err != nil {
				return err
			}
			l.Color = Color(n)
			n++
			p.Add(l)
			if name != "" {
				p.Legend.Add(name, l)
			}
			name = ""
		default:
			return fmt.Errorf("plot: AddLines: unexpected %T", v)
		}
	}
	return nil
}
