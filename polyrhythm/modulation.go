package polyrhythm

import (
	"github.com/almerlucke/genny/midi"
	"github.com/almerlucke/genny/note"
)

// Modulation is a metric modulation: Old beats of the old tempo last exactly as long as New beats of
// the new tempo. The beat is the unit of the tempo, so with a quarter note beat "dotted quarter =
// quarter" is Equate(F(3, 2), Int(1)) (120 bpm becomes 80), and "quarter = dotted quarter" is
// Equate(Int(1), F(3, 2)) (120 becomes 180). Ratio(3, 2) also gives 120 -> 180. The zero value
// changes nothing.
//
// The new tempo is old*New/Old. Seconds are kept: a duration of d old beats is d*New/Old new beats
type Modulation struct {
	Old, New Frac
}

// Equate returns the modulation in which old beats become new beats: a subdivision of the old
// tempo (a triplet, a quintuplet, a dotted note, in beats) is the beat of the new one. It panics if
// either is not positive
func Equate(old, new Frac) Modulation {
	if old.Sign() <= 0 || new.Sign() <= 0 {
		panic("polyrhythm: a modulation needs positive durations")
	}
	return Modulation{old.norm(), new.norm()}
}

// Ratio returns the modulation that multiplies the tempo by n/m (3/2 speeds up by half). It
// panics if n or m is less than 1
func Ratio(n, m int) Modulation { return Equate(Int(int64(m)), Int(int64(n))) }

func (m Modulation) ok() bool { return m.Old.Sign() > 0 && m.New.Sign() > 0 }

// Factor is the multiplier of the tempo, New/Old (1 for the zero value)
func (m Modulation) Factor() Frac {
	if !m.ok() {
		return Int(1)
	}
	return m.New.Div(m.Old)
}

// Tempo is the new tempo in beats per minute for an old one, exactly
func (m Modulation) Tempo(bpm Frac) Frac { return bpm.Mul(m.Factor()) }

// Inverse undoes the modulation
func (m Modulation) Inverse() Modulation {
	if !m.ok() {
		return Modulation{}
	}
	return Modulation{m.New, m.Old}
}

// Then is the modulation m followed by n
func (m Modulation) Then(n Modulation) Modulation {
	f := m.Factor().Mul(n.Factor())
	return Modulation{Int(1), f}
}

// Carry converts a duration in beats of the old tempo into the beats of the new tempo that last
// the same time in seconds
func (m Modulation) Carry(beats Frac) Frac { return beats.Mul(m.Factor()) }

// CarryEvents returns the events (starts and durations in steps) rescaled to the new tempo, so
// that the same material lasts the same seconds. The step is 1/StepsPerBeat of a beat in both
func (m Modulation) CarryEvents(events []note.Event) []note.Event {
	f := m.Factor().Float()
	out := make([]note.Event, len(events))
	for i, e := range events {
		e.Start *= f
		e.Duration *= f
		out[i] = e
	}
	return out
}

// Section is a stretch of bars at one tempo. Into is the modulation from the previous section into
// this one (ignored for the first section). Num/Den is the time signature (default 4/4), so a bar
// is 4*Num/Den beats
type Section struct {
	Name     string
	Bars     int
	Num, Den int
	Into     Modulation
}

// BarBeats is the length of one bar in beats
func (s Section) BarBeats() Frac {
	n, d := s.Num, s.Den
	if n < 1 || d < 1 {
		n, d = 4, 4
	}
	return F(4*int64(n), int64(d))
}

// Planned is a section placed in time
type Planned struct {
	Section
	FirstBar int  // bar number of its first bar, counted from 0
	Start    Frac // beats from the start of the piece
	Length   Frac // beats
	Tempo    Frac // beats per minute
	Time     Frac // seconds from the start of the piece
}

// End is the start of the next section in beats
func (p Planned) End() Frac { return p.Start.Add(p.Length) }

