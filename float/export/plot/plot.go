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
	"path/filepath"
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
	// Scale is Linear (the default) or Log. A log axis needs positive data and a positive Min
	Scale Scale
}

// Scale is the scale of an axis
type Scale int

const (
	Linear Scale = iota
	Log
)

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
	thumbnail(c drawer, r image.Rectangle)
}

// Add adds an entry to the legend
func (l *Legend) Add(name string, t Thumbnailer) { l.entries = append(l.entries, legendEntry{name, t}) }

// Plotter is something that can be drawn in a plot
type Plotter interface {
	dataRange() (xmin, xmax, ymin, ymax float64, ok bool)
	draw(c drawer, m *mapping)
}

// mapping converts data to canvas coordinates
type mapping struct {
	area           image.Rectangle
	xmin, xmax     float64
	ymin, ymax     float64
	xticks, yticks []Tick
	xlog, ylog     bool
	scale          float64 // canvas pixels per point
}

// frac is the position of v between lo and hi, from 0 to 1, on a linear or a log scale
func frac(v, lo, hi float64, log bool) float64 {
	if log {
		if v <= 0 {
			return math.Inf(-1)
		}
		return (math.Log10(v) - math.Log10(lo)) / (math.Log10(hi) - math.Log10(lo))
	}
	return (v - lo) / (hi - lo)
}

func (m *mapping) x(v float64) float64 {
	return float64(m.area.Min.X) + frac(v, m.xmin, m.xmax, m.xlog)*float64(m.area.Dx())
}

func (m *mapping) y(v float64) float64 {
	return float64(m.area.Max.Y) - frac(v, m.ymin, m.ymax, m.ylog)*float64(m.area.Dy())
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

// fixed reports whether both Min and Max of the axis are set
func fixed(a Axis) bool {
	return a.Min < a.Max && !math.IsInf(a.Min, 0) && !math.IsInf(a.Max, 0)
}

// resolve gives the range of an axis: its Min and Max when set, otherwise the data with some margin
func resolve(a Axis, lo, hi float64, haveData bool) (float64, float64) {
	if fixed(a) {
		return a.Min, a.Max
	}
	if a.Scale == Log {
		if !haveData || hi <= 0 {
			return 1, 10
		}
		if lo <= 0 {
			lo = hi / 1e3
		}
		if lo == hi {
			return lo / 2, hi * 2
		}
		pad := math.Pow(hi/lo, 0.02)
		return lo / pad, hi * pad
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

// Save draws the plot at a size in a file: a PNG or an SVG, chosen by the extension of the path
func (p *Plot) Save(w, h Length, path string) error {
	if path == "" {
		return errors.New("plot: Save needs a path")
	}
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".png" && ext != ".svg" {
		return fmt.Errorf("plot: %s: only PNG and SVG files are supported", path)
	}
	if w <= 0 || h <= 0 {
		return errors.New("plot: size must be positive")
	}
	pw, ph := pixelSize(int(float64(w)*dpi/72), int(float64(h)*dpi/72))
	if ext == ".svg" {
		s := newSVG(pw, ph)
		p.paint(s, pw, ph)
		return os.WriteFile(path, []byte(s.finish()), 0o644)
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := png.Encode(f, p.render(pw, ph)); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// pixelSize applies the minimum size of a plot
func pixelSize(pw, ph int) (int, int) { return max(pw, 40), max(ph, 40) }

var (
	axisColor  = color.RGBA{60, 60, 60, 255}
	textColor  = color.RGBA{20, 20, 20, 255}
	gridColor  = color.RGBA{225, 225, 225, 255}
	boxColor   = color.NRGBA{255, 255, 255, 225}
	borderGray = color.RGBA{170, 170, 170, 255}
)

func (p *Plot) render(pw, ph int) *image.RGBA {
	pw, ph = pixelSize(pw, ph)
	c := newCanvas(pw*supersample, ph*supersample)
	p.paint(c, pw, ph)
	return c.downsample()
}

// paint draws the plot on a drawer of pw x ph pixels (times supersample)
func (p *Plot) paint(c drawer, pw, ph int) {
	W, H := pw*supersample, ph*supersample
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
	m.xlog, m.ylog = p.X.Scale == Log, p.Y.Scale == Log
	xm, ym := p.X.Tick.Marker, p.Y.Tick.Marker
	for _, pl := range p.plotters {
		// bars start at the axis: no margin below zero
		if zb, ok := pl.(interface{ zeroBased() bool }); ok && zb.zeroBased() && !m.ylog && !fixed(p.Y) && ylo == 0 {
			m.ymin = 0
		}
		if tp, ok := pl.(TickMarker); ok && xm == nil {
			xm = tp
		}
	}
	if xm == nil && m.xlog {
		xm = LogTicks{}
	}
	if ym == nil && m.ylog {
		ym = LogTicks{}
	}
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
		return
	}

	// grid-like plotters and data, clipped to the area
	full := c.getClip()
	c.setClip(m.area)
	for _, pl := range p.plotters {
		pl.draw(c, m)
	}
	c.setClip(full)

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
}

func (p *Plot) drawLegend(c drawer, m *mapping, u int) {
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

// LogTicks puts labelled ticks on the powers of ten (1, 10, 100, 1k, 10k, 1M, 0.1, ...) and small
// unlabelled ticks on 2..9 times a power of ten. When less than two powers of ten fit the range, 2
// and 5 get a label too. Use it on an axis with Scale Log (that is the default there)
type LogTicks struct{}

func (LogTicks) Ticks(min, max float64) []Tick {
	if !(min > 0) || !(max > min) {
		return nil
	}
	lo, hi := int(math.Floor(math.Log10(min))), int(math.Ceil(math.Log10(max)))
	decades := 0
	for e := lo; e <= hi; e++ {
		if v := math.Pow10(e); v >= min*(1-1e-9) && v <= max*(1+1e-9) {
			decades++
		}
	}
	var ticks []Tick
	for e := lo; e <= hi; e++ {
		for k := 1; k <= 9; k++ {
			v := float64(k) * math.Pow10(e)
			if v < min*(1-1e-9) || v > max*(1+1e-9) {
				continue
			}
			t := Tick{Value: v}
			if k == 1 || (decades < 2 && (k == 2 || k == 5)) {
				t.Label = formatLog(v)
			}
			ticks = append(ticks, t)
		}
	}
	return ticks
}

// formatLog writes 1, 10, 100, 1k, 10k, 1M, 0.1, 0.01
func formatLog(v float64) string {
	switch a := math.Abs(v); {
	case a >= 1e6:
		return strconv.FormatFloat(v/1e6, 'g', 4, 64) + "M"
	case a >= 1e3:
		return strconv.FormatFloat(v/1e3, 'g', 4, 64) + "k"
	case a >= 1e-3:
		return strconv.FormatFloat(v, 'f', -1, 64)
	}
	return strconv.FormatFloat(v, 'g', 4, 64)
}
