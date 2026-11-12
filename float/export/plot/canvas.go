package plot

import (
	"image"
	"image/color"
	"math"
	"sort"
)

// supersample is the factor the drawing is rendered larger by and averaged down, for smooth edges
const supersample = 2

// canvas is an RGBA image with a clip rectangle and a few drawing primitives
type canvas struct {
	img  *image.RGBA
	clip image.Rectangle
}

func newCanvas(w, h int) *canvas {
	c := &canvas{img: image.NewRGBA(image.Rect(0, 0, w, h))}
	c.clip = c.img.Bounds()
	c.fillRect(c.img.Bounds(), color.White)
	return c
}

// blend paints one pixel (inside the clip) with the colour, honouring its alpha
func (c *canvas) blend(x, y int, col color.Color) {
	if !(image.Point{x, y}).In(c.clip) {
		return
	}
	r, g, b, a := col.RGBA()
	if a == 0 {
		return
	}
	if a == 0xffff {
		c.img.SetRGBA(x, y, color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), 255})
		return
	}
	d := c.img.RGBAAt(x, y)
	inv := 0xffff - a
	c.img.SetRGBA(x, y, color.RGBA{
		uint8((r + uint32(d.R)*0x101*inv/0xffff) >> 8),
		uint8((g + uint32(d.G)*0x101*inv/0xffff) >> 8),
		uint8((b + uint32(d.B)*0x101*inv/0xffff) >> 8),
		255,
	})
}

func (c *canvas) fillRect(r image.Rectangle, col color.Color) {
	r = r.Intersect(c.clip)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			c.blend(x, y, col)
		}
	}
}

type pt struct{ x, y float64 }

// fillPoly fills a polygon (even-odd rule)
func (c *canvas) fillPoly(pts []pt, col color.Color) {
	if len(pts) < 3 {
		return
	}
	ymin, ymax := math.Inf(1), math.Inf(-1)
	for _, p := range pts {
		ymin, ymax = math.Min(ymin, p.y), math.Max(ymax, p.y)
	}
	y0 := max(int(math.Floor(ymin)), c.clip.Min.Y)
	y1 := min(int(math.Ceil(ymax)), c.clip.Max.Y-1)
	var xs []float64
	for y := y0; y <= y1; y++ {
		sy := float64(y) + 0.5
		xs = xs[:0]
		for i := range pts {
			a, b := pts[i], pts[(i+1)%len(pts)]
			if (a.y <= sy) != (b.y <= sy) {
				xs = append(xs, a.x+(sy-a.y)/(b.y-a.y)*(b.x-a.x))
			}
		}
		sort.Float64s(xs)
		for i := 0; i+1 < len(xs); i += 2 {
			xa := max(int(math.Ceil(xs[i]-0.5)), c.clip.Min.X)
			xb := min(int(math.Ceil(xs[i+1]-0.5)), c.clip.Max.X)
			for x := xa; x < xb; x++ {
				c.blend(x, y, col)
			}
		}
	}
}

func (c *canvas) disc(cx, cy, r float64, col color.Color) {
	for y := int(math.Floor(cy - r)); y <= int(math.Ceil(cy+r)); y++ {
		for x := int(math.Floor(cx - r)); x <= int(math.Ceil(cx+r)); x++ {
			dx, dy := float64(x)+0.5-cx, float64(y)+0.5-cy
			if dx*dx+dy*dy <= r*r {
				c.blend(x, y, col)
			}
		}
	}
}

// segment draws a line segment of a width
func (c *canvas) segment(a, b pt, width float64, col color.Color) {
	dx, dy := b.x-a.x, b.y-a.y
	l := math.Hypot(dx, dy)
	if l == 0 {
		c.disc(a.x, a.y, width/2, col)
		return
	}
	nx, ny := -dy/l*width/2, dx/l*width/2
	c.fillPoly([]pt{{a.x + nx, a.y + ny}, {b.x + nx, b.y + ny}, {b.x - nx, b.y - ny}, {a.x - nx, a.y - ny}}, col)
}

// polyline draws connected segments with round joins
func (c *canvas) polyline(pts []pt, width float64, col color.Color) {
	for i := 1; i < len(pts); i++ {
		c.segment(pts[i-1], pts[i], width, col)
		if width > 2 && i > 1 {
			c.disc(pts[i-1].x, pts[i-1].y, width/2, col)
		}
	}
}

// textWidth is the width of a string drawn with a unit of u pixels per font pixel
func textWidth(s string, u int) int { return len([]rune(s)) * 6 * u }

// textHeight is the height of a line of text
func textHeight(u int) int { return 8 * u }

// text draws a string with its top-left corner at (x, y); rotated, it reads from bottom to top and
// (x, y) is the bottom-left corner
func (c *canvas) text(s string, x, y, u int, col color.Color, rotated bool) {
	for n, r := range []rune(s) {
		if r < 32 || r > 126 {
			r = '?'
		}
		glyph := font5x7[r-32]
		for i, column := range glyph {
			for j := 0; j < 8; j++ {
				if column>>j&1 == 0 {
					continue
				}
				gx, gy := (n*6+i)*u, j*u
				var rect image.Rectangle
				if rotated {
					rect = image.Rect(x+gy, y-gx-u, x+gy+u, y-gx)
				} else {
					rect = image.Rect(x+gx, y+gy, x+gx+u, y+gy+u)
				}
				c.fillRect(rect, col)
			}
		}
	}
}

// downsample averages blocks of supersample x supersample pixels
func (c *canvas) downsample() *image.RGBA {
	b := c.img.Bounds()
	out := image.NewRGBA(image.Rect(0, 0, b.Dx()/supersample, b.Dy()/supersample))
	n := uint32(supersample * supersample)
	for y := 0; y < out.Bounds().Dy(); y++ {
		for x := 0; x < out.Bounds().Dx(); x++ {
			var r, g, bl uint32
			for j := 0; j < supersample; j++ {
				for i := 0; i < supersample; i++ {
					p := c.img.RGBAAt(x*supersample+i, y*supersample+j)
					r, g, bl = r+uint32(p.R), g+uint32(p.G), bl+uint32(p.B)
				}
			}
			out.SetRGBA(x, y, color.RGBA{uint8(r / n), uint8(g / n), uint8(bl / n), 255})
		}
	}
	return out
}
