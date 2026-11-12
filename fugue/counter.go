package fugue

import (
	"errors"
	"math"
	"slices"

	"github.com/almerlucke/genny/counterpoint"
	"github.com/almerlucke/genny/note"
	"github.com/almerlucke/genny/scale"
)

// Errors of Countersubject
var (
	// ErrNoCountersubject: no countersubject could be written to the subject within the attempts
	ErrNoCountersubject = errors.New("fugue: no countersubject found")
	// ErrNotInvertible: a countersubject was found but it is not invertible at the octave and fitting the
	// answer; it is returned with this error and works above the subject only
	ErrNotInvertible = errors.New("fugue: the countersubject is not invertible at the octave")
	// ErrWholeSemitones: the key has degrees that are not whole semitones
	ErrWholeSemitones = errors.New("fugue: a countersubject needs a key of whole semitones")
)

func pitches(s Subject, key *scale.Scale, shift int) ([]int, bool) {
	out := make([]int, len(s.Degrees))
	for i, d := range s.Degrees {
		p := key.Degree(d + shift)
		if math.Abs(p-math.Round(p)) > 1e-9 {
			return nil, false
		}
		out[i] = int(math.Round(p))
	}
	return out, true
}

// Countersubject writes a countersubject to the subject with package counterpoint: a line above the subject
// that keeps to the strict rules and is invertible at the octave (no fifths, within an octave of the
// subject), so it works as well below it when moved down an octave. It is second species (two notes
// against each note, the last note one) or first species, see Config.Species. The line is tried against the
// answer too (moved up a fifth together with it), above and below; if every combination passes the rules
// of counterpoint the result is invertible, otherwise the first line found is returned with
// ErrNotInvertible, and when no line is found at all ErrNoCountersubject. The countersubject has the
// length of the subject and uses the random stream, so seed first
func Countersubject(subject Subject, key *scale.Scale, cfg Config) (Subject, error) {
	if len(subject.Degrees) == 0 {
		return Subject{}, ErrNoCountersubject
	}
	cantus, ok := pitches(subject, key, 0)
	if !ok {
		return Subject{}, ErrWholeSemitones
	}
	species := 2
	if cfg.Species == 1 {
		species = 1
	}
	attempts := cfg.Attempts
	if attempts <= 0 {
		attempts = 400
	}

	// the durations: every note of the subject is split in two, except the last one
	var durations []float64
	for i, d := range subject.Durations {
		if species == 2 && i < len(subject.Durations)-1 {
			durations = append(durations, d/2, d/2)
		} else {
			durations = append(durations, d)
		}
	}

	var fallback Subject
	found := false
	for range attempts {
		gen := counterpoint.Generate
		if species == 2 {
			gen = counterpoint.GenerateSecond
		}
		cc := counterpoint.Config{Side: counterpoint.Above, Scale: key, Invertible: true}
		cp, err := gen(cantus, cc)
		if err != nil || len(cp) != len(durations) {
			continue
		}
		degrees := make([]int, len(cp))
		for i, p := range cp {
			degrees[i] = key.DegreeOf(float64(p))
		}
		cs := Subject{degrees, durations}
		if !found {
			fallback, found = cs, true
		}
		if bothSides(cs, 0, cantus, key, species) {
			if _, ok := adaptToAnswer(subject, cs, key, cfg.Answer, species); ok {
				return cs, nil
			}
		}
	}
	if found {
		return fallback, ErrNotInvertible
	}
	return Subject{}, ErrNoCountersubject
}

// CounterAnswer is the countersubject to play against the answer. With a real answer, or a tonal answer
// that is the subject a fifth up unchanged, it is the countersubject moved a fifth up. A tonal answer
// changes the intervals of its head, so the notes of the countersubject against that head are written
// again (a search of the notes near the old ones that keeps the strict rules, above and below the
// answer); the rest stays. It returns ErrNotInvertible, with the countersubject moved up a fifth, if no
// such notes exist
func CounterAnswer(subject, cs Subject, key *scale.Scale, kind AnswerKind) (Subject, error) {
	species := 2
	if len(cs.Durations) == len(subject.Durations) {
		species = 1
	}
	out, ok := adaptToAnswer(subject, cs, key, kind, species)
	if !ok {
		return out, ErrNotInvertible
	}
	return out, nil
}

