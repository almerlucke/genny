// Package fugue builds the sections of a fugue from a subject: an exposition in which the voices enter
// one after the other with the subject and the answer, a countersubject that is invertible at the octave,
// episodes made of sequences of fragments, middle entries in other keys, a stretto in which the entries
// overlap, and a closing on a pedal. The result is a Fugue with one []note.Event per voice, ready for
// package midi, and a checker (Check) that every entry is the subject or the answer, transposed correctly.
//
// Everything is written in scale degrees of a key (package scale), so a transposition is a diatonic
// one and the pitches stay in the scale. Voice 0 is the highest voice. Time is in steps, like the rest of
// the library; a bar is Config.StepsPerBar steps (default 16) and every section starts on a barline.
//
// The honest limitations of the method:
//   - The key is one fixed scale. A real fugue modulates, and answers at the dominant use a raised
//     leading tone of the new key; here the answer is the subject moved up four degrees inside the same
//     scale (in D minor the subject on D becomes the subject on A with the notes of D minor, so the C
//     stays natural). Only the closing can raise the leading tone (CloseConfig.LeadingTone). Moving a
//     subject by two degrees (a third) is a real modulation, to the relative major.
//   - A tonal answer (the tonic-dominant mapping) is made for the head of the subject only: the run of
//     notes at its start that are tonic or dominant. It is the textbook rule, not a judgement of taste.
//   - The countersubject and the free voices are written with the strict rules of package counterpoint
//     (second and first species), so they are correct in a school sense but plain. The voices in an
//     episode or middle entry that carry no material rest. The vertical check of the free voices is by
//     consonance on the beats, not a full voice leading analysis, and the closing does not check
//     parallels.
//   - A countersubject needs a key of whole semitones (every scale of package scale except the microtonal
//     ones).
package fugue

import (
	"errors"
	"fmt"
	"math"
	"slices"

	"github.com/almerlucke/genny/note"
	"github.com/almerlucke/genny/scale"
)

// Subject is a melody as scale degrees with a rhythm: Degrees[i] lasts Durations[i] steps. Degree 0 is the
// tonic of the key, 1 the next scale note, 7 the tonic an octave up (in a seven note scale), -1 the
// note below the tonic. The same type holds a countersubject
type Subject struct {
	Degrees   []int
	Durations []float64
}

// NewSubject makes a subject. It panics if there are no notes, if the lengths differ or a duration is not
// positive
func NewSubject(degrees []int, durations []float64) Subject {
	if len(degrees) == 0 || len(degrees) != len(durations) {
		panic("fugue: a subject needs as many durations as degrees, and at least one note")
	}
	for _, d := range durations {
		if !(d > 0) || math.IsInf(d, 0) {
			panic("fugue: durations must be positive")
		}
	}
	return Subject{slices.Clone(degrees), slices.Clone(durations)}
}

// SubjectFromPitches makes a subject from pitches (in semitones) that must all be notes of the key. It
// panics if a pitch is not in the key, and as NewSubject
func SubjectFromPitches(key *scale.Scale, pitches, durations []float64) Subject {
	degrees := make([]int, len(pitches))
	for i, p := range pitches {
		degrees[i] = key.DegreeOf(p)
		if math.Abs(key.Degree(degrees[i])-p) > 1e-9 {
			panic(fmt.Sprintf("fugue: pitch %g is not in the key", p))
		}
	}
	return NewSubject(degrees, durations)
}

// Length is the duration of the subject in steps
func (s Subject) Length() float64 {
	t := 0.0
	for _, d := range s.Durations {
		t += d
	}
	return t
}

// Shifted returns the subject moved by a number of scale degrees (a diatonic transposition)
func (s Subject) Shifted(degrees int) Subject {
	out := Subject{slices.Clone(s.Degrees), slices.Clone(s.Durations)}
	for i := range out.Degrees {
		out.Degrees[i] += degrees
	}
	return out
}

// Events turns the subject into notes in the key, starting at step start and moved by shift degrees
func (s Subject) Events(key *scale.Scale, start float64, shift int) []note.Event {
	out := make([]note.Event, len(s.Degrees))
	t := start
	for i, d := range s.Degrees {
		out[i] = note.Event{Pitch: key.Degree(d + shift), Velocity: 0.8, Start: t, Duration: s.Durations[i]}
		t += s.Durations[i]
	}
	return out
}

// AnswerKind tells how the answer is made from the subject
type AnswerKind int

const (
	// Tonal keeps the answer in the key: the tonic of the subject is answered by the dominant and the
	// dominant by the tonic, so a fifth up in the head of the subject is answered by a fourth up
	Tonal AnswerKind = iota
	// Real moves every note up a fifth (four degrees), the intervals stay the same
	Real
)

