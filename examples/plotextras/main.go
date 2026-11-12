// plotextras shows the histogram, bar and spectrum plotters and the SVG output of the plot package:
// go run ./examples/plotextras [dir] writes every plot as a PNG and as an SVG in dir (default "plots")
package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/float/dist"
	"github.com/almerlucke/genny/float/export/plot"
	"github.com/almerlucke/genny/float/shape/shapers/wavefold"
)

var dir = "plots"

// save writes the plot in both formats
func save(p *plot.Plot, name string) {
	for _, ext := range []string{".png", ".svg"} {
		if err := p.Save(18*plot.Centimeter, 9*plot.Centimeter, filepath.Join(dir, name+ext)); err != nil {
			log.Fatalf("%s%s: %v", name, ext, err)
		}
	}
	fmt.Println("wrote", name+".png and .svg")
}

func must[T any](v T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return v
}

// averaged is the power spectrum in dB averaged over blocks of a size (Welch's method), which
// smooths the random ripple of the spectrum of noise
func averaged(data []float64, block int, rate float64) plot.XYs {
	var freqs, power []float64
	blocks := len(data) / block
	for b := 0; b < blocks; b++ {
		f, m := plot.MagnitudeSpectrum(data[b*block:(b+1)*block], rate, true)
		if power == nil {
			freqs, power = f, make([]float64, len(f))
		}
		for k := range m {
			power[k] += m[k] * m[k] / float64(blocks)
		}
	}
	var pts plot.XYs
	for k := 1; k < len(freqs); k++ {
		pts = append(pts, plot.XY{X: freqs[k], Y: 10 * math.Log10(power[k])})
	}
	return pts
}

func main() {
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Fatal(err)
	}
	genny.Seed(1)

	// histogram of beta distributed draws, as a density
	beta := dist.NewBeta(2, 5)
	draws := make([]float64, 5000)
	for i := range draws {
		draws[i] = beta.Generate()
	}
	h := must(plot.NewHistogram(draws, 25))
	h.Density = true
	h.Min, h.Max = 0, 1
	p := plot.New()
	p.Title.Text, p.X.Label.Text, p.Y.Label.Text = "5000 draws of Beta(2, 5)", "value", "density"
	p.Add(plot.NewGrid(), h)
	p.Legend.Add("draws", h)
	save(p, "histogram")

	// bars: how often each pitch class occurs in a melody
	melody := []int{62, 64, 65, 67, 69, 67, 65, 64, 62, 62, 65, 69, 72, 71, 69, 67, 65, 67, 69, 65, 62, 64, 62, 57, 62, 64, 65, 64, 62}
	names := []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}
	counts := make([]float64, 12)
	for _, pitch := range melody {
		counts[pitch%12]++
	}
	bars := must(plot.NewBars(names, counts))
	p = plot.New()
	p.Title.Text, p.X.Label.Text, p.Y.Label.Text = "pitch classes of a melody", "pitch class", "notes"
	p.Add(plot.NewGrid(), bars)
	save(p, "pitchclasses")

	// spectra of a sine through the sine wavefolder at three fold gains, in dB on a log axis
	const rate, n = 44100.0, 8192
	p = plot.New()
	p.Title.Text, p.X.Label.Text, p.Y.Label.Text = "wavefolded 441 Hz sine", "frequency (Hz)", "dB"
	p.X.Scale = plot.Log
	p.X.Min, p.X.Max = 100, 20000
	p.Y.Min, p.Y.Max = -100, 5
	p.Legend.Left = true
	p.Add(plot.NewGrid())
	for i, gain := range []float64{1, 2, 4} {
		// 441 Hz at 44100 Hz is exactly 8192/100 periods away from a bin; 2^13 samples at bin 82 is close
		// enough with the Hann window
		l := must(plot.ShaperSpectrum(wavefold.NewSine(gain), 441, 0.9, n, rate,
			plot.SpectrumOptions{Hann: true, DB: true, LogFreq: true, FloorDB: -100}))
		l.Color = plot.Color(i)
		p.Add(l)
		p.Legend.Add(fmt.Sprintf("gain %g", gain), l)
	}
	save(p, "wavefold_spectrum")

	// white against pink noise
	genny.Seed(7)
	pink := dist.NewPink(12)
	white, pinks := make([]float64, 16384), make([]float64, 16384)
	for i := range white {
		white[i] = genny.Float64() - 0.5
		pinks[i] = pink.Generate() - 0.5
	}
	p = plot.New()
	p.Title.Text, p.X.Label.Text, p.Y.Label.Text = "white and pink noise", "frequency (Hz)", "dB"
	p.X.Scale = plot.Log
	p.Legend.Left = true
	p.Add(plot.NewGrid())
	for i, s := range []struct {
		name string
		data []float64
	}{{"white", white}, {"pink", pinks}} {
		l := must(plot.NewLine(averaged(s.data, 1024, rate)))
		l.Color = plot.Color(i)
		p.Add(l)
		p.Legend.Add(s.name, l)
	}
	save(p, "noise_spectrum")
}
