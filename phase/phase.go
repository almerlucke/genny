// Package phase makes phasing music in the style of Steve Reich (Piano Phase, Clapping Music): two
// voices play the same pattern, one of them a little faster, so that it slowly drifts away from the other
// and comes back into unison
package phase

import (
	"math"

	"github.com/almerlucke/genny/note"
)

// Length is the length of one cycle of the pattern: the end of the note that ends last (0 for an empty
// pattern)
func Length(pattern []note.Event) float64 {
	n := 0.0
	for _, e := range pattern {
		n = max(n, e.Start+e.Duration)
	}
	return n
}

// Cycles is the number of cycles after which a voice that gains drift steps on every cycle has gained a
// whole cycle and is in unison again. It is 0 for an empty pattern
func Cycles(pattern []note.Event, drift float64) int {
	l := Length(pattern)
	if l == 0 || !(drift > 0) {
		return 0
	}
	return int(math.Ceil(l/drift - 1e-9))
}

// Events plays the pattern for the given number of cycles with two voices. The first voice repeats it
// every cycle (the length of the pattern). The second repeats it every cycle minus drift, so it gets
// ahead by drift each cycle; after Cycles(pattern, drift) cycles it has gained a whole cycle and
// plays in unison with the first again. Notes of the second voice that start after the end of the first
// voice (cycles times the length of the pattern) are left out. It panics if cycles < 1 or if drift is not
// in (0, length of the pattern); an empty pattern gives two empty lines
func Events(pattern []note.Event, cycles int, drift float64) (first, second []note.Event) {
	if cycles < 1 {
		panic("phase: at least one cycle is needed")
	}
	l := Length(pattern)
	if l == 0 {
		return nil, nil
	}
	if !(drift > 0) || drift >= l {
		panic("phase: drift must be positive and shorter than the pattern")
	}

	total := float64(cycles) * l
	for k := range cycles {
		for _, e := range pattern {
			e.Start += float64(k) * l
			first = append(first, e)
		}
	}
	for k := 0; ; k++ {
		offset := float64(k) * (l - drift)
		if offset >= total {
			break
		}
		for _, e := range pattern {
			e.Start += offset
			if e.Start < total {
				second = append(second, e)
			}
		}
	}
	return first, second
}
