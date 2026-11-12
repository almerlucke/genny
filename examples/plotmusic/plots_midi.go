package main

import (
	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/ngram/style"
	"github.com/almerlucke/genny/note"
)

// odeNotes is the tune Ode to Joy (public domain) as events: pitch and duration in sixteenths
func odeNotes() []note.Event {
	pitches := []float64{
		64, 64, 65, 67, 67, 65, 64, 62, 60, 60, 62, 64, 64, 62, 62,
		64, 64, 65, 67, 67, 65, 64, 62, 60, 60, 62, 64, 62, 60, 60,
		62, 62, 64, 60, 62, 64, 65, 64, 60, 62, 64, 65, 64, 62, 60, 62, 55,
		64, 64, 65, 67, 67, 65, 64, 62, 60, 60, 62, 64, 64, 62, 60,
	}
	durs := []float64{
		4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 6, 2, 8,
		4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 6, 2, 8,
		4, 4, 4, 4, 4, 2, 2, 4, 4, 4, 2, 2, 4, 4, 4, 4, 8,
		4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 6, 2, 8,
	}
	var out []note.Event
	t := 0.0
	for i, p := range pitches {
		out = append(out, note.Event{Pitch: p, Velocity: 0.8, Start: t, Duration: durs[i]})
		t += durs[i]
	}
	return out
}

// msHistogram returns a function that walks over the bins lo..hi of the relative frequency of the values,
// one bin per call. Values outside the range are left out
func msHistogram(values []float64, lo, hi int) func() float64 {
	h := make([]float64, hi-lo+1)
	total := 0.0
	for _, v := range values {
		if i := int(v+0.5) - lo; i >= 0 && i < len(h) {
			h[i]++
			total++
		}
	}
	i := -1
	return func() float64 {
		i = (i + 1) % len(h)
		return h[i] / max(total, 1)
	}
}

// plotsMidi writes the plots of style learning to dir: how often each interval and pitch occurs in a tune and in a
// long melody made in its style (an n-gram of order 2)
func plotsMidi(dir string) {
	genny.Seed(5)
	tune := odeNotes()
	original := style.Melody(tune)
	learned := style.Learn(tune, 2, style.Intervals(), style.Ratios()).Phrase(4000)

	intervals := func(notes []style.Note) (out []float64) {
		for i := 1; i < len(notes); i++ {
			out = append(out, notes[i].Pitch-notes[i-1].Pitch)
		}
		return
	}
	pitches := func(notes []style.Note) (out []float64) {
		for _, n := range notes {
			out = append(out, n.Pitch)
		}
		return
	}

	// the intervals (bin 0 is -12 semitones, bin 12 is a repeated note): the style makes the same moves as often
	lines(dir, "style_intervals.png", "Intervals of a tune and of its learned style (bin 12 = unison)", 25,
		[]string{"original", "learned"},
		msHistogram(intervals(original), -12, 12), msHistogram(intervals(learned), -12, 12))

	// the pitches (bin 0 is MIDI note 53): the learned melody stays in the range of the tune
	lines(dir, "style_pitches.png", "Pitches of a tune and of its learned style (bin 0 = note 53)", 28,
		[]string{"original", "learned"},
		msHistogram(pitches(original), 53, 80), msHistogram(pitches(learned), 53, 80))
}
