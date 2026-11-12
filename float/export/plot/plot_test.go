package plot_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/almerlucke/genny/float"
	"github.com/almerlucke/genny/float/export/plot"
	"github.com/almerlucke/genny/float/phasor"
)

func TestFramesWritesAPNG(t *testing.T) {
	file := filepath.Join(t.TempDir(), "p.png")
	g := float.ToFrame(phasor.New(1, 16, 0))
	if err := plot.Frames(g, 32, 200, 100, file); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 8 || string(data[1:4]) != "PNG" {
		t.Fatalf("not a png: %d bytes", len(data))
	}

	file2 := filepath.Join(t.TempDir(), "p2.png")
	if err := plot.FramesX(g, 16, "title", "x", "y", []string{"phase"}, 2*plot.Inch, plot.Inch, file2); err != nil {
		t.Fatal(err)
	}
	if st, err := os.Stat(file2); err != nil || st.Size() == 0 {
		t.Fatal("PlotX wrote nothing")
	}
}
