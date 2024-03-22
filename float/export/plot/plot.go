package plot

import (
	"github.com/almerlucke/genny/float"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

func getPointStreams(g float.FrameGenerator, n int) []plotter.XYs {
	var (
		numDimensions = g.Dimensions()
		streams       = make([]plotter.XYs, numDimensions)
	)

	for i := 0; i < numDimensions; i++ {
		streams[i] = make(plotter.XYs, n)
	}

	for i := 0; i < n; i++ {
		for index, v := range g.Generate() {
			streams[index][i] = plotter.XY{
				X: float64(i),
				Y: v,
			}
		}
	}

	return streams
}

func PlotX(g float.FrameGenerator, n int, title string, xLabel string, yLabel string, lineNames []string, w vg.Length, h vg.Length, file string) error {
	var (
		p             = plot.New()
		numDimensions = g.Dimensions()
	)

	p.Title.Text = title
	p.X.Label.Text = xLabel
	p.Y.Label.Text = yLabel

	vs := make([]any, numDimensions*2)
	streams := getPointStreams(g, n)

	for i := 0; i < numDimensions; i++ {
		vs[i*2] = lineNames[i]
		vs[i*2+1] = streams[i]
	}

	err := plotutil.AddLines(p, vs...)
	if err != nil {
		return err
	}

	return p.Save(w, h, file)
}

func Plot(g float.FrameGenerator, n int, w vg.Length, h vg.Length, file string) error {
	return PlotX(g, n, "", "", "", make([]string, g.Dimensions()), w, h, file)
}
