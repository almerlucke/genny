package midi

import (
	"cmp"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"maps"
	"math"
	"os"
	"slices"

	"github.com/almerlucke/genny/note"
)

// maxFileSize is the most Read takes from a reader, to be safe against endless input
const maxFileSize = 256 << 20

// DefaultStepsPerBeat is how many steps a beat is divided in by Read (sixteenth notes)
const DefaultStepsPerBeat = 4

// Read reads a Standard MIDI File of format 0 or 1 (and the tracks of format 2 as if they were format 1)
// into a File, with a beat divided in 4 steps. See ReadWith for the details
func Read(r io.Reader) (File, error) { return ReadWith(r, DefaultStepsPerBeat) }

// ReadFile reads the file at path, see Read
func ReadFile(path string) (File, error) { return ReadFileWith(path, DefaultStepsPerBeat) }

// ReadFileWith reads the file at path with a beat divided in stepsPerBeat steps, see ReadWith
func ReadFileWith(path string, stepsPerBeat float64) (File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return File{}, err
	}
	return parseFile(data, stepsPerBeat)
}

// ReadWith reads a Standard MIDI File into a File whose times are in steps, stepsPerBeat of them in a beat
// (File.StepsPerBeat is set to it). A time that is not on the grid becomes a fractional step, nothing is
// rounded. It does what Write undoes:
//
//   - Tempo and the other TempoChanges, and the TimeSignatures come from the meta events of all tracks
//   - every MIDI track with channel messages gives a Track for each channel in it, named after the
//     track name, with Channel and Program (the program changes before the first note, later ones
//     are ProgramChanges), the Controllers, and the note events (a note on with velocity 0 is a note off)
//   - a pitch bend is taken as a bend of up to 2 semitones, and applied as a fractional pitch to the notes
//     after it. A track that has pitch bends has PitchBend set. Smooth bends during a note are not kept
//   - a note that is never switched off lasts until the end of its track
//
// Other messages (aftertouch, sysex, lyrics, ...) are skipped. SMPTE time division is not supported.
// Malformed input gives an error, never a panic
func ReadWith(r io.Reader, stepsPerBeat float64) (File, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxFileSize+1))
	if err != nil {
		return File{}, err
	}
	if len(data) > maxFileSize {
		return File{}, errors.New("midi: file is too large")
	}
	return parseFile(data, stepsPerBeat)
}

type reader struct {
	data []byte
	pos  int
}

var errTruncated = errors.New("midi: unexpected end of data")

func (r *reader) byte() (byte, error) {
	if r.pos >= len(r.data) {
		return 0, errTruncated
	}
	b := r.data[r.pos]
	r.pos++
	return b, nil
}

func (r *reader) bytes(n int) ([]byte, error) {
	if n < 0 || n > len(r.data)-r.pos {
		return nil, errTruncated
	}
	b := r.data[r.pos : r.pos+n]
	r.pos += n
	return b, nil
}

func (r *reader) vlq() (int, error) {
	v := 0
	for range 4 {
		b, err := r.byte()
		if err != nil {
			return 0, err
		}
		v = v<<7 | int(b&0x7f)
		if b&0x80 == 0 {
			return v, nil
		}
	}
	return 0, errors.New("midi: variable length number is too long")
}

// pending is a sounding note
type pending struct {
	start    int // tick
	velocity float64
	pitch    float64
}

// channelTrack collects the Track of one channel in one MIDI track
type channelTrack struct {
	channel int
	track   Track
	bend    int // current pitch bend, 0 is the center
	notes   map[int]pending
	events  []tickEvent
	ctrl    []tickController
	prog    []tickProgram
	firstOn int // tick of the first note on, -1 when there is none
}

type tickEvent struct {
	start, end int
	e          note.Event
}

type tickController struct {
	tick int
	c    Controller
}

type tickProgram struct {
	tick, program int
}

