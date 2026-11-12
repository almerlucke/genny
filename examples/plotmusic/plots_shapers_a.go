package main

import (
	"fmt"
	"github.com/almerlucke/genny/float/export/plot"
	"image/color"
	"log"
	"math"

	"github.com/almerlucke/genny/float/shape"
	"github.com/almerlucke/genny/float/shape/shapers/curve"
	"github.com/almerlucke/genny/float/shape/shapers/deadzone"
	"github.com/almerlucke/genny/float/shape/shapers/harmonics"
	"github.com/almerlucke/genny/float/shape/shapers/rectify"
	"github.com/almerlucke/genny/float/shape/shapers/saturate"
	"github.com/almerlucke/genny/float/shape/shapers/smoothstep"
	"github.com/almerlucke/genny/float/shape/shapers/wavefold"
)

// shaperANamed is a shaper with the name used in a legend
type shaperANamed struct {
	name string
	s    shape.Shaper
}

// shaperAColor gives the i-th line color
func shaperAColor(i int) color.Color {
	palette := []color.RGBA{{31, 119, 180, 255}, {214, 39, 40, 255}, {44, 160, 44, 255}, {255, 127, 14, 255}, {148, 103, 189, 255}, {23, 190, 207, 255}}
	return palette[i%len(palette)]
}

// shaperALines draws one line per series with a legend
func shaperALines(dir, name, title, xLabel, yLabel string, series []shaperANamed, xs []float64, f func(shape.Shaper, float64) float64) {
	p := plot.New()
	p.Title.Text, p.X.Label.Text, p.Y.Label.Text = title, xLabel, yLabel
	p.Legend.Top = true
	p.Legend.Left = true
	for i, ns := range series {
		pts := make(plot.XYs, len(xs))
		for j, x := range xs {
			pts[j] = plot.XY{X: x, Y: f(ns.s, x)}
		}
		l, err := plot.NewLine(pts)
		if err != nil {
			log.Fatalf("%s: %v", name, err)
		}
		l.Color = shaperAColor(i)
		l.Width = plot.Points(1.5)
		p.Add(l)
		p.Legend.Add(ns.name, l)
	}
	p.Add(plot.NewGrid())
	if err := p.Save(14*plot.Centimeter, 10*plot.Centimeter, path(dir, name)); err != nil {
		log.Fatalf("%s: %v", name, err)
	}
}

// shaperARange returns n+1 evenly spaced values from lo to hi
func shaperARange(lo, hi float64, n int) []float64 {
	xs := make([]float64, n+1)
	for i := range xs {
		xs[i] = lo + (hi-lo)*float64(i)/float64(n)
	}
	return xs
}

// shaperATransfer plots the transfer curves y(x) for x in -1..1
func shaperATransfer(dir, name, title string, series ...shaperANamed) {
	shaperALines(dir, name, title, "input", "output", series, shaperARange(-1, 1, 800),
		func(s shape.Shaper, x float64) float64 { return s.Shape(x) })
}

// shaperATime plots one and a half periods of a sine of the given amplitude through each shaper
func shaperATime(dir, name, title string, amp float64, series ...shaperANamed) {
	in := shaperANamed{"input", nil}
	all := append([]shaperANamed{in}, series...)
	shaperALines(dir, name, title, "phase (periods)", "output", all, shaperARange(0, 2, 800),
		func(s shape.Shaper, t float64) float64 {
			x := amp * math.Sin(2*math.Pi*t)
			if s == nil {
				return x
			}
			return s.Shape(x)
		})
}

// shaperALabel formats a parameter for a legend
func shaperALabel(format string, v float64) string { return fmt.Sprintf(format, v) }

