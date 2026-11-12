package plot_test

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/almerlucke/genny/float/export/plot"
)

func TestSaveDrawsEveryPlotter(t *testing.T) {
	p := plot.New()
	p.Title.Text, p.X.Label.Text, p.Y.Label.Text = "all of them", "x", "y"
	p.Legend.Top, p.Legend.Left = true, true
	p.Add(plot.NewGrid())
	line, err := plot.NewLine(plot.XYs{{0, 0}, {1, 1}, {2, 0.5}})
	if err != nil {
		t.Fatal(err)
	}
	dots, _ := plot.NewScatter(plot.XYs{{0.5, 0.5}, {1.5, 0.2}})
	dots.Shape = plot.BoxGlyph{}
	poly, err := plot.NewPolygon(plot.XYs{{0, 0}, {1, 0}, {1, 1}})
	if err != nil {
		t.Fatal(err)
	}
	p.Add(line, dots, poly)
	p.Legend.Add("line", line)
	p.Legend.Add("dots", dots)
	p.Legend.Add("poly", poly)

	file := filepath.Join(t.TempDir(), "all.png")
	if err := p.Save(10*plot.Centimeter, 6*plot.Centimeter, file); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(file)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	if b := img.Bounds(); b.Dx() < 400 || b.Dy() < 200 {
		t.Errorf("image is too small: %v", b)
	}
}

func TestErrors(t *testing.T) {
	if _, err := plot.NewLine(plot.XYs{}); err == nil {
		t.Error("a line without points should fail")
	}
	if _, err := plot.NewLine(plot.XYs{{0, 1.0 / zero()}}); err == nil {
		t.Error("an infinite point should fail")
	}
	if _, err := plot.NewPolygon(plot.XYs{{0, 0}, {1, 1}}); err == nil {
		t.Error("a polygon needs three points")
	}
	p := plot.New()
	if err := p.Save(plot.Inch, plot.Inch, filepath.Join(t.TempDir(), "x.svg")); err == nil {
		t.Error("only png is supported")
	}
	// an empty plot still renders
	if err := p.Save(plot.Inch, plot.Inch, filepath.Join(t.TempDir(), "x.png")); err != nil {
		t.Error(err)
	}
}

func zero() float64 { return 0 }
