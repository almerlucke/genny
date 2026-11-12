package main

import (
	"fmt"
	"github.com/almerlucke/genny/float/export/plot"
	"log"
	"math"

	"github.com/almerlucke/genny/float/shape"
	"github.com/almerlucke/genny/float/shape/shapers/bitcrush"
	"github.com/almerlucke/genny/float/shape/shapers/funnel"
	"github.com/almerlucke/genny/float/shape/shapers/hysteresis"
	"github.com/almerlucke/genny/float/shape/shapers/onepole"
	"github.com/almerlucke/genny/float/shape/shapers/phasedist"
	"github.com/almerlucke/genny/float/shape/shapers/polynomial"
	"github.com/almerlucke/genny/float/shape/shapers/samplehold"
	"github.com/almerlucke/genny/float/shape/shapers/window"
	"github.com/almerlucke/genny/scale"
	"github.com/almerlucke/genny/sequence"
	"github.com/almerlucke/genny/sieve"
)

// shaperBCurve is a named curve to plot
type shaperBCurve struct {
	name string
	pts  plot.XYs
}

// shaperBSample evaluates f at n+1 evenly spaced points of [x0,x1]
func shaperBSample(name string, x0, x1 float64, n int, f func(x float64) float64) shaperBCurve {
	pts := make(plot.XYs, n+1)
	for i := range pts {
		x := x0 + (x1-x0)*float64(i)/float64(n)
		pts[i] = plot.XY{X: x, Y: f(x)}
	}
	return shaperBCurve{name, pts}
}

// shaperBSave draws the curves as lines to a PNG file
func shaperBSave(dir, name, title, xLabel, yLabel string, curves ...shaperBCurve) {
	p := plot.New()
	p.Title.Text = title
	p.X.Label.Text = xLabel
	p.Y.Label.Text = yLabel
	p.Legend.Top = true
	var vs []any
	for _, c := range curves {
		vs = append(vs, c.name, c.pts)
	}
	if err := plot.AddLines(p, vs...); err != nil {
		log.Fatalf("%s: %v", name, err)
	}
	if err := p.Save(20*plot.Centimeter, 9*plot.Centimeter, path(dir, name)); err != nil {
		log.Fatalf("%s: %v", name, err)
	}
}

// shaperBRun feeds the inputs through a shaper and returns the output as a curve over the call index
func shaperBRun(name string, s shape.Shaper, in func(i int) float64, n int) shaperBCurve {
	pts := make(plot.XYs, n)
	for i := range pts {
		pts[i] = plot.XY{X: float64(i), Y: s.Shape(in(i))}
	}
	return shaperBCurve{name, pts}
}

// shaperBTriangle is a triangle between -amp and amp with the given period in calls
func shaperBTriangle(amp float64, period int) func(i int) float64 {
	return func(i int) float64 {
		ph := float64(i%period) / float64(period)
		return amp * (1 - 4*math.Abs(ph-0.5))
	}
}

