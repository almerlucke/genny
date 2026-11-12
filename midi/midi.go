// Package midi writes note events (package note) to a Standard MIDI File, so you can hear what the
// generators make in any DAW or player
package midi

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"io"
	"math"
	"math/bits"
	"os"
	"slices"

	"github.com/almerlucke/genny/note"
)

// ticksPerBeat is the resolution of the file
const ticksPerBeat = 480

// Track is a track of notes on one MIDI channel
type Track struct {
	Name string
	// Channel is 0..15 (channel 10 for drums is 9, which gets no program change)
	Channel int
	// Program is the instrument 0..127 (General MIDI: 0 piano, 33 fingered bass, ...)
	Program int
	Events  []note.Event
	// PitchBend plays fractional pitches (microtonal scales) with a pitch bend before every note, assuming the
	// usual bend range of 2 semitones. Without it pitches are rounded to the nearest semitone. Only
	// for single note lines: a pitch bend moves all sounding notes of the channel
	PitchBend bool
	// Controllers are control changes (volume 7, pan 10, modulation 1, ...) at a step of the track
	Controllers []Controller
	// ProgramChanges switch the instrument in the middle of the track, after the starting Program
	ProgramChanges []ProgramChange
}

// Controller is a control change on the channel of a track. Number and Value are in 0..127
type Controller struct {
	Step          float64
	Number, Value int
}

// ProgramChange switches the instrument of a track at a step. Program is in 0..127
type ProgramChange struct {
	Step    float64
	Program int
}

// TempoChange sets a new tempo in beats per minute from a step on
type TempoChange struct {
	Step float64
	BPM  float64
}

// TimeSignature is a meter from a step on: Num beats of a Den note, so 6/8 is {Num: 6, Den: 8}. Den is a
// power of two up to 128. It is only information for the player, it does not change the timing
type TimeSignature struct {
	Step     float64
	Num, Den int
}

// File is a MIDI file of one or more tracks
type File struct {
	// Tempo in beats per minute (default 120)
	Tempo float64
	// StepsPerBeat is how many steps of the events fit in a beat (default 4, sixteenth notes)
	StepsPerBeat float64
	Tracks       []Track
	// TempoChanges is the tempo map: new tempi after the initial Tempo
	TempoChanges []TempoChange
	// TimeSignatures are the meters of the file (default none, players assume 4/4)
	TimeSignatures []TimeSignature
}

type message struct {
	tick  int
	order int // at the same tick: note off, then pitch bend, then note on
	data  []byte
}

// okStep tells if a step can be turned into a tick
func okStep(s float64) bool { return s >= 0 && s < 1e9 }

// tempoMeta is the tempo meta event (without delta time) of a tempo in beats per minute
func tempoMeta(bpm float64) []byte {
	us := min(int(math.Round(60e6/bpm)), 0xffffff)
	return meta(0, 0x51, []byte{byte(us >> 16), byte(us >> 8), byte(us)})[1:]
}

func vlq(n int) []byte {
	b := []byte{byte(n & 0x7f)}
	for n >>= 7; n > 0; n >>= 7 {
		b = append([]byte{byte(n&0x7f) | 0x80}, b...)
	}
	return b
}

func meta(delta int, typ byte, data []byte) []byte {
	out := append(vlq(delta), 0xff, typ)
	out = append(out, vlq(len(data))...)
	return append(out, data...)
}

