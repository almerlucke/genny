package plot

import (
	"errors"
	"fmt"
	"math"

	"github.com/almerlucke/genny/note"
)

// Track is a named line of notes for Roll
type Track struct {
	Name   string
	Events []note.Event
}

var noteNames = []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}

// noteTicks puts a label on every C and a minor tick on the other pitches
type noteTicks struct{}

func (noteTicks) Ticks(min, max float64) []Tick {
	var ticks []Tick
	for p := int(math.Ceil(min)); p <= int(math.Floor(max)); p++ {
		pc := ((p % 12) + 12) % 12
		if pc == 0 {
			ticks = append(ticks, Tick{Value: float64(p), Label: fmt.Sprintf("C%d", p/12-1)})
		} else if max-min < 40 {
			ticks = append(ticks, Tick{Value: float64(p)})
		}
	}
	return ticks
}

// Roll draws tracks of notes as a piano roll in a PNG (or any other format of the extension of path):
// time in steps horizontal, pitch vertical, a bar for every note, every track in its own colour with a
// legend. Tracks without notes are drawn as an empty
// entry in the legend, and no tracks at all give an empty roll. It returns an error if the file cannot
// be written
func Roll(tracks []Track, width, height Length, path string) error {
	if path == "" {
		return errors.New("plot: Roll needs a path")
	}
	p := New()
	p.X.Label.Text = "step"
	p.Y.Label.Text = "pitch"
	p.Y.Tick.Marker = noteTicks{}
	p.Legend.Top = true

	xmax := 1.0
	ymin, ymax := math.Inf(1), math.Inf(-1)
	for i, tr := range tracks {
		var polys []XYer
		for _, e := range tr.Events {
			if e.Duration <= 0 {
				continue
			}
			polys = append(polys, XYs{
				{X: e.Start, Y: e.Pitch - 0.4}, {X: e.Start + e.Duration, Y: e.Pitch - 0.4},
				{X: e.Start + e.Duration, Y: e.Pitch + 0.4}, {X: e.Start, Y: e.Pitch + 0.4},
			})
			xmax = max(xmax, e.Start+e.Duration)
			ymin, ymax = min(ymin, e.Pitch), max(ymax, e.Pitch)
		}
		col := Color(i)
		if len(polys) == 0 {
			// still a legend entry
			dummy, err := NewLine(XYs{{X: 0, Y: 0}})
			if err != nil {
				return err
			}
			dummy.Color = col
			p.Legend.Add(tr.Name, dummy)
			continue
		}
		poly, err := NewPolygon(polys...)
		if err != nil {
			return err
		}
		poly.Color = col
		poly.LineStyle.Color = col
		poly.LineStyle.Width = Points(0.3)
		p.Add(poly)
		p.Legend.Add(tr.Name, poly)
	}
	if math.IsInf(ymin, 1) {
		ymin, ymax = 59, 72
	}
	p.X.Min, p.X.Max = 0, xmax
	p.Y.Min, p.Y.Max = ymin-2, ymax+5
	return p.Save(width, height, path)
}
