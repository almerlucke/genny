package plot

import (
	"github.com/almerlucke/genny"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

func getPointStreams(g genny.Generator[[]float64], dim int, n int) []plotter.XYs {
	streams := make([]plotter.XYs, dim)
	for i := 0; i < dim; i++ {
		streams[i] = make(plotter.XYs, n)
	}

	for i := 0; i < n; i++ {
		vec := g.Generate()
		for index, v := range vec {
			streams[index][i] = plotter.XY{
				X: float64(i),
				Y: v,
			}
		}
	}

	return streams
}

func Plotf(g genny.Generator[[]float64], dim int, n int, title string, xLabel string, yLabel string, lineNames []string, w vg.Length, h vg.Length, file string) error {
	p := plot.New()

	p.Title.Text = title
	p.X.Label.Text = xLabel
	p.Y.Label.Text = yLabel

	vs := make([]any, dim*2)
	streams := getPointStreams(g, dim, n)

	for i := 0; i < dim; i++ {
		vs[i*2] = lineNames[i]
		vs[i*2+1] = streams[i]
	}

	err := plotutil.AddLines(p, vs...)
	if err != nil {
		return err
	}

	return p.Save(w, h, file)
}

func Plot(g genny.Generator[[]float64], dim int, n int, w vg.Length, h vg.Length, file string) error {
	return Plotf(g, dim, n, "", "", "", make([]string, dim), w, h, file)
}
