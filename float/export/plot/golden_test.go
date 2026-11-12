package plot_test

import (
	"crypto/sha256"
	"encoding/hex"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/almerlucke/genny/float/export/plot"
	"github.com/almerlucke/genny/note"
)

// goldenScenes are plots whose PNG pixels were hashed before the drawer refactor
func goldenScenes(t *testing.T) map[string]*plot.Plot {
	p := plot.New()
	p.Title.Text, p.X.Label.Text, p.Y.Label.Text = "all of them", "x", "y"
	p.Legend.Top, p.Legend.Left = true, true
	p.Add(plot.NewGrid())
	line, _ := plot.NewLine(plot.XYs{{0, 0}, {1, 1}, {2, 0.5}})
	dots, _ := plot.NewScatter(plot.XYs{{0.5, 0.5}, {1.5, 0.2}})
	dots.Shape = plot.BoxGlyph{}
	poly, _ := plot.NewPolygon(plot.XYs{{0, 0}, {1, 0}, {1, 1}})
	p.Add(line, dots, poly)
	p.Legend.Add("line", line)
	p.Legend.Add("dots", dots)
	p.Legend.Add("poly", poly)

	q := plot.New()
	q.Title.Text = "second"
	l2, _ := plot.NewLine(plot.XYs{{0, 3}, {5, -2}, {9, 4}})
	l2.Color = plot.Color(1)
	l2.Width = plot.Points(3)
	q.Add(plot.NewGrid(), l2)
	q.Legend.Add("thick", l2)
	return map[string]*plot.Plot{"all": p, "second": q}
}

func pixelHash(t *testing.T, file string) string {
	f, err := os.Open(file)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	h := sha256.New()
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := img.At(x, y).RGBA()
			h.Write([]byte{byte(r >> 8), byte(g >> 8), byte(bl >> 8), byte(a >> 8)})
		}
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

var goldenHashes = map[string]string{
	"all": "810ae7f192e694b0", "second": "050cda1881db56c9", "roll": "c794dc4930ddedb9",
}

func TestPNGUnchangedByRefactor(t *testing.T) {
	if runtime.GOARCH != "amd64" {
		t.Skip("golden hashes were taken on amd64")
	}
	dir := t.TempDir()
	for name, p := range goldenScenes(t) {
		file := filepath.Join(dir, name+".png")
		if err := p.Save(10*plot.Centimeter, 6*plot.Centimeter, file); err != nil {
			t.Fatal(err)
		}
		got := pixelHash(t, file)
		if want, ok := goldenHashes[name]; !ok {
			t.Logf("GOLDEN %s %s", name, got)
		} else if got != want {
			t.Errorf("%s: pixel hash %s, want %s", name, got, want)
		}
	}
	roll := filepath.Join(dir, "roll.png")
	err := plot.Roll([]plot.Track{{Name: "a", Events: []note.Event{{Pitch: 60, Start: 0, Duration: 2}, {Pitch: 64, Start: 2, Duration: 2}}}}, 12*plot.Centimeter, 6*plot.Centimeter, roll)
	if err != nil {
		t.Fatal(err)
	}
	got := pixelHash(t, roll)
	if want, ok := goldenHashes["roll"]; !ok {
		t.Logf("GOLDEN roll %s", got)
	} else if got != want {
		t.Errorf("roll: pixel hash %s, want %s", got, want)
	}
}
