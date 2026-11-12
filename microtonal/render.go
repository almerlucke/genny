package microtonal

import (
	"cmp"
	"math"
	"slices"

	"github.com/almerlucke/genny/midi"
	"github.com/almerlucke/genny/note"
)

// Bend values: 14 bit, 8192 is no bend. The midi package writes the bend of a Track.PitchBend from the
// fraction of the pitch (as if the range is 2 semitones), so Render keeps the bend within [8192-2048,
// 8192+2047], a fraction below half a unit, and chooses the neighbouring note when more is needed
const (
	centre  = 8192
	maxUp   = 2047
	maxDown = 2048
)

// Strategy says how fractional pitches reach the MIDI channels
type Strategy struct {
	// MPE gives every simultaneous note its own member channel (rotating through channels 2..16, the one
	// free the longest first), and sends the MPE configuration on channel 1. Without it all notes go to
	// Channel, which only works for a single line of notes
	MPE bool
	// Range is the pitch bend range in semitones the synth is told to use (RPN 0), from 2 up to 96
	// (default 2). A wider range makes the bend steps coarser (0.024 cents at 2, 0.15 at 12)
	Range float64
	// Channel is the channel 0..15 of a single line (default 0, ignored with MPE)
	Channel int
	// Members is the number of MPE member channels 1..15 (default 15)
	Members int
	// Program is the instrument 0..127 of all tracks
	Program int
	// Name names the tracks
	Name string
}

// Bend is a note with a pitch bend: the MIDI note number and the 14 bit bend value (8192 is no bend)
type Bend struct {
	Note, Value int
}

// Locate finds the MIDI note and 14 bit pitch bend for a pitch in semitones with the given bend range in
// semitones (range 2 if it is smaller). The nearest note is used unless the bend would not reach (or,
// for the exact half semitone, be written reliably), then the neighbour. ok is false if the note lies
// outside 0..127 or the pitch is not a number. Cents(b, range) turns the result back to a pitch
func Locate(pitch, bendRange float64) (b Bend, ok bool) {
	if math.IsNaN(pitch) || math.IsInf(pitch, 0) || pitch < -0.5 || pitch >= 127.5 {
		return Bend{}, false
	}
	bendRange = max(bendRange, 2)
	n := int(math.Round(pitch))
	u := int(math.Round((pitch - float64(n)) / bendRange * 8192))
	if u > maxUp {
		n++
		u = int(math.Round((pitch - float64(n)) / bendRange * 8192))
	}
	if n < 0 || n > 127 {
		return Bend{}, false
	}
	u = min(max(u, -maxDown), maxUp)
	return Bend{n, centre + u}, true
}

// Pitch is the pitch in semitones that a note with a bend sounds at, with the bend range in semitones
func (b Bend) Pitch(bendRange float64) float64 {
	return float64(b.Note) + float64(b.Value-centre)/8192*max(bendRange, 2)
}

// Cents is the offset in cents of the bend from the note
func (b Bend) Cents(bendRange float64) float64 {
	return float64(b.Value-centre) / 8192 * max(bendRange, 2) * 100
}

func (s Strategy) rangeOrDefault() float64 {
	if s.Range < 2 || math.IsNaN(s.Range) {
		return 2
	}
	return min(s.Range, 96)
}

func rpn(step float64, lsb, msb, dataMSB, dataLSB int) []midi.Controller {
	c := []midi.Controller{{Step: step, Number: 101, Value: msb}, {Step: step, Number: 100, Value: lsb}, {Step: step, Number: 6, Value: dataMSB}}
	if dataLSB >= 0 {
		c = append(c, midi.Controller{Step: step, Number: 38, Value: dataLSB})
	}
	// RPN null, so later data entry changes nothing
	return append(c, midi.Controller{Step: step, Number: 101, Value: 127}, midi.Controller{Step: step, Number: 100, Value: 127})
}

// rangeSetup sets the pitch bend range (RPN 0) of a channel
func (s Strategy) rangeSetup() []midi.Controller {
	r := s.rangeOrDefault()
	semis := int(math.Floor(r))
	cents := int(math.Round((r - float64(semis)) * 100))
	return rpn(0, 0, 0, semis, cents)
}

// Render turns note events into tracks for midi.File. The pitches are in semitones; if tuning is not nil each
// pitch is first snapped to the nearest degree of the tuning (pitches made by scale.Scale of the same
// tuning stay as they are). Notes with an invalid event or outside the MIDI range are skipped, and so are
// notes that find no free channel with MPE (more simultaneous notes than member channels), see Plan
func Render(events []note.Event, tuning *Tuning, s Strategy) []midi.Track {
	tracks, _ := Plan(events, tuning, s)
	return tracks
}

// Plan is Render that also returns how many notes were left out
func Plan(events []note.Event, tuning *Tuning, s Strategy) (tracks []midi.Track, skipped int) {
	r := s.rangeOrDefault()
	evs := slices.Clone(events)
	slices.SortStableFunc(evs, func(a, b note.Event) int { return cmp.Compare(a.Start, b.Start) })

	members := s.Members
	if members < 1 || members > 15 {
		members = 15
	}
	byChannel := map[int][]note.Event{}
	freeAt := make([]float64, 16) // end of the last note on a channel
	for _, e := range evs {
		if math.IsNaN(e.Start) || math.IsNaN(e.Duration) || e.Start < 0 || e.Duration <= 0 {
			skipped++
			continue
		}
		p := e.Pitch
		if tuning != nil {
			p = tuning.Snap(p)
		}
		b, ok := Locate(p, r)
		if !ok {
			skipped++
			continue
		}
		ch := min(max(s.Channel, 0), 15)
		if s.MPE {
			ch = -1
			for c := 1; c <= members; c++ {
				if freeAt[c] <= e.Start && (ch < 0 || freeAt[c] < freeAt[ch]) {
					ch = c
				}
			}
			if ch < 0 {
				skipped++
				continue
			}
		}
		freeAt[ch] = e.Start + e.Duration
		// the midi package derives the bend from the fraction of the pitch, assuming a range of 2
		e.Pitch = float64(b.Note) + float64(b.Value-centre)/4096
		byChannel[ch] = append(byChannel[ch], e)
	}

	name := s.Name
	if name == "" {
		name = "microtonal"
	}
	if s.MPE {
		// the lower zone: channel 1 manages the member channels 2..members+1 (RPN 6)
		tracks = append(tracks, midi.Track{Name: name + " manager", Channel: 0, Program: s.Program, Controllers: rpn(0, 6, 0, members, -1)})
	}
	for ch := range 16 {
		if evs, ok := byChannel[ch]; ok {
			tracks = append(tracks, midi.Track{Name: name, Channel: ch, Program: s.Program, Events: evs, PitchBend: true, Controllers: s.rangeSetup()})
		}
	}
	return tracks, skipped
}