// plotsShapersA writes the transfer curves and sine responses of the wavefold, saturate, rectify, curve,
// smoothstep, deadzone and harmonics shapers to dir
func plotsShapersA(dir string) {
	var sine, tri []shaperANamed
	for _, g := range []float64{1, 2, 3, 4.5} {
		sine = append(sine, shaperANamed{shaperALabel("sine gain %g", g), wavefold.NewSine(g)})
		tri = append(tri, shaperANamed{shaperALabel("triangle gain %g", g), wavefold.NewTriangle(g)})
	}
	shaperATransfer(dir, "shaper_wavefold_sine.png", "Sine wave folder", sine...)
	shaperATransfer(dir, "shaper_wavefold_triangle.png", "Triangle wave folder", tri...)
	shaperATime(dir, "shaper_wavefold_time.png", "Wave folders on a sine (amplitude 1)", 1,
		shaperANamed{"sine gain 3", wavefold.NewSine(3)}, shaperANamed{"triangle gain 3", wavefold.NewTriangle(3)})

	sat := []shaperANamed{
		{"tanh drive 3", saturate.New(saturate.Tanh, 3)},
		{"atan drive 3", saturate.New(saturate.Atan, 3)},
		{"logistic drive 6", saturate.New(saturate.Logistic, 6)},
		{"tube drive 3 bias 0.4", saturate.NewTube(3, 0.4)},
		{"tube drive 3 bias -0.4", saturate.NewTube(3, -0.4)},
	}
	shaperATransfer(dir, "shaper_saturate.png", "Saturators", sat...)
	shaperATime(dir, "shaper_saturate_time.png", "Saturators on a sine (amplitude 1)", 1, sat[0], sat[2], sat[3])

	rec := []shaperANamed{
		{"half wave", rectify.New(rectify.HalfWave)},
		{"full wave", rectify.New(rectify.FullWave)},
		{"diode 0.2 hard", rectify.NewDiode(0.2, 0)},
		{"diode 0.2 knee 0.3", rectify.NewDiode(0.2, 0.3)},
	}
	shaperATransfer(dir, "shaper_rectify.png", "Rectifiers", rec...)
	shaperATime(dir, "shaper_rectify_time.png", "Rectifiers on a sine", 1, rec[0], rec[1], rec[3])

	var uni, bi []shaperANamed
	for _, t := range []float64{-0.9, -0.5, 0, 0.5, 0.9} {
		uni = append(uni, shaperANamed{shaperALabel("tension %g", t), curve.NewUnipolar(t)})
		bi = append(bi, shaperANamed{shaperALabel("tension %g", t), curve.NewBipolar(t)})
	}
	shaperALines(dir, "shaper_curve_unipolar.png", "Unipolar curves (like adsr shape)", "input", "output", uni,
		shaperARange(0, 1, 400), func(s shape.Shaper, x float64) float64 { return s.Shape(x) })
	shaperATransfer(dir, "shaper_curve_bipolar.png", "Bipolar curves", bi...)

	ss := []shaperANamed{
		{"linear", curve.NewUnipolar(0)},
		{"smoothstep", smoothstep.New(smoothstep.Smoothstep, 0, 1)},
		{"smootherstep", smoothstep.New(smoothstep.Smootherstep, 0, 1)},
		{"inverse smoothstep", smoothstep.New(smoothstep.Inverse, 0, 1)},
	}
	shaperALines(dir, "shaper_smoothstep.png", "Smoothstep family", "input", "output", ss,
		shaperARange(-0.2, 1.2, 600), func(s shape.Shaper, x float64) float64 { return s.Shape(x) })

	dz := []shaperANamed{
		{"dead zone 0.3 jump", deadzone.NewDeadZone(0.3, deadzone.Jump)},
		{"dead zone 0.3 continuous", deadzone.NewDeadZone(0.3, deadzone.Continuous)},
		{"expander 0.5 ratio 2", deadzone.NewExpander(0.5, 2)},
		{"expander 0.5 ratio 4", deadzone.NewExpander(0.5, 4)},
	}
	shaperATransfer(dir, "shaper_deadzone.png", "Dead zones and expanders", dz...)
	shaperATime(dir, "shaper_deadzone_time.png", "Dead zones and expander on a sine", 1, dz[0], dz[1], dz[2])

	hs := []shaperANamed{
		{"1", harmonics.New([]float64{1}, nil, true)},
		{"1 0 1/3 0 1/5 (square-ish)", harmonics.New([]float64{1, 0, 1.0 / 3, 0, 1.0 / 5}, nil, true)},
		{"1 1/2 1/3 1/4 (saw-ish)", harmonics.New([]float64{1, 0.5, 1.0 / 3, 0.25}, nil, true)},
		{"1 and 0.5 at phase pi/2", harmonics.New([]float64{1, 0.5}, []float64{0, math.Pi / 2}, true)},
	}
	shaperATransfer(dir, "shaper_harmonics.png", "Additive harmonic shapers", hs...)
	shaperATime(dir, "shaper_harmonics_time.png", "Harmonic shapers on a sine (amplitude 1)", 1, hs[1], hs[2], hs[3])
}