func parseFile(data []byte, stepsPerBeat float64) (File, error) {
	if !(stepsPerBeat > 0) || math.IsInf(stepsPerBeat, 0) {
		return File{}, errors.New("midi: steps per beat must be positive")
	}
	r := &reader{data: data}
	head, err := r.bytes(8)
	if err != nil || string(head[:4]) != "MThd" {
		return File{}, errors.New("midi: not a MIDI file (no MThd header)")
	}
	size := int(binary.BigEndian.Uint32(head[4:]))
	if size < 6 {
		return File{}, errors.New("midi: header is too short")
	}
	hd, err := r.bytes(size)
	if err != nil {
		return File{}, err
	}
	format, ntrks, division := int(binary.BigEndian.Uint16(hd)), int(binary.BigEndian.Uint16(hd[2:])), int(binary.BigEndian.Uint16(hd[4:]))
	if format > 2 {
		return File{}, fmt.Errorf("midi: unknown format %d", format)
	}
	if division&0x8000 != 0 {
		return File{}, errors.New("midi: SMPTE time division is not supported")
	}
	if division == 0 {
		return File{}, errors.New("midi: time division is zero")
	}

	step := func(tick int) float64 { return float64(tick) / float64(division) * stepsPerBeat }
	f := File{StepsPerBeat: stepsPerBeat}
	var tempos []TempoChange
	haveTempo := false
	for i := 0; i < ntrks; i++ {
		id, err := r.bytes(8)
		if err != nil {
			return File{}, fmt.Errorf("track %d: %w", i, err)
		}
		body, err := r.bytes(int(binary.BigEndian.Uint32(id[4:])))
		if err != nil {
			return File{}, fmt.Errorf("track %d: %w", i, err)
		}
		if string(id[:4]) != "MTrk" {
			ntrks++ // an unknown chunk is skipped and does not count as a track
			if ntrks > len(data) {
				return File{}, errors.New("midi: too many chunks")
			}
			continue
		}
		tracks, err := parseTrack(body, step, func(tick int, bpm float64) {
			tempos = append(tempos, TempoChange{Step: step(tick), BPM: bpm})
			haveTempo = true
		}, func(tick, num, den int) {
			f.TimeSignatures = append(f.TimeSignatures, TimeSignature{Step: step(tick), Num: num, Den: den})
		})
		if err != nil {
			return File{}, fmt.Errorf("track %d: %w", i, err)
		}
		f.Tracks = append(f.Tracks, tracks...)
	}

	slices.SortStableFunc(tempos, func(a, b TempoChange) int { return cmp.Compare(a.Step, b.Step) })
	slices.SortStableFunc(f.TimeSignatures, func(a, b TimeSignature) int { return cmp.Compare(a.Step, b.Step) })
	if haveTempo {
		// a tempo at the start is the initial tempo, if there are more at the start the last one counts
		for len(tempos) > 0 && tempos[0].Step == 0 {
			f.Tempo = tempos[0].BPM
			tempos = tempos[1:]
		}
		f.TempoChanges = tempos
	}
	if f.Tempo == 0 {
		f.Tempo = 120 // the MIDI default
	}
	return f, nil
}