func chunk(id string, body []byte) []byte {
	out := []byte(id)
	n := len(body)
	out = append(out, byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
	return append(out, body...)
}

func (t Track) body(tick func(float64) int) ([]byte, error) {
	if t.Channel < 0 || t.Channel > 15 {
		return nil, fmt.Errorf("midi: channel %d is not in 0..15", t.Channel)
	}
	if t.Program < 0 || t.Program > 127 {
		return nil, fmt.Errorf("midi: program %d is not in 0..127", t.Program)
	}

	ch := byte(t.Channel)
	var msgs []message
	for _, e := range t.Events {
		if math.IsNaN(e.Pitch) || math.IsNaN(e.Start) || math.IsNaN(e.Duration) || e.Start < 0 || e.Duration <= 0 {
			return nil, fmt.Errorf("midi: invalid event %+v", e)
		}
		pitch := min(max(int(math.Round(e.Pitch)), 0), 127)
		vel := min(max(int(math.Round(e.Velocity*127)), 1), 127)
		on := tick(e.Start)
		off := max(tick(e.Start+e.Duration), on+1)

		if t.PitchBend {
			bend := 8192 + int(math.Round((e.Pitch-float64(pitch))/2*8192))
			bend = min(max(bend, 0), 16383)
			msgs = append(msgs, message{on, 1, []byte{0xe0 | ch, byte(bend & 0x7f), byte(bend >> 7)}})
		}
		msgs = append(msgs,
			message{on, 2, []byte{0x90 | ch, byte(pitch), byte(vel)}},
			message{off, 0, []byte{0x80 | ch, byte(pitch), 0}},
		)
	}
	var out []byte
	if t.Name != "" {
		out = append(out, meta(0, 0x03, []byte(t.Name))...)
	}
	if t.Channel != 9 {
		out = append(out, 0, 0xc0|ch, byte(t.Program))
	}
	for _, c := range t.Controllers {
		if !okStep(c.Step) || c.Number < 0 || c.Number > 127 || c.Value < 0 || c.Value > 127 {
			return nil, fmt.Errorf("midi: invalid controller %+v", c)
		}
		msgs = append(msgs, message{tick(c.Step), 1, []byte{0xb0 | ch, byte(c.Number), byte(c.Value)}})
	}
	for _, c := range t.ProgramChanges {
		if !okStep(c.Step) || c.Program < 0 || c.Program > 127 {
			return nil, fmt.Errorf("midi: invalid program change %+v", c)
		}
		msgs = append(msgs, message{tick(c.Step), 1, []byte{0xc0 | ch, byte(c.Program)}})
	}
	slices.SortStableFunc(msgs, func(a, b message) int {
		return cmp.Or(cmp.Compare(a.tick, b.tick), cmp.Compare(a.order, b.order))
	})

	last := 0
	for _, m := range msgs {
		out = append(out, vlq(m.tick-last)...)
		out = append(out, m.data...)
		last = m.tick
	}
	return append(out, meta(0, 0x2f, nil)...), nil
}

// Write writes the file in format 1: a tempo track and a track for every Track
func Write(w io.Writer, f File) error {
	tempo, spb := f.Tempo, f.StepsPerBeat
	if tempo == 0 {
		tempo = 120
	}
	if spb == 0 {
		spb = 4
	}
	if !(tempo > 0) || !(spb > 0) {
		return errors.New("midi: tempo and steps per beat must be positive")
	}
	tick := func(steps float64) int { return int(math.Round(steps / spb * ticksPerBeat)) }

	tempoMsgs := []message{{0, 0, tempoMeta(tempo)}}
	for _, c := range f.TempoChanges {
		if !okStep(c.Step) || !(c.BPM > 0) || c.BPM > 1e6 {
			return fmt.Errorf("midi: invalid tempo change %+v", c)
		}
		tempoMsgs = append(tempoMsgs, message{tick(c.Step), 1, tempoMeta(c.BPM)})
	}
	for _, s := range f.TimeSignatures {
		dd := bits.TrailingZeros(uint(max(s.Den, 0)))
		if !okStep(s.Step) || s.Num < 1 || s.Num > 255 || s.Den < 1 || s.Den > 128 || s.Den != 1<<dd {
			return fmt.Errorf("midi: invalid time signature %+v", s)
		}
		tempoMsgs = append(tempoMsgs, message{tick(s.Step), 2, meta(0, 0x58, []byte{byte(s.Num), byte(dd), 24, 8})[1:]})
	}
	slices.SortStableFunc(tempoMsgs, func(a, b message) int {
		return cmp.Or(cmp.Compare(a.tick, b.tick), cmp.Compare(a.order, b.order))
	})
	var tempoTrack []byte
	last := 0
	for _, m := range tempoMsgs {
		tempoTrack = append(tempoTrack, vlq(m.tick-last)...)
		tempoTrack = append(tempoTrack, m.data...)
		last = m.tick
	}
	tempoTrack = append(tempoTrack, meta(0, 0x2f, nil)...)

	var buf bytes.Buffer
	n := len(f.Tracks) + 1
	buf.Write(chunk("MThd", []byte{0, 1, byte(n >> 8), byte(n), byte(ticksPerBeat >> 8), byte(ticksPerBeat & 0xff)}))
	buf.Write(chunk("MTrk", tempoTrack))
	for i, t := range f.Tracks {
		body, err := t.body(tick)
		if err != nil {
			return fmt.Errorf("track %d: %w", i, err)
		}
		buf.Write(chunk("MTrk", body))
	}

	_, err := w.Write(buf.Bytes())
	return err
}

// WriteFile writes the file to path
func WriteFile(path string, f File) error {
	var buf bytes.Buffer
	if err := Write(&buf, f); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}
