package spec

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/almerlucke/genny/scale"
)

// Pitch is a MIDI pitch in semitones, written as a number (60, 61.5) or a note name ("C4" is 60, "F#3",
// "Bb2", "A-1"). Fractions are allowed for microtonal lines (use pitch_bend on the track). A value that
// does not parse is remembered and reported by check.pitch with the path of its field
type Pitch struct {
	v   float64
	err string
}

// F is the pitch in semitones. It panics if the pitch did not parse: kinds must call check.pitch first
func (p Pitch) F() float64 {
	if p.err != "" {
		panic("unchecked pitch: " + p.err)
	}
	return p.v
}

// P makes a valid Pitch
func P(v float64) Pitch { return Pitch{v: v} }

func (p *Pitch) UnmarshalJSON(data []byte) error {
	var f float64
	if err := json.Unmarshal(data, &f); err == nil {
		*p = Pitch{v: f}
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		*p = Pitch{err: "expected a number or a note name, got " + jsonKind(data)}
		return nil
	}
	v, err := ParseNoteName(s)
	if err != nil {
		*p = Pitch{err: err.Error()}
		return nil
	}
	*p = Pitch{v: v}
	return nil
}

func jsonKind(data []byte) string {
	switch strings.TrimSpace(string(data))[0] {
	case '{':
		return "object"
	case '[':
		return "array"
	case '"':
		return "string"
	case 't', 'f':
		return "bool"
	case 'n':
		return "null"
	}
	return "number"
}

// ParseNoteName parses a note name such as "C4" (60), "F#3", "Bb2" or "C-1" (0)
func ParseNoteName(s string) (float64, error) {
	orig := s
	if s == "" {
		return 0, fmt.Errorf("empty note name")
	}
	base := map[byte]int{'C': 0, 'D': 2, 'E': 4, 'F': 5, 'G': 7, 'A': 9, 'B': 11}
	pc, ok := base[s[0]&^0x20] // upper case
	if !ok {
		return 0, fmt.Errorf("invalid note name %q (expected a letter A-G, like C4, F#3 or Bb2)", orig)
	}
	s = s[1:]
	for len(s) > 0 && (s[0] == '#' || s[0] == 'b') {
		if s[0] == '#' {
			pc++
		} else {
			pc--
		}
		s = s[1:]
	}
	var oct int
	if _, err := fmt.Sscanf(s, "%d", &oct); err != nil || fmt.Sprint(oct) != strings.TrimPrefix(s, "+") {
		return 0, fmt.Errorf("invalid note name %q (expected an octave number, like C4, F#3 or Bb2)", orig)
	}
	if oct < -2 || oct > 9 {
		return 0, fmt.Errorf("note name %q has octave %d, expected -2..9", orig, oct)
	}
	return float64((oct+1)*12 + pc), nil
}

// scales by name, as in package scale
var scaleNames = map[string][]float64{
	"major": scale.Major, "minor": scale.Minor, "harmonic_minor": scale.HarmonicMinor,
	"melodic_minor": scale.MelodicMinor, "dorian": scale.Dorian, "phrygian": scale.Phrygian,
	"lydian": scale.Lydian, "mixolydian": scale.Mixolydian, "locrian": scale.Locrian,
	"major_pentatonic": scale.MajorPentatonic, "minor_pentatonic": scale.MinorPentatonic,
	"blues": scale.Blues, "whole_tone": scale.WholeTone, "diminished": scale.Diminished,
	"chromatic": scale.Chromatic,
}

