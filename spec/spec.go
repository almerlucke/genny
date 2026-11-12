// Package spec renders a piece described in JSON to a MIDI file. A spec has a tempo, a seed and tracks; the
// source of every track is a tree of nodes, each {"kind": ..., "params": {...}}: generator kinds (melody,
// euclid, sieve, row, ...) make a line of notes and combinator kinds (concat, repeat, transpose, layer, ...)
// work on whole lines. The kinds are in one table (see Kind and Register), which also documents them
// for the command line tool genny.
//
// Everything is validated before it is used: unknown fields are rejected, and every error quotes the JSON
// path of the problem, for instance "tracks[1].source.params.length: must be > 0". Rendering is
// deterministic: the same spec and seed give the same file. Malformed input gives errors, never panics
package spec

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"slices"
	"strings"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/midi"
	"github.com/almerlucke/genny/note"
)

// Limits that keep a spec from asking for absurd amounts of work
const (
	maxTracks      = 64
	maxSteps       = 100000   // length of any line in steps
	maxNodeNotes   = 200000   // notes of one generator
	maxTotalNotes  = 1000000  // notes of a whole render
	maxDepth       = 32       // nesting of nodes
	maxSpecBytes   = 16 << 20 // size of a spec file
	defaultSeed    = 1
	defaultTempo   = 120
	defaultSPB     = 4
	maxTempo       = 1000
	maxStepsPerBar = 64
)

// Error is a problem in a spec, at a JSON path (empty for the document as a whole)
type Error struct {
	Path string
	Msg  string
}

func (e *Error) Error() string {
	if e.Path == "" {
		return e.Msg
	}
	return e.Path + ": " + e.Msg
}

// Fail makes an *Error at path
func Fail(path, format string, args ...any) error {
	return &Error{Path: path, Msg: fmt.Sprintf(format, args...)}
}

// Spec is a piece
type Spec struct {
	// Tempo in beats per minute (default 120)
	Tempo float64 `json:"tempo"`
	// StepsPerBeat is how many steps fit in a beat (default 4, sixteenth notes)
	StepsPerBeat float64 `json:"steps_per_beat"`
	// Seed seeds the random stream (default 1); the command line can override it
	Seed *uint64 `json:"seed"`
	// TimeSignature is information for the player (default none)
	TimeSignature *TimeSignature `json:"time_signature"`
	Tracks        []json.RawMessage
}

// TimeSignature is a meter such as 3/4
type TimeSignature struct {
	Num int `json:"num"`
	Den int `json:"den"`
}

// Track is one MIDI track with the tree that makes its notes
type Track struct {
	Name string `json:"name"`
	// Channel 0..15 (9 is the drum channel)
	Channel int `json:"channel"`
	// Program is the instrument 0..127
	Program int `json:"program"`
	// PitchBend plays fractional pitches with pitch bends (only for single note lines)
	PitchBend bool            `json:"pitch_bend"`
	Source    json.RawMessage `json:"source"`
}

// specDoc is how a Spec is read: tracks stay raw so that errors can name the track by index
type specDoc struct {
	Tempo         float64           `json:"tempo"`
	StepsPerBeat  float64           `json:"steps_per_beat"`
	Seed          *uint64           `json:"seed"`
	TimeSignature *TimeSignature    `json:"time_signature"`
	Tracks        []json.RawMessage `json:"tracks"`
}

// Parse reads a spec, rejecting unknown fields and checking the header. The tracks are checked by Render
func Parse(r io.Reader) (*Spec, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxSpecBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxSpecBytes {
		return nil, Fail("", "spec is larger than %d bytes", maxSpecBytes)
	}
	return ParseBytes(data)
}

// ParseBytes is like Parse for data in memory
func ParseBytes(data []byte) (*Spec, error) {
	if !json.Valid(data) {
		// find the syntax error for a position
		var v any
		err := json.Unmarshal(data, &v)
		var se *json.SyntaxError
		if errors.As(err, &se) {
			line, col := lineCol(data, se.Offset)
			return nil, Fail("", "line %d, column %d: %s", line, col, se.Error())
		}
		return nil, Fail("", "invalid JSON: %v", err)
	}
	var doc specDoc
	if err := decodeStrict("", data, &doc); err != nil {
		return nil, err
	}
	s := &Spec{Tempo: doc.Tempo, StepsPerBeat: doc.StepsPerBeat, Seed: doc.Seed, TimeSignature: doc.TimeSignature, Tracks: doc.Tracks}
	if err := s.checkHeader(); err != nil {
		return nil, err
	}
	return s, nil
}