// Plan is a list of sections placed in time with their tempi
type Plan []Planned

// NewPlan places the sections one after the other starting at the tempo start (beats per minute).
// It panics for a non-positive tempo or a section with fewer than 1 bar
func NewPlan(start Frac, sections ...Section) Plan {
	if start.Sign() <= 0 {
		panic("polyrhythm: the tempo must be positive")
	}
	var plan Plan
	pos, secs, tempo := Int(0), Int(0), start.norm()
	bar := 0
	for i, s := range sections {
		if s.Bars < 1 {
			panic("polyrhythm: a section needs at least 1 bar")
		}
		if i > 0 {
			tempo = s.Into.Tempo(tempo)
		}
		length := s.BarBeats().MulInt(int64(s.Bars))
		plan = append(plan, Planned{s, bar, pos, length, tempo, secs})
		pos = pos.Add(length)
		secs = secs.Add(length.MulInt(60).Div(tempo))
		bar += s.Bars
	}
	return plan
}

// Beats is the total length in beats
func (p Plan) Beats() Frac {
	if len(p) == 0 {
		return Frac{}
	}
	return p[len(p)-1].End()
}

// Seconds is the total length in seconds
func (p Plan) Seconds() Frac { return p.SecondsAt(p.Beats()) }

// SecondsAt is the time in seconds of a position in beats (a position past the end continues at the
// last tempo)
func (p Plan) SecondsAt(beats Frac) Frac {
	if len(p) == 0 {
		return Frac{}
	}
	i := 0
	for i+1 < len(p) && beats.Cmp(p[i+1].Start) >= 0 {
		i++
	}
	return p[i].Time.Add(beats.Sub(p[i].Start).MulInt(60).Div(p[i].Tempo))
}

// BeatsAt is the position in beats of a time in seconds, the inverse of SecondsAt
func (p Plan) BeatsAt(seconds Frac) Frac {
	if len(p) == 0 {
		return Frac{}
	}
	i := 0
	for i+1 < len(p) && seconds.Cmp(p[i+1].Time) >= 0 {
		i++
	}
	return p[i].Start.Add(seconds.Sub(p[i].Time).Mul(p[i].Tempo).Div(Int(60)))
}

// DurationAt is how many beats, counted from the position from, a note must be written to last the
// given seconds: a note that crosses a tempo change keeps its absolute duration
func (p Plan) DurationAt(from, seconds Frac) Frac {
	return p.BeatsAt(p.SecondsAt(from).Add(seconds)).Sub(from)
}

// TempoChanges is the tempo map for midi.File: a change at the start of every section whose tempo
// differs from the one before. The first tempo is the file's Tempo (see Start)
func (p Plan) TempoChanges(stepsPerBeat float64) []midi.TempoChange {
	var out []midi.TempoChange
	for i := 1; i < len(p); i++ {
		if p[i].Tempo != p[i-1].Tempo {
			out = append(out, midi.TempoChange{Step: p[i].Start.Float() * stepsPerBeat, BPM: p[i].Tempo.Float()})
		}
	}
	return out
}

// Start is the first tempo in beats per minute (the Tempo of the midi.File)
func (p Plan) Start() float64 {
	if len(p) == 0 {
		return 120
	}
	return p[0].Tempo.Float()
}

// TimeSignatures is the meter map for midi.File: one entry at every section whose signature
// differs from the one before (and the first)
func (p Plan) TimeSignatures(stepsPerBeat float64) []midi.TimeSignature {
	var out []midi.TimeSignature
	for i, s := range p {
		n, d := s.Num, s.Den
		if n < 1 || d < 1 {
			n, d = 4, 4
		}
		if len(out) > 0 && out[len(out)-1].Num == n && out[len(out)-1].Den == d {
			continue
		}
		out = append(out, midi.TimeSignature{Step: p[i].Start.Float() * stepsPerBeat, Num: n, Den: d})
	}
	return out
}
