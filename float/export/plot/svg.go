package plot

import (
	"encoding/xml"
	"fmt"
	"image"
	"image/color"
	"strconv"
	"strings"
)

// svg is a drawer that writes vector SVG. It takes the same coordinates as the raster canvas (the
// supersampled pixels become the user units of the viewBox), so both formats have the same layout.
type svg struct {
	b      strings.Builder
	w, h   int // size in pixels of one PNG pixel (the viewBox is supersample times larger)
	full   image.Rectangle
	clip   image.Rectangle
	inClip bool
	clips  int
}

// svgFont is the font of the text; the text is stretched to the width of the bitmap font so that
// labels take the same room as in a PNG
const svgFont = "Helvetica, Arial, sans-serif"

func newSVG(pw, ph int) *svg {
	s := &svg{w: pw, h: ph, full: image.Rect(0, 0, pw*supersample, ph*supersample)}
	s.clip = s.full
	fmt.Fprintf(&s.b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`+"\n",
		pw, ph, s.full.Dx(), s.full.Dy())
	s.fillRect(s.full, color.White)
	return s
}

// finish closes the document and returns it
func (s *svg) finish() string {
	s.closeClip()
	s.b.WriteString("</svg>\n")
	return s.b.String()
}

func (s *svg) closeClip() {
	if s.inClip {
		s.b.WriteString("</g>\n")
		s.inClip = false
	}
}

func (s *svg) getClip() image.Rectangle { return s.clip }

func (s *svg) setClip(r image.Rectangle) {
	if r == s.clip {
		return
	}
	s.closeClip()
	s.clip = r
	if r == s.full {
		return
	}
	s.clips++
	fmt.Fprintf(&s.b, `<clipPath id="clip%d"><rect x="%d" y="%d" width="%d" height="%d"/></clipPath>`+"\n<g clip-path=\"url(#clip%d)\">\n",
		s.clips, r.Min.X, r.Min.Y, r.Dx(), r.Dy(), s.clips)
	s.inClip = true
}

// num formats a coordinate with at most two decimals
func num(v float64) string {
	return strings.TrimSuffix(strings.TrimRight(strconv.FormatFloat(v, 'f', 2, 64), "0"), ".")
}

// paint returns the fill (or stroke) attributes of a colour; prefix is "fill" or "stroke"
func paint(prefix string, col color.Color) string {
	c := color.NRGBAModel.Convert(col).(color.NRGBA)
	out := fmt.Sprintf(`%s="#%02x%02x%02x"`, prefix, c.R, c.G, c.B)
	if c.A != 255 {
		out += fmt.Sprintf(` %s-opacity="%s"`, prefix, strconv.FormatFloat(float64(c.A)/255, 'f', 3, 64))
	}
	return out
}

func points(pts []pt) string {
	var sb strings.Builder
	for i, p := range pts {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(num(p.x) + "," + num(p.y))
	}
	return sb.String()
}

func (s *svg) fillRect(r image.Rectangle, col color.Color) {
	fmt.Fprintf(&s.b, `<rect x="%d" y="%d" width="%d" height="%d" %s/>`+"\n", r.Min.X, r.Min.Y, r.Dx(), r.Dy(), paint("fill", col))
}

func (s *svg) fillPoly(pts []pt, col color.Color) {
	if len(pts) < 3 {
		return
	}
	fmt.Fprintf(&s.b, `<polygon points="%s" %s fill-rule="evenodd"/>`+"\n", points(pts), paint("fill", col))
}

func (s *svg) disc(cx, cy, r float64, col color.Color) {
	fmt.Fprintf(&s.b, `<circle cx="%s" cy="%s" r="%s" %s/>`+"\n", num(cx), num(cy), num(r), paint("fill", col))
}

func (s *svg) segment(a, b pt, width float64, col color.Color) {
	if a == b {
		s.disc(a.x, a.y, width/2, col)
		return
	}
	fmt.Fprintf(&s.b, `<line x1="%s" y1="%s" x2="%s" y2="%s" %s stroke-width="%s"/>`+"\n",
		num(a.x), num(a.y), num(b.x), num(b.y), paint("stroke", col), num(width))
}

func (s *svg) polyline(pts []pt, width float64, col color.Color) {
	switch len(pts) {
	case 0:
		return
	case 1:
		s.disc(pts[0].x, pts[0].y, width/2, col)
		return
	}
	fmt.Fprintf(&s.b, `<polyline points="%s" fill="none" %s stroke-width="%s" stroke-linejoin="round"/>`+"\n",
		points(pts), paint("stroke", col), num(width))
}

func (s *svg) text(str string, x, y, u int, col color.Color, rotated bool) {
	if str == "" {
		return
	}
	var esc strings.Builder
	_ = xml.EscapeText(&esc, []byte(str))
	attrs := fmt.Sprintf(`font-family="%s" font-size="%d" textLength="%d" lengthAdjust="spacingAndGlyphs" %s`,
		svgFont, 10*u, textWidth(str, u), paint("fill", col))
	if rotated {
		// reads from bottom to top, (x, y) is the bottom-left corner of the bitmap text
		fmt.Fprintf(&s.b, `<text transform="translate(%d,%d) rotate(-90)" %s>%s</text>`+"\n", x+7*u, y, attrs, esc.String())
		return
	}
	fmt.Fprintf(&s.b, `<text x="%d" y="%d" %s>%s</text>`+"\n", x, y+7*u, attrs, esc.String())
}