// Answer makes the answer to the subject: the subject at the dominant, a fifth up. A Real answer moves every
// degree up to the dominant. A Tonal answer does that too, except for the head: the notes at the start
// of the subject that are tonic or dominant. There the tonic goes up a fifth to the dominant and the
// dominant goes up a fourth to the tonic, so the answer stays in the key where a real answer would
// modulate. The head ends at the first note that is neither
func Answer(s Subject, key *scale.Scale, kind AnswerKind) Subject {
	n := key.Len()
	fifth := key.DegreeOf(key.Root() + 7)
	out := s.Shifted(fifth)
	if kind != Tonal {
		return out
	}
	for i, d := range s.Degrees {
		switch mod(d, n) {
		case 0:
			out.Degrees[i] = d + fifth
		case fifth:
			out.Degrees[i] = d + n - fifth
		default:
			return out
		}
	}
	return out
}

func mod(a, n int) int { return ((a % n) + n) % n }

// Kind tells what an entry states
type Kind int

const (
	// KindSubject is the subject itself
	KindSubject Kind = iota
	// KindAnswer is the answer
	KindAnswer
)

func (k Kind) String() string {
	if k == KindAnswer {
		return "answer"
	}
	return "subject"
}

// Entry is one statement of the subject or the answer in a voice
type Entry struct {
	Voice int
	// Start of the entry in steps
	Start float64
	Kind  Kind
	// Shift is the number of degrees the entry is moved from the subject (Kind subject) or from the answer
	// (Kind answer) as made by Answer. In an exposition and a stretto it is a whole number of octaves
	// (a multiple of the number of degrees of the key), in a middle entry it is a change of key area
	Shift   int
	Section string
}

// Section is a stretch of time with a name: "exposition", "episode", "middle entry", "stretto", "close"
type Section struct {
	Name       string
	Start, End float64
}

// Config is the setup of a fugue. The zero value gives a tonal answer, bars of 16 steps, a beat of 4 steps,
// voices that enter from the top voice down and a second species countersubject
type Config struct {
	// Answer is the kind of answer (default Tonal)
	Answer AnswerKind
	// StepsPerBar (default 16) and Beat (default half a bar) in steps. The beat is the grid of accents
	// where the entries of a stretto must be consonant
	StepsPerBar, Beat float64
	// Centers are the middle pitches of the voices, voice 0 first. By default 74 for voice 0 and 7
	// semitones lower for every next voice (12 with two voices)
	Centers []float64
	// Order is the voice of each entry of the exposition, in order of entry. By default the voices enter
	// from the top down, so each voice answers the one above it and the countersubject sits next to it. Other
	// orders are possible but the countersubject then has to jump over the voices in between
	Order []int
	// Species of the countersubject, 1 (note against note) or 2 (two notes against one, the default)
	Species int
	// Velocity of the notes (default 0.8)
	Velocity float64
	// Attempts is how many countersubjects are tried before giving up (default 400)
	Attempts int
}

func (c Config) bar() float64 {
	if c.StepsPerBar > 0 {
		return c.StepsPerBar
	}
	return 16
}

func (c Config) beat() float64 {
	if c.Beat > 0 {
		return c.Beat
	}
	return c.bar() / 2
}

func (c Config) velocity() float64 {
	if c.Velocity > 0 {
		return c.Velocity
	}
	return 0.8
}

func (c Config) centers(voices int) []float64 {
	if len(c.Centers) == voices {
		return c.Centers
	}
	out := make([]float64, voices)
	spacing := 7.0
	if voices == 2 {
		spacing = 12
	}
	for v := range out {
		out[v] = 74 - spacing*float64(v)
	}
	return out
}

func (c Config) order(voices int) []int {
	if len(c.Order) == voices {
		seen := make([]bool, voices)
		ok := true
		for _, v := range c.Order {
			ok = ok && v >= 0 && v < voices && !seen[v]
			if ok {
				seen[v] = true
			}
		}
		if ok {
			return c.Order
		}
	}
	out := make([]int, voices)
	for v := range out {
		out[v] = v
	}
	return out
}

// Fugue is the result: the voices as notes and a record of what happened where
type Fugue struct {
	Key     *scale.Scale
	Config  Config
	Subject Subject
	// Answer is the answer to the subject as made by Answer
	Answer Subject
	// Counter is the countersubject, empty if none could be written (see Exposition)
	Counter Subject
	// CounterAnswer is the countersubject as it is played against the answer, at the pitches of the answer
	// at its standard place. With a real answer it is the countersubject a fifth up; with a tonal answer
	// the notes against the changed head of the answer are rewritten (see CounterAnswer)
	CounterAnswer Subject
	// CounterInvertible tells that the countersubject works above and below the subject and the answer
	CounterInvertible bool
	// Voices holds the notes of every voice, voice 0 the highest, in order of start
	Voices   [][]note.Event
	Entries  []Entry
	Sections []Section
}