func lineCol(data []byte, offset int64) (line, col int) {
	line, col = 1, 1
	for i := 0; i < int(offset) && i < len(data); i++ {
		if data[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return
}

func (s *Spec) checkHeader() error {
	if s.Tempo != 0 && !(s.Tempo > 0 && s.Tempo <= maxTempo) {
		return Fail("tempo", "must be in (0, %d] beats per minute, got %v", maxTempo, s.Tempo)
	}
	if s.StepsPerBeat != 0 && !(s.StepsPerBeat > 0 && s.StepsPerBeat <= 64) {
		return Fail("steps_per_beat", "must be in (0, 64], got %v", s.StepsPerBeat)
	}
	if ts := s.TimeSignature; ts != nil {
		if ts.Num < 1 || ts.Num > 255 {
			return Fail("time_signature.num", "must be in 1..255, got %d", ts.Num)
		}
		if ts.Den < 1 || ts.Den > 128 || ts.Den&(ts.Den-1) != 0 {
			return Fail("time_signature.den", "must be a power of two up to 128, got %d", ts.Den)
		}
	}
	if len(s.Tracks) == 0 {
		return Fail("tracks", "must have at least one track")
	}
	if len(s.Tracks) > maxTracks {
		return Fail("tracks", "at most %d tracks, got %d", maxTracks, len(s.Tracks))
	}
	return nil
}

// Result is a rendered piece
type Result struct {
	File midi.File
	// Seed is the seed that was used
	Seed uint64
	// Notes is the total number of notes
	Notes int
	// Steps is the length of the longest track in steps
	Steps float64
}

// Seconds is the duration of the piece at its tempo (ignoring tempo changes, there are none)
func (r Result) Seconds() float64 { return r.Steps / r.File.StepsPerBeat * 60 / r.File.Tempo }

// Render builds all tracks. If seed is not nil it overrides the seed of the spec. Errors of all tracks
// that are wrong are reported together (joined), each with its JSON path
func (s *Spec) Render(seed *uint64) (res Result, err error) {
	used := uint64(defaultSeed)
	if s.Seed != nil {
		used = *s.Seed
	}
	if seed != nil {
		used = *seed
	}
	genny.Seed(used)

	file := midi.File{Tempo: s.Tempo, StepsPerBeat: s.StepsPerBeat}
	if file.Tempo == 0 {
		file.Tempo = defaultTempo
	}
	if file.StepsPerBeat == 0 {
		file.StepsPerBeat = defaultSPB
	}
	if ts := s.TimeSignature; ts != nil {
		file.TimeSignatures = []midi.TimeSignature{{Num: ts.Num, Den: ts.Den}}
	}

	b := &Builder{}
	var errs []error
	for i, raw := range s.Tracks {
		path := fmt.Sprintf("tracks[%d]", i)
		track, length, err := b.track(path, raw)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		file.Tracks = append(file.Tracks, track)
		res.Notes += len(track.Events)
		res.Steps = max(res.Steps, length)
	}
	if len(errs) > 0 {
		return Result{}, errors.Join(errs...)
	}
	res.File, res.Seed = file, used
	return res, nil
}

// Validate checks a spec completely by building it (the result is thrown away)
func (s *Spec) Validate(seed *uint64) error {
	_, err := s.Render(seed)
	return err
}

func (b *Builder) track(path string, raw json.RawMessage) (t midi.Track, length float64, err error) {
	var tr Track
	if err := decodeStrict(path, raw, &tr); err != nil {
		return t, 0, err
	}
	if tr.Channel < 0 || tr.Channel > 15 {
		return t, 0, Fail(path+".channel", "must be in 0..15, got %d", tr.Channel)
	}
	if tr.Program < 0 || tr.Program > 127 {
		return t, 0, Fail(path+".program", "must be in 0..127, got %d", tr.Program)
	}
	if len(tr.Source) == 0 {
		return t, 0, Fail(path+".source", "is required")
	}
	line, err := b.Child(path+".source", tr.Source)
	if err != nil {
		return t, 0, err
	}
	events := slices.Clone(line.Events)
	for i := range events {
		e := &events[i]
		if math.IsNaN(e.Pitch) || e.Pitch < 0 || e.Pitch > 127.99 {
			return t, 0, Fail(path+".source", "note %d (step %.4g) has pitch %.4g, outside the MIDI range 0..127: transpose the source", i, e.Start, e.Pitch)
		}
		if !(e.Start >= 0 && e.Start < 1e8) {
			return t, 0, Fail(path+".source", "note %d has start %v outside 0..1e8 steps", i, e.Start)
		}
		e.Velocity = min(max(e.Velocity, 0), 1)
		if !(e.Duration > 0) {
			e.Duration = 1e-3
		}
		e.Duration = min(e.Duration, 1e8)
	}
	slices.SortStableFunc(events, func(a, c note.Event) int {
		switch {
		case a.Start < c.Start:
			return -1
		case a.Start > c.Start:
			return 1
		}
		return 0
	})
	end := line.Length
	for _, e := range events {
		end = max(end, e.Start+e.Duration)
	}
	return midi.Track{Name: tr.Name, Channel: tr.Channel, Program: tr.Program, PitchBend: tr.PitchBend, Events: events}, end, nil
}

// decodeStrict decodes raw into v, rejecting unknown fields, with errors that quote path
func decodeStrict(path string, raw []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	err := dec.Decode(v)
	if err == nil {
		if _, err2 := dec.Token(); err2 != io.EOF {
			return Fail(path, "unexpected data after the JSON value")
		}
		return nil
	}
	var te *json.UnmarshalTypeError
	if errors.As(err, &te) {
		p := path
		if te.Field != "" {
			p = join(path, te.Field)
		}
		return Fail(p, "expected %s, got %s", typeName(te.Type), te.Value)
	}
	if msg, ok := strings.CutPrefix(err.Error(), "json: unknown field "); ok {
		name := strings.Trim(msg, `"`)
		return Fail(join(path, name), "unknown field")
	}
	if errors.Is(err, io.EOF) {
		return Fail(path, "is empty")
	}
	return Fail(path, "%s", strings.TrimPrefix(err.Error(), "json: "))
}

func join(path, field string) string {
	if path == "" {
		return field
	}
	return path + "." + field
}

func typeName(t interface{ String() string }) string {
	s := t.String()
	switch {
	case strings.HasPrefix(s, "[]"):
		return "an array"
	case strings.HasPrefix(s, "float"), strings.HasPrefix(s, "int"), strings.HasPrefix(s, "uint"):
		return "a number"
	case s == "string":
		return "a string"
	case s == "bool":
		return "true or false"
	case strings.HasPrefix(s, "spec."), strings.HasPrefix(s, "struct"), strings.HasPrefix(s, "*spec."):
		return "an object"
	}
	return s
}