func parseTrack(body []byte, step func(int) float64, tempo func(tick int, bpm float64), timeSig func(tick, num, den int)) ([]Track, error) {
	r := &reader{data: body}
	var name string
	var cts []*channelTrack
	byChannel := map[int]*channelTrack{}
	get := func(ch int) *channelTrack {
		if ct := byChannel[ch]; ct != nil {
			return ct
		}
		ct := &channelTrack{channel: ch, notes: map[int]pending{}, firstOn: -1}
		byChannel[ch] = ct
		cts = append(cts, ct)
		return ct
	}

	tick, running := 0, byte(0)
	for r.pos < len(body) {
		delta, err := r.vlq()
		if err != nil {
			return nil, err
		}
		tick += delta

		status, err := r.byte()
		if err != nil {
			return nil, err
		}
		if status >= 0x80 {
			if status < 0xf0 {
				running = status
			} else if status == 0xf0 || status == 0xf7 {
				running = 0
			}
		} else {
			if running == 0 {
				return nil, errors.New("midi: data byte without running status")
			}
			status = running
			r.pos-- // the byte is the first data byte
		}

		switch {
		case status == 0xff:
			typ, err := r.byte()
			if err != nil {
				return nil, err
			}
			n, err := r.vlq()
			if err != nil {
				return nil, err
			}
			d, err := r.bytes(n)
			if err != nil {
				return nil, err
			}
			switch {
			case typ == 0x03 && name == "":
				name = string(d)
			case typ == 0x51 && len(d) == 3:
				if us := int(d[0])<<16 | int(d[1])<<8 | int(d[2]); us > 0 {
					tempo(tick, math.Round(60e6/float64(us)*1000)/1000)
				}
			case typ == 0x58 && len(d) >= 2 && d[0] > 0 && d[1] <= 7:
				timeSig(tick, int(d[0]), 1<<d[1])
			case typ == 0x2f:
				r.pos = len(body) // the track ends here
			}
		case status == 0xf0 || status == 0xf7:
			n, err := r.vlq()
			if err != nil {
				return nil, err
			}
			if _, err := r.bytes(n); err != nil {
				return nil, err
			}
		case status >= 0xf0:
			return nil, fmt.Errorf("midi: unexpected status byte %#x", status)
		default:
			n := 2
			if k := status & 0xf0; k == 0xc0 || k == 0xd0 {
				n = 1
			}
			d, err := r.bytes(n)
			if err != nil {
				return nil, err
			}
			for _, x := range d {
				if x >= 0x80 {
					return nil, errors.New("midi: data byte has the high bit set")
				}
			}
			ct := get(int(status & 0xf))
			ct.message(tick, status&0xf0, d)
		}
	}

	var out []Track
	for _, ct := range cts {
		ct.finish(tick)
		out = append(out, ct.build(name, step))
	}
	return out, nil
}

func (ct *channelTrack) message(tick int, kind byte, d []byte) {
	switch kind {
	case 0x90, 0x80:
		pitch := int(d[0])
		if kind == 0x90 && d[1] > 0 {
			if ct.firstOn < 0 {
				ct.firstOn = tick
			}
			ct.stop(pitch, tick)
			bend := float64(ct.bend) / 8192 * 2
			ct.notes[pitch] = pending{tick, float64(d[1]) / 127, float64(pitch) + bend}
		} else {
			ct.stop(pitch, tick)
		}
	case 0xb0:
		ct.ctrl = append(ct.ctrl, tickController{tick, Controller{Number: int(d[0]), Value: int(d[1])}})
	case 0xc0:
		ct.prog = append(ct.prog, tickProgram{tick, int(d[0])})
	case 0xe0:
		ct.bend = (int(d[0]) | int(d[1])<<7) - 8192
		ct.track.PitchBend = true
	}
}

// stop ends the sounding note of a pitch
func (ct *channelTrack) stop(pitch, tick int) {
	p, ok := ct.notes[pitch]
	if !ok {
		return
	}
	delete(ct.notes, pitch)
	ct.events = append(ct.events, tickEvent{p.start, max(tick, p.start+1), note.Event{Pitch: p.pitch, Velocity: p.velocity}})
}

// finish ends the notes that are still sounding at the end of the track
func (ct *channelTrack) finish(end int) {
	for _, pitch := range slices.Sorted(maps.Keys(ct.notes)) {
		ct.stop(pitch, end)
	}
}

func (ct *channelTrack) build(name string, step func(int) float64) Track {
	t := ct.track
	t.Name, t.Channel = name, ct.channel
	for _, e := range ct.events {
		e.e.Start = step(e.start)
		e.e.Duration = step(e.end) - e.e.Start
		t.Events = append(t.Events, e.e)
	}
	slices.SortStableFunc(t.Events, func(a, b note.Event) int {
		return cmp.Or(cmp.Compare(a.Start, b.Start), cmp.Compare(a.Pitch, b.Pitch))
	})
	for _, c := range ct.ctrl {
		c.c.Step = step(c.tick)
		t.Controllers = append(t.Controllers, c.c)
	}
	// the program changes up to the first note (or all of them when there are no notes) are the starting
	// program, the last of them counts
	for _, p := range ct.prog {
		if ct.firstOn < 0 || p.tick <= ct.firstOn {
			t.Program = p.program
		} else {
			t.ProgramChanges = append(t.ProgramChanges, ProgramChange{Step: step(p.tick), Program: p.program})
		}
	}
	return t
}