// End is the end of the last note of the fugue in steps
func (f *Fugue) End() float64 {
	end := 0.0
	for _, v := range f.Voices {
		for _, e := range v {
			end = max(end, e.Start+e.Duration)
		}
	}
	return end
}

// next is the start of the next section: the end rounded up to a barline
func (f *Fugue) next() float64 {
	bar := f.Config.bar()
	return math.Ceil(f.End()/bar-1e-9) * bar
}

// Bar is the bar (counting from 1) of a step
func (f *Fugue) Bar(step float64) int { return int(math.Floor(step/f.Config.bar()+1e-9)) + 1 }

func (f *Fugue) add(voice int, ev []note.Event) {
	for i := range ev {
		ev[i].Velocity = f.Config.velocity()
	}
	f.Voices[voice] = append(f.Voices[voice], ev...)
	slices.SortStableFunc(f.Voices[voice], func(a, b note.Event) int {
		switch {
		case a.Start < b.Start:
			return -1
		case a.Start > b.Start:
			return 1
		}
		return 0
	})
}

func (f *Fugue) base(k Kind) Subject {
	if k == KindAnswer {
		return f.Answer
	}
	return f.Subject
}

func (f *Fugue) section(name string, start float64) {
	f.Sections = append(f.Sections, Section{name, start, f.End()})
}

// fit picks the shift (the base plus a whole number of octaves) that brings the notes nearest to the center
func fit(key *scale.Scale, degrees []int, base int, center float64) int {
	n := key.Len()
	best, bestDist := base, math.Inf(1)
	for m := -5; m <= 5; m++ {
		sum := 0.0
		for _, d := range degrees {
			sum += key.Degree(d + base + m*n)
		}
		if dist := math.Abs(sum/float64(len(degrees)) - center); dist < bestDist-1e-9 {
			best, bestDist = base+m*n, dist
		}
	}
	return best
}

// Problem is something the checker found wrong
type Problem struct {
	// Entry is the index into Fugue.Entries, or -1 for a problem of a voice
	Entry   int
	Voice   int
	Message string
}

func (p Problem) String() string {
	if p.Entry >= 0 {
		return fmt.Sprintf("entry %d (voice %d): %s", p.Entry, p.Voice, p.Message)
	}
	return fmt.Sprintf("voice %d: %s", p.Voice, p.Message)
}

// ErrBadEntry is wrapped by the problems of Check that name an entry that is not the subject or answer
var ErrBadEntry = errors.New("fugue: entry is not the subject or answer")

// Check verifies that every entry of the fugue is what it says it is: the notes of the voice from the
// start of the entry are the subject or the answer (made again here by Answer from the subject and the
// key) moved by the shift of the entry, with the pitches and the rhythm of the subject. Entries of the
// exposition and the stretto must be moved by whole octaves, so a subject really is on the tonic and an
// answer on the dominant. It also checks that no voice plays two notes at once. An empty result means all
// is right
func (f *Fugue) Check() []Problem {
	var out []Problem
	n := f.Key.Len()
	answer := Answer(f.Subject, f.Key, f.Config.Answer)
	for i, en := range f.Entries {
		bad := func(format string, args ...any) {
			out = append(out, Problem{i, en.Voice, fmt.Sprintf(format, args...)})
		}
		if en.Voice < 0 || en.Voice >= len(f.Voices) {
			bad("no such voice")
			continue
		}
		want := f.Subject
		if en.Kind == KindAnswer {
			want = answer
		}
		if (en.Section == "exposition" || en.Section == "stretto") && en.Shift%n != 0 {
			bad("shift %d is not a whole number of octaves", en.Shift)
		}
		voice := f.Voices[en.Voice]
		first := slices.IndexFunc(voice, func(e note.Event) bool { return math.Abs(e.Start-en.Start) < 1e-9 })
		if first < 0 || first+len(want.Degrees) > len(voice) {
			bad("no notes at step %g", en.Start)
			continue
		}
		t := en.Start
		for j, d := range want.Degrees {
			e := voice[first+j]
			pitch := f.Key.Degree(d + en.Shift)
			switch {
			case math.Abs(e.Start-t) > 1e-9 || math.Abs(e.Duration-want.Durations[j]) > 1e-9:
				bad("note %d has start %g, length %g, want %g, %g (%v)", j, e.Start, e.Duration, t, want.Durations[j], ErrBadEntry)
			case math.Abs(e.Pitch-pitch) > 1e-9:
				bad("note %d is %g, want %g (%v)", j, e.Pitch, pitch, ErrBadEntry)
			}
			t += want.Durations[j]
		}
	}
	for v, voice := range f.Voices {
		for i := 1; i < len(voice); i++ {
			if voice[i].Start < voice[i-1].Start+voice[i-1].Duration-1e-9 {
				out = append(out, Problem{-1, v, fmt.Sprintf("notes %d and %d sound together", i-1, i)})
			}
		}
	}
	return out
}