func adaptToAnswer(subject, cs Subject, key *scale.Scale, kind AnswerKind, species int) (Subject, bool) {
	fifth := key.DegreeOf(key.Root() + 7)
	answer := Answer(subject, key, kind)
	plain := cs.Shifted(fifth)
	// the head: the notes of the answer that differ from the subject a fifth up
	head := 0
	for i := range answer.Degrees {
		if answer.Degrees[i] != subject.Degrees[i]+fifth {
			head = i + 1
		}
	}
	cf, ok := pitches(answer, key, 0)
	if !ok {
		return plain, false
	}
	if head == 0 {
		return plain, bothSides(plain, 0, cf, key, species)
	}
	until := 0.0
	for _, d := range subject.Durations[:head] {
		until += d
	}
	free := 0 // the notes of the countersubject that lie under the head
	t := 0.0
	for free < len(cs.Durations) && t < until-1e-9 {
		t += cs.Durations[free]
		free++
	}
	// the slots on the beat: where a note of the subject starts
	starts := map[float64]bool{}
	t = 0
	for _, d := range subject.Durations {
		starts[t] = true
		t += d
	}
	pitchAt := func(at float64) int {
		t := 0.0
		for i, d := range answer.Durations {
			if at < t+d-1e-9 {
				return cf[i]
			}
			t += d
		}
		return cf[len(cf)-1]
	}
	out := Subject{slices.Clone(plain.Degrees), plain.Durations}
	slotStart := make([]float64, len(out.Durations))
	t = 0
	for i, d := range out.Durations {
		slotStart[i] = t
		t += d
	}
	offsets := []int{0, 1, -1, 2, -2, 3, -3}
	budget := 200000
	var search func(j int) bool
	search = func(j int) bool {
		if j == free {
			budget--
			return bothSides(out, 0, cf, key, species)
		}
		for _, off := range offsets {
			if budget <= 0 {
				return false
			}
			d := plain.Degrees[j] + off
			p := key.Degree(d)
			iv := int(math.Round(p)) - pitchAt(slotStart[j])
			if starts[slotStart[j]] && (iv < 3 || iv > 12 || !counterpoint.Consonant(iv) || iv%12 == 7) {
				continue
			}
			if j > 0 && abs(d-out.Degrees[j-1]) > 3 {
				continue
			}
			out.Degrees[j] = d
			if search(j + 1) {
				return true
			}
		}
		out.Degrees[j] = plain.Degrees[j]
		return false
	}
	if search(0) {
		return out, true
	}
	return plain, false
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// bothSides tests the countersubject, moved by shift degrees, above the cantus and an octave lower below
// it, with the strict rules of counterpoint
func bothSides(cs Subject, shift int, cf []int, key *scale.Scale, species int) bool {
	for _, t := range []struct {
		side  counterpoint.Side
		shift int
	}{{counterpoint.Above, shift}, {counterpoint.Below, shift - key.Len()}} {
		cp, ok := pitches(cs, key, t.shift)
		if !ok {
			return false
		}
		cfg := counterpoint.Config{Side: t.side, Scale: key}
		var vs []counterpoint.Violation
		if species == 2 {
			vs = cfg.CheckSecond(cf, cp)
		} else {
			vs = cfg.Check(cf, cp)
		}
		if len(vs) != 0 {
			return false
		}
	}
	return true
}

// sounding returns the pitch of the line at time t
func sounding(line []note.Event, t float64) (float64, bool) {
	for _, e := range line {
		if e.Start <= t+1e-9 && t < e.Start+e.Duration-1e-9 {
			return e.Pitch, true
		}
	}
	return 0, false
}

// accents are the moments that count for the vertical check: the beats and the starts of notes of at
// least a beat
func accents(beat float64, lines ...[]note.Event) []float64 {
	end := 0.0
	var out []float64
	for _, l := range lines {
		for _, e := range l {
			end = max(end, e.Start+e.Duration)
			if e.Duration >= beat-1e-9 {
				out = append(out, e.Start)
			}
		}
	}
	for t := 0.0; t < end; t += beat {
		out = append(out, t)
	}
	slices.Sort(out)
	return slices.Compact(out)
}

// clashes counts the accents at which both lines sound and form a dissonance
func clashes(a, b []note.Event, beat float64) int {
	n := 0
	for _, t := range accents(beat, a, b) {
		pa, okA := sounding(a, t)
		pb, okB := sounding(b, t)
		if okA && okB && !counterpoint.Consonant(int(math.Round(pa-pb))) {
			n++
		}
	}
	return n
}

// clashesAt counts the points at which both lines sound and form a dissonance
func clashesAt(a, b []note.Event, points []float64) int {
	n := 0
	for _, t := range points {
		pa, okA := sounding(a, t)
		pb, okB := sounding(b, t)
		if okA && okB && !counterpoint.Consonant(int(math.Round(pa-pb))) {
			n++
		}
	}
	return n
}

// StrettoDistances lists the distances (in whole beats, shorter than the subject) at which the entries
// of a stretto with the given number of voices sound well together: the subject and the answer follow
// each other at that distance, and at every accent (the beats, and the starts of notes of a beat or more)
// where two entries sound together the interval is a consonance. This is the check Stretto applies; it
// does not judge parallel fifths or the shape of the result
func StrettoDistances(subject Subject, key *scale.Scale, kind AnswerKind, voices int, beat float64) []float64 {
	if beat <= 0 || voices < 2 {
		return nil
	}
	var out []float64
	for d := beat; d < subject.Length()-1e-9; d += beat {
		if strettoClashes(strettoEvents(subject, key, kind, voices, d, 0), beat) == 0 {
			out = append(out, d)
		}
	}
	return out
}

func strettoEvents(subject Subject, key *scale.Scale, kind AnswerKind, voices int, distance, start float64) [][]note.Event {
	answer := Answer(subject, key, kind)
	out := make([][]note.Event, voices)
	for i := range out {
		s := subject
		if i%2 == 1 {
			s = answer
		}
		out[i] = s.Events(key, start+float64(i)*distance, 0)
	}
	return out
}

func strettoClashes(entries [][]note.Event, beat float64) int {
	n := 0
	for i := range entries {
		for j := i + 1; j < len(entries); j++ {
			n += clashes(entries[i], entries[j], beat)
		}
	}
	return n
}