// ScaleNames lists the names of the built in scales
func ScaleNames() []string {
	names := make([]string, 0, len(scaleNames))
	for n := range scaleNames {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// ScaleSpec is a scale given by name ("minor") or as ascending semitone steps within the octave
// ([0, 2, 3, 7, 8])
type ScaleSpec struct {
	Steps []float64
	err   string
}

func (s *ScaleSpec) UnmarshalJSON(data []byte) error {
	*s = ScaleSpec{}
	var name string
	if err := json.Unmarshal(data, &name); err == nil {
		steps, ok := scaleNames[strings.ToLower(name)]
		if !ok {
			s.err = fmt.Sprintf("unknown scale %q (known: %s)", name, strings.Join(ScaleNames(), ", "))
			return nil
		}
		s.Steps = steps
		return nil
	}
	var steps []float64
	if err := json.Unmarshal(data, &steps); err != nil {
		s.err = "expected a scale name or an array of semitone steps, got " + jsonKind(data)
		return nil
	}
	if len(steps) == 0 || len(steps) > 64 {
		s.err = fmt.Sprintf("a scale needs 1..64 steps, got %d", len(steps))
		return nil
	}
	for i, st := range steps {
		if math.IsNaN(st) || st < 0 || st >= 12 || (i > 0 && st <= steps[i-1]) {
			s.err = fmt.Sprintf("scale steps must ascend strictly within [0, 12), got %v", steps)
			return nil
		}
	}
	s.Steps = steps
	return nil
}

// Scale builds the scale on root, falling back to def steps if the spec has none
func (s *ScaleSpec) Scale(root float64, def []float64) *scale.Scale {
	if s != nil && s.err != "" {
		panic("unchecked scale: " + s.err)
	}
	if s == nil || len(s.Steps) == 0 {
		return scale.New(root, def)
	}
	return scale.New(root, s.Steps)
}

// check collects the first problem found among the parameters of one node
type check struct {
	path string
	err  error
}

func newCheck(nodePath string) *check { return &check{path: nodePath + ".params"} }

func (c *check) fail(field, format string, args ...any) {
	if c.err == nil {
		c.err = Fail(c.path+"."+field, format, args...)
	}
}

func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

// pos requires v > 0 (and at most max)
func (c *check) pos(field string, v, max float64) {
	if !finite(v) || v <= 0 {
		c.fail(field, "must be > 0, got %v", v)
	} else if v > max {
		c.fail(field, "must be <= %v, got %v", max, v)
	}
}

// posOrZero allows 0 for "default"
func (c *check) posOrZero(field string, v, max float64) {
	if v != 0 {
		c.pos(field, v, max)
	}
}

func (c *check) rangeF(field string, v, lo, hi float64) {
	if !finite(v) || v < lo || v > hi {
		c.fail(field, "must be in [%v, %v], got %v", lo, hi, v)
	}
}

func (c *check) rangeI(field string, v, lo, hi int) {
	if v < lo || v > hi {
		c.fail(field, "must be in %d..%d, got %d", lo, hi, v)
	}
}

func (c *check) unit(field string, v float64) { c.rangeF(field, v, 0, 1) }

func (c *check) finite(field string, v float64) {
	if !finite(v) || math.Abs(v) > 1e6 {
		c.fail(field, "must be a finite number of at most 1e6, got %v", v)
	}
}

func (c *check) pitch(field string, v Pitch) {
	if v.err != "" {
		c.fail(field, "%s", v.err)
	} else if v.v < 0 || v.v > 127 || !finite(v.v) {
		c.fail(field, "pitch %v is outside the MIDI range 0..127", v.v)
	}
}

// scale reports a scale that did not parse (a nil scale is fine)
func (c *check) scale(field string, s *ScaleSpec) {
	if s != nil && s.err != "" {
		c.fail(field, "%s", s.err)
	}
}

// grid holds the parameters of generators that play notes on a regular grid of steps
type grid struct {
	// Length of the line in steps
	Length float64 `json:"length"`
	// Step is the distance between notes in steps
	Step float64 `json:"step"`
	// NoteLength is the duration of a note in steps (default: step)
	NoteLength float64 `json:"note_length"`
	// Velocity in [0, 1]
	Velocity float64 `json:"velocity"`
}

func defaultGrid() grid { return grid{Step: 1, Velocity: 0.8} }

// check validates the grid and returns the number of notes. A zero length is replaced by def if def > 0
func (g *grid) check(c *check, def float64) int {
	if g.Length == 0 && def > 0 {
		g.Length = def
	}
	c.pos("length", g.Length, maxSteps)
	c.pos("step", g.Step, maxSteps)
	c.posOrZero("note_length", g.NoteLength, maxSteps)
	c.unit("velocity", g.Velocity)
	if c.err != nil {
		return 0
	}
	if g.NoteLength == 0 {
		g.NoteLength = g.Step
	}
	n := int(math.Ceil(g.Length/g.Step - 1e-9))
	if n > maxNodeNotes {
		c.fail("step", "length/step makes %d notes, at most %d are allowed", n, maxNodeNotes)
	}
	return n
}

func paramDocs(extra ...Param) []Param {
	return append(extra,
		Param{"length", "number", "required", "length of the line in steps"},
		Param{"step", "number", "1", "steps between notes"},
		Param{"note_length", "number", "step", "duration of a note in steps"},
		Param{"velocity", "number", "0.8", "velocity in [0, 1]"},
	)
}
