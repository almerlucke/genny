package plot_test

import (
	"encoding/xml"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/almerlucke/genny/float/export/plot"
)

func TestHistogramCounts(t *testing.T) {
	vals := []float64{0, 0.1, 0.2, 0.5, 0.9, 1, 1, math.NaN()}
	h, err := plot.NewHistogram(vals, 4)
	if err != nil {
		t.Fatal(err)
	}
	edges, counts := h.Counts()
	if len(edges) != 5 || edges[0] != 0 || edges[4] != 1 {
		t.Errorf("edges %v", edges)
	}
	sum := 0
	for _, c := range counts {
		sum += c
	}
	if sum != 7 {
		t.Errorf("counts %v add up to %d, want 7 (the NaN is left out)", counts, sum)
	}
	if counts[3] != 3 { // 0.9 and both ones: the upper edge counts in the last bin
		t.Errorf("counts %v", counts)
	}
	h.Min, h.Max = 0, 0.5
	_, counts = h.Counts()
	sum = 0
	for _, c := range counts {
		sum += c
	}
	if sum != 4 { // 0, 0.1, 0.2, 0.5
		t.Errorf("fixed range counts %v", counts)
	}
	same, _ := plot.NewHistogram([]float64{3, 3, 3}, 5)
	if _, c := same.Counts(); c[2] != 3 {
		t.Errorf("equal values should land in the middle bin: %v", c)
	}
	for _, bad := range []struct {
		v []float64
		n int
	}{{vals, 0}, {nil, 3}, {[]float64{math.NaN()}, 3}} {
		if _, err := plot.NewHistogram(bad.v, bad.n); err == nil {
			t.Errorf("%v with %d bins should fail", bad.v, bad.n)
		}
	}
}

func TestBarsErrors(t *testing.T) {
	if _, err := plot.NewBars([]string{"a"}, []float64{1, 2}); err == nil {
		t.Error("mismatched labels should fail")
	}
	if _, err := plot.NewBars(nil, nil); err == nil {
		t.Error("no bars should fail")
	}
	b, err := plot.NewBars(nil, []float64{1, 2})
	if err != nil || b.Ticks(0, 1)[1].Label != "1" {
		t.Error("unlabelled bars use their index")
	}
}

// svgStats decodes an SVG and counts its elements and text
func svgStats(t *testing.T, file string) (count map[string]int, text string) {
	t.Helper()
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	dec := xml.NewDecoder(strings.NewReader(string(data)))
	count = map[string]int{}
	var sb strings.Builder
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("not well-formed XML: %v", err)
		}
		switch tk := tok.(type) {
		case xml.StartElement:
			count[tk.Name.Local]++
		case xml.CharData:
			sb.Write(tk)
		}
	}
	return count, sb.String()
}

func TestSaveSVG(t *testing.T) {
	p := plot.New()
	p.Title.Text, p.X.Label.Text, p.Y.Label.Text = `a <b> & "c"`, "x", "y"
	p.Legend.Top = true
	p.Add(plot.NewGrid())
	line, _ := plot.NewLine(plot.XYs{{0, 0}, {1, 1}, {2, 0.5}})
	dots, _ := plot.NewScatter(plot.XYs{{0.5, 0.5}, {1.5, 0.2}, {1, 1}})
	poly, _ := plot.NewPolygon(plot.XYs{{0, 0}, {1, 0}, {1, 1}})
	p.Add(line, dots, poly)
	p.Legend.Add("line", line)
	file := filepath.Join(t.TempDir(), "p.svg")
	if err := p.Save(10*plot.Centimeter, 6*plot.Centimeter, file); err != nil {
		t.Fatal(err)
	}
	count, text := svgStats(t, file)
	if !strings.Contains(text, `a <b> & "c"`) {
		t.Errorf("title missing (escaping must round trip): %q", text)
	}
	if count["svg"] != 1 || count["clipPath"] != 1 {
		t.Errorf("elements: %v", count)
	}
	if count["circle"] != 3 {
		t.Errorf("want one circle per point, got %d", count["circle"])
	}
	// the data polygon, its outline, the polyline of the line, the legend swatch and box
	if count["polygon"] != 1 || count["polyline"] != 2 {
		t.Errorf("want 1 polygon and 2 polylines (line and polygon outline): %v", count)
	}
	if count["text"] < 5 { // title, labels, ticks, legend
		t.Errorf("text elements: %d", count["text"])
	}
}

func TestSVGCountsMatchData(t *testing.T) {
	h, _ := plot.NewHistogram([]float64{1, 2, 2, 3, 3, 3, 4, 4, 4, 4}, 4)
	p := plot.New()
	p.Add(h)
	file := filepath.Join(t.TempDir(), "h.svg")
	if err := p.Save(8*plot.Centimeter, 5*plot.Centimeter, file); err != nil {
		t.Fatal(err)
	}
	if count, _ := svgStats(t, file); count["polygon"] != 4 {
		t.Errorf("one polygon per bar: %v", count)
	}
}

func TestExtrasRenderInBothFormats(t *testing.T) {
	samples := make([]float64, 256)
	for i := range samples {
		samples[i] = math.Sin(float64(i) * 0.3)
	}
	spec, err := plot.NewSpectrum(samples, 44100, plot.SpectrumOptions{DB: true, LogFreq: true, Hann: true})
	if err != nil {
		t.Fatal(err)
	}
	bars, _ := plot.NewBars([]string{"a", "b", "c"}, []float64{3, -1, 2})
	for name, build := range map[string]func() *plot.Plot{
		"spectrum": func() *plot.Plot {
			p := plot.New()
			p.X.Scale = plot.Log
			p.Add(spec)
			return p
		},
		"bars": func() *plot.Plot { p := plot.New(); p.Add(bars); p.Legend.Add("bars", bars); return p },
	} {
		for _, ext := range []string{".png", ".svg"} {
			file := filepath.Join(t.TempDir(), name+ext)
			if err := build().Save(10*plot.Centimeter, 6*plot.Centimeter, file); err != nil {
				t.Fatalf("%s%s: %v", name, ext, err)
			}
			if ext == ".svg" {
				svgStats(t, file)
			}
		}
	}
	// the labels of a log axis are drawn
	p := plot.New()
	p.X.Scale = plot.Log
	p.Add(spec)
	file := filepath.Join(t.TempDir(), "log.svg")
	if err := p.Save(12*plot.Centimeter, 6*plot.Centimeter, file); err != nil {
		t.Fatal(err)
	}
	if _, text := svgStats(t, file); !strings.Contains(text, "1k") {
		t.Errorf("log axis should be labelled 1k: %q", text)
	}
}