// plotsShapersB draws the phase distortion waveforms, windows, bit crusher, hysteresis, one pole filter,
// sample and hold, Legendre polynomials and funnels
func plotsShapersB(dir string) {
	// phase distortion: one cycle for several amounts d
	for _, v := range []struct {
		name string
		make func(d float64) shape.Shaper
	}{
		{"saw", func(d float64) shape.Shaper { return phasedist.NewSaw(d) }},
		{"square", func(d float64) shape.Shaper { return phasedist.NewSquare(d) }},
		{"pulse", func(d float64) shape.Shaper { return phasedist.NewPulse(d) }},
		{"resonant (harmonic 6)", func(d float64) shape.Shaper { return phasedist.NewResonant(d, 6) }},
	} {
		var curves []shaperBCurve
		for _, d := range []float64{0, 0.5, 0.8, 1} {
			curves = append(curves, shaperBSample(fmt.Sprintf("d=%.1f", d), 0, 1, 400, v.make(d).Shape))
		}
		shaperBSave(dir, "phasedist_"+v.name[:3]+".png", "Phase distortion: "+v.name, "phase", "", curves...)
	}

	// window family
	var wins []shaperBCurve
	for _, w := range []struct {
		name string
		w    *window.Window
	}{
		{"hann", window.Hann()}, {"hamming", window.Hamming()}, {"blackman", window.Blackman()},
		{"tukey 0.5", window.Tukey(0.5)}, {"gaussian 0.3", window.Gaussian(0.3)},
		{"kaiser 8", window.Kaiser(8)}, {"welch", window.Welch()}, {"triangle", window.Triangle()},
	} {
		wins = append(wins, shaperBSample(w.name, 0, 1, 400, w.w.Shape))
	}
	shaperBSave(dir, "windows.png", "Window functions", "position", "", wins...)

	// bit crusher: a sine at several depths, and with sample rate reduction
	sine := func(i int) float64 { return 0.95 * math.Sin(2*math.Pi*float64(i)/200) }
	crush := []shaperBCurve{shaperBRun("sine", bitcrush.New(24, 1, 1), sine, 200)}
	for _, bits := range []int{2, 3, 4} {
		crush = append(crush, shaperBRun(fmt.Sprintf("%d bits", bits), bitcrush.New(bits, 1, 1), sine, 200))
	}
	crush = append(crush, shaperBRun("4 bits, hold 10", bitcrush.New(4, 10, 1), sine, 200))
	shaperBSave(dir, "bitcrush.png", "Bit crusher", "call", "", crush...)

	// hysteresis: a triangle in, the loops as x/y and over time
	tri := shaperBTriangle(1, 100)
	schmitt := hysteresis.NewSchmitt(-0.3, 0.3, -1, 1)
	play := hysteresis.NewPlay(0.8)
	loopS, loopP := make(plot.XYs, 200), make(plot.XYs, 200)
	for i := range loopS {
		x := tri(i)
		loopS[i] = plot.XY{X: x, Y: schmitt.Shape(x)}
		loopP[i] = plot.XY{X: x, Y: play.Shape(x)}
	}
	shaperBSave(dir, "hysteresis_loop.png", "Hysteresis loops (input on x)", "in", "out",
		shaperBCurve{"schmitt -0.3..0.3", loopS}, shaperBCurve{"play, width 0.8", loopP})
	schmitt.Reset()
	play.Reset()
	shaperBSave(dir, "hysteresis_time.png", "Hysteresis over time", "call", "",
		shaperBRun("input", shaperBIdentity{}, tri, 200),
		shaperBRun("schmitt", schmitt, tri, 200),
		shaperBRun("play", play, tri, 200))

	// one pole step response, lowpass and highpass
	step := func(i int) float64 {
		if i < 10 {
			return 0
		}
		return 1
	}
	shaperBSave(dir, "onepole.png", "One pole step response", "call", "",
		shaperBRun("step", shaperBIdentity{}, step, 120),
		shaperBRun("lowpass tau 10", onepole.NewSamples(10), step, 120),
		shaperBRun("lowpass tau 30", onepole.NewSamples(30), step, 120),
		shaperBRun("highpass tau 20", onepole.NewHighpassTime(0.02, 1000), step, 120))

	// sample and hold, track and hold of a sine
	slow := func(i int) float64 { return math.Sin(2 * math.Pi * float64(i) / 100) }
	sh := samplehold.New(sequence.NewLoop(true, false, false, false, false, false, false, false, false, false))
	th := samplehold.NewTrack(sequence.NewLoop(true, true, true, true, true, false, false, false, false, false))
	shaperBSave(dir, "samplehold.png", "Sample and hold, track and hold", "call", "",
		shaperBRun("sine", shaperBIdentity{}, slow, 200),
		shaperBRun("sample & hold (every 10)", sh, slow, 200),
		shaperBRun("track & hold (5 on, 5 off)", th, slow, 200))

	// Legendre polynomials
	var leg []shaperBCurve
	for n := 1; n <= 5; n++ {
		leg = append(leg, shaperBSample(fmt.Sprintf("P%d", n), -1, 1, 400, polynomial.NewLegendre(n).Shape))
	}
	shaperBSave(dir, "legendre.png", "Legendre polynomials", "x", "", leg...)

	// funnel: a ramp snapped to A minor and to a sieve
	ramp := func(i int) float64 { return 55 + 24*float64(i)/400 }
	fs := funnel.FromScale(scale.New(57, scale.Minor))
	fv := funnel.FromSieve(sieve.Of(0, 24, 0, 3, 7, 8, 12, 15, 19, 20), 55)
	shaperBSave(dir, "funnel.png", "Funnel to a scale and a sieve", "call", "pitch",
		shaperBRun("ramp", shaperBIdentity{}, ramp, 400),
		shaperBRun("A minor", fs, ramp, 400),
		shaperBRun("sieve", fv, ramp, 400))
}

// shaperBIdentity passes its input, to draw the input next to the output
type shaperBIdentity struct{}

func (shaperBIdentity) Shape(x float64) float64 { return x }
