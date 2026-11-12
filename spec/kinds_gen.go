package spec

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/choose"
	"github.com/almerlucke/genny/chord"
	"github.com/almerlucke/genny/euclid"
	"github.com/almerlucke/genny/float/tendency"
	"github.com/almerlucke/genny/melody"
	"github.com/almerlucke/genny/note"
	"github.com/almerlucke/genny/numseq"
	"github.com/almerlucke/genny/progression"
	"github.com/almerlucke/genny/pulse"
	"github.com/almerlucke/genny/row"
	"github.com/almerlucke/genny/scale"
	"github.com/almerlucke/genny/sequence"
	"github.com/almerlucke/genny/sieve"
	"github.com/almerlucke/genny/stochastic"
)

func init() {
	Register(Kind{Name: "notes", Category: "generator", Build: buildNotes,
		Doc: "literal notes: [{pitch, start, duration, velocity}]",
		Params: []Param{
			{"notes", "array", "required", "objects with pitch (number or name), start (steps), duration (default 1), velocity (default 0.8)"},
			{"length", "number", "end of the last note", "length of the line in steps"},
		}})
	Register(Kind{Name: "melody", Category: "generator", Build: buildMelody,
		Doc: "a melodic walk over scale degrees: mostly steps, sometimes leaps (package scale)",
		Params: paramDocs(
			Param{"scale", "scale", "major", "name or array of semitone steps"},
			Param{"root", "pitch", "60", "root of the scale, degree 0"},
			Param{"start", "int", "0", "first degree"},
			Param{"low", "int", "-7", "lowest degree"},
			Param{"high", "int", "7", "highest degree"},
			Param{"max_step", "int", "2", "largest ordinary move in degrees"},
			Param{"leap_prob", "number", "0.15", "chance of a leap"},
			Param{"max_leap", "int", "4", "largest leap in degrees"},
			Param{"rest_prob", "number", "0", "chance of a rest instead of a note"},
		)})
	Register(Kind{Name: "chordmelody", Category: "generator", Build: buildChordMelody,
		Doc: "a melody over a chord progression: chord tones on strong beats, scale tones between (package melody)",
		Params: harmonyDocs(
			Param{"beat_steps", "int", "4", "steps in a beat; must divide bar_steps"},
			Param{"low", "pitch", "60", "lowest pitch, at least an octave below high"},
			Param{"high", "pitch", "84", "highest pitch"},
			Param{"max_leap", "int", "3", "largest leap in scale degrees"},
		)})
	Register(Kind{Name: "progression", Category: "generator", Build: buildProgression,
		Doc: "chords, roots or arpeggios over a progression (a Markov chain over scale degrees, or your own degrees)",
		Params: harmonyDocs(
			Param{"style", "string", "block", "block (held chords), root (root notes) or arpeggio"},
			Param{"step", "number", "2", "steps between notes of an arpeggio"},
			Param{"octave", "int", "0", "shift in octaves"},
		)})
	Register(Kind{Name: "euclid", Category: "generator", Build: buildEuclid,
		Doc: "a Euclidean rhythm: pulses spread evenly over steps, repeated (package euclid)",
		Params: paramDocs(
			Param{"steps", "int", "required", "steps of the pattern, 1..256"},
			Param{"pulses", "int", "required", "pulses, 0..steps"},
			Param{"rotation", "int", "0", "rotate the pattern left"},
			Param{"pitch", "pitch", "36", "pitch of every pulse"},
		)})
	Register(Kind{Name: "sieve", Category: "generator", Build: buildSieve,
		Doc: "Xenakis sieves: an expression of residue classes, used as rhythm gate or pitch set (package sieve)",
		Params: paramDocs(
			Param{"expr", "string", "required", "residue classes mod@start joined with & | - ! and (), e.g. (3@0|5@1)&!2@0"},
			Param{"lo", "int", "0", "lowest member of the range"},
			Param{"hi", "int", "15", "highest member of the range"},
			Param{"mode", "string", "gate", "gate (a note on each step that is a member, cycle hi+1) or pitches (members played in turn, looping)"},
			Param{"pitch", "pitch", "60", "gate mode: pitch of the notes"},
			Param{"root", "pitch", "60", "pitches mode: pitch of member 0"},
			Param{"scale", "scale", "none", "pitches mode: members are scale degrees instead of semitones"},
		)})
	Register(Kind{Name: "row", Category: "generator", Build: buildRow,
		Doc: "a tone row played in forms (prime, retrograde, inversion, retrograde_inversion) (package row)",
		Params: paramDocs(
			Param{"row", "array", "random", "pitch classes 0..11; omitted: a random row of `size` notes"},
			Param{"size", "int", "12", "size of the random row"},
			Param{"forms", "array", "[prime]", "objects {op, transpose} played one after the other, looping"},
			Param{"root", "pitch", "60", "added to every pitch class"},
		)})
	Register(Kind{Name: "numseq", Category: "generator", Build: buildNumseq,
		Doc: "number sequences as scale degrees (package numseq)",
		Params: paramDocs(
			Param{"seq", "string", "required", "fibonacci, collatz, thue_morse or primes"},
			Param{"mod", "int", "7", "values are reduced modulo mod before they become degrees"},
			Param{"start", "int", "7", "collatz: start number"},
			Param{"scale", "scale", "major", "scale for the degrees"},
			Param{"root", "pitch", "60", "root of the scale"},
		)})
	Register(Kind{Name: "pulse", Category: "generator", Build: buildPulse,
		Doc: "metric hierarchy (indispensability): thinned by a level, velocity follows importance (package pulse)",
		Params: paramDocs(
			Param{"divisions", "array", "required", "nested divisions of the meter, e.g. [2,2,2,2] for 16 steps or [3,2,2]"},
			Param{"level", "int", "0", "pulses less important than this rest; 0 plays all"},
			Param{"velocity_low", "number", "0.4", "velocity of the weakest pulse (velocity is that of the first)"},
			Param{"pitch", "pitch", "42", "pitch of the notes"},
		)})
	Register(Kind{Name: "tendency", Category: "generator", Build: buildTendency,
		Doc: "Koenig tendency mask: random pitches inside a band that moves (package float/tendency)",
		Params: paramDocs(
			Param{"segments", "array", "required", "objects {n, low, high, end_low, end_high}: over n notes the band glides from low..high to end_low..end_high (default: stays)"},
			Param{"loop", "bool", "false", "start the mask again after the last segment"},
			Param{"scale", "scale", "none", "quantize the pitches to this scale"},
			Param{"root", "pitch", "60", "root of the scale"},
			Param{"beta", "array", "uniform", "[a, b] beta distribution: a<b favours the low edge, a>b the high edge"},
		)})
	Register(Kind{Name: "choose", Category: "generator", Build: buildChoose,
		Doc: "weighted random choice of pitches (or rests) on every step (package choose)",
		Params: paramDocs(
			Param{"options", "array", "required", "objects {pitch, weight}; pitch null is a rest; weight default 1"},
		)})
	Register(Kind{Name: "cloud", Category: "generator", Build: buildCloud,
		Doc: "stochastic cloud: Poisson onsets with a density that can ramp, uniform pitches (package stochastic)",
		Params: []Param{
			{"length", "number", "required", "length in steps"},
			{"density", "number", "required", "mean events per step"},
			{"density_end", "number", "density", "density at the end: the density ramps linearly"},
			{"low", "pitch", "36", "lowest pitch"},
			{"high", "pitch", "84", "highest pitch"},
			{"scale", "scale", "none", "quantize pitches to this scale"},
			{"root", "pitch", "60", "root of the scale"},
			{"duration_min", "number", "0.5", "shortest duration in steps"},
			{"duration_max", "number", "2", "longest duration in steps"},
			{"velocity_min", "number", "0.3", "softest velocity"},
			{"velocity_max", "number", "0.9", "loudest velocity"},
		}})
}

func pitchNote(p, vel, start, dur float64) note.Event {
	return note.Event{Pitch: p, Velocity: vel, Start: start, Duration: dur}
}

// ---- notes

type notesParams struct {
	Notes []struct {
		Pitch    Pitch    `json:"pitch"`
		Start    float64  `json:"start"`
		Duration *float64 `json:"duration"`
		Velocity *float64 `json:"velocity"`
	} `json:"notes"`
	Length float64 `json:"length"`
}

func buildNotes(b *Builder, path string, raw json.RawMessage) (Line, error) {
	var p notesParams
	if err := b.Decode(path, raw, &p); err != nil {
		return Line{}, err
	}
	c := newCheck(path)
	if len(p.Notes) == 0 {
		c.fail("notes", "needs at least one note")
	}
	if len(p.Notes) > maxNodeNotes {
		c.fail("notes", "at most %d notes", maxNodeNotes)
	}
	c.posOrZero("length", p.Length, maxSteps)
	line := Line{Length: p.Length}
	for i, n := range p.Notes {
		f := fmt.Sprintf("notes[%d]", i)
		c.pitch(f+".pitch", n.Pitch)
		c.rangeF(f+".start", n.Start, 0, maxSteps)
		d, v := 1.0, 0.8
		if n.Duration != nil {
			d = *n.Duration
		}
		if n.Velocity != nil {
			v = *n.Velocity
		}
		c.pos(f+".duration", d, maxSteps)
		c.unit(f+".velocity", v)
		if c.err != nil {
			return Line{}, c.err
		}
		line.Events = append(line.Events, pitchNote(n.Pitch.F(), v, n.Start, d))
		line.Length = max(line.Length, n.Start+d)
	}
	return line, c.err
}

// ---- melody

type melodyParams struct {
	grid
	Scale    *ScaleSpec `json:"scale"`
	Root     Pitch      `json:"root"`
	Start    int        `json:"start"`
	Low      int        `json:"low"`
	High     int        `json:"high"`
	MaxStep  int        `json:"max_step"`
	LeapProb float64    `json:"leap_prob"`
	MaxLeap  int        `json:"max_leap"`
	RestProb float64    `json:"rest_prob"`
}

func buildMelody(b *Builder, path string, raw json.RawMessage) (Line, error) {
	p := melodyParams{grid: defaultGrid(), Root: P(60), Low: -7, High: 7, MaxStep: 2, LeapProb: 0.15, MaxLeap: 4}
	if err := b.Decode(path, raw, &p); err != nil {
		return Line{}, err
	}
	c := newCheck(path)
	c.scale("scale", p.Scale)
	n := p.grid.check(c, 0)
	c.pitch("root", p.Root)
	c.rangeI("low", p.Low, -64, 64)
	c.rangeI("high", p.High, -64, 64)
	if c.err == nil && p.Low > p.High {
		c.fail("low", "must be <= high (%d), got %d", p.High, p.Low)
	}
	if c.err == nil && (p.Start < p.Low || p.Start > p.High) {
		c.fail("start", "must be within low..high (%d..%d), got %d", p.Low, p.High, p.Start)
	}
	c.rangeI("max_step", p.MaxStep, 1, 8)
	c.rangeI("max_leap", p.MaxLeap, 1, 16)
	c.unit("leap_prob", p.LeapProb)
	c.unit("rest_prob", p.RestProb)
	if c.err != nil {
		return Line{}, c.err
	}
	sc := p.Scale.Scale(p.Root.F(), scale.Major)
	walk := scale.NewWalk(sc, p.Start, p.Low, p.High, p.MaxStep, p.LeapProb, p.MaxLeap)
	line := Line{Length: p.Length}
	for i := range n {
		if p.RestProb > 0 && genny.Float64() < p.RestProb {
			continue
		}
		line.Events = append(line.Events, pitchNote(walk.Generate(), p.Velocity, float64(i)*p.Step, p.NoteLength))
	}
	return line, nil
}

// ---- harmony: chordmelody and progression

type harmonyParams struct {
	Scale     *ScaleSpec `json:"scale"`
	Root      Pitch      `json:"root"`
	Bars      int        `json:"bars"`
	BarSteps  int        `json:"bar_steps"`
	Degrees   []int      `json:"degrees"`
	Table     string     `json:"table"`
	Size      int        `json:"size"`
	Velocity  float64    `json:"velocity"`
	BeatSteps int        `json:"beat_steps"`
	Low       Pitch      `json:"low"`
	High      Pitch      `json:"high"`
	MaxLeap   int        `json:"max_leap"`
	Style     string     `json:"style"`
	Step      float64    `json:"step"`
	Octave    int        `json:"octave"`
}

func harmonyDocs(extra ...Param) []Param {
	return append([]Param{
		{"bars", "int", "required", "number of bars (one chord per bar)"},
		{"bar_steps", "int", "16", "steps per bar"},
		{"scale", "scale", "major", "scale the chords are built on"},
		{"root", "pitch", "60", "root of the scale (degree 1)"},
		{"degrees", "array", "markov", "chord degrees 1-based (1 is I, 5 is V), repeated over the bars; omitted: drawn from table"},
		{"table", "string", "pop", "Markov table for the degrees: pop, classical, jazz or blues"},
		{"size", "int", "3", "notes per chord (3 triad, 4 seventh)"},
		{"velocity", "number", "0.7", "velocity in [0, 1]"},
	}, extra...)
}

var progressionTables = map[string]progression.Table{
	"pop": progression.Pop, "classical": progression.Classical, "jazz": progression.Jazz, "blues": progression.Blues,
}

func (h *harmonyParams) check(c *check, kind string) {
	c.rangeI("bars", h.Bars, 1, 2000)
	c.rangeI("bar_steps", h.BarSteps, 1, maxStepsPerBar*4)
	c.pitch("root", h.Root)
	c.scale("scale", h.Scale)
	c.rangeI("size", h.Size, 2, 7)
	c.unit("velocity", h.Velocity)
	for i, d := range h.Degrees {
		if d < 1 || d > 14 {
			c.fail(fmt.Sprintf("degrees[%d]", i), "must be in 1..14, got %d", d)
			break
		}
	}
	if len(h.Degrees) == 0 {
		if _, ok := progressionTables[h.Table]; !ok {
			names := []string{"pop", "classical", "jazz", "blues"}
			c.fail("table", "unknown table %q (known: %s)", h.Table, strings.Join(names, ", "))
		}
	}
	if c.err == nil && h.Bars*h.BarSteps > maxSteps {
		c.fail("bars", "bars*bar_steps is %d steps, at most %d are allowed", h.Bars*h.BarSteps, maxSteps)
	}
}

// chords draws the chord of every bar
func (h *harmonyParams) chords(sc *scale.Scale) []chord.Chord {
	deg := make([]int, 0, h.Bars)
	if len(h.Degrees) > 0 {
		for i := range h.Bars {
			deg = append(deg, h.Degrees[i%len(h.Degrees)]-1)
		}
	} else {
		g := progression.New(progressionTables[h.Table], 0)
		for len(deg) < h.Bars && !g.Done() {
			deg = append(deg, g.Generate())
		}
		for i := 0; len(deg) < h.Bars; i++ { // a table that ends: repeat what there is
			deg = append(deg, deg[i%len(deg)])
		}
	}
	out := make([]chord.Chord, len(deg))
	for i, d := range deg {
		out[i] = chord.Diatonic(sc, d, h.Size)
	}
	return out
}

func buildChordMelody(b *Builder, path string, raw json.RawMessage) (Line, error) {
	p := harmonyParams{Root: P(60), BarSteps: 16, Table: "pop", Size: 3, Velocity: 0.7, BeatSteps: 4, Low: P(60), High: P(84), MaxLeap: 3}
	if err := b.Decode(path, raw, &p); err != nil {
		return Line{}, err
	}
	c := newCheck(path)
	p.check(c, "chordmelody")
	c.rangeI("beat_steps", p.BeatSteps, 1, p.BarSteps)
	c.pitch("low", p.Low)
	c.pitch("high", p.High)
	c.rangeI("max_leap", p.MaxLeap, 1, 16)
	if c.err == nil && p.BarSteps%p.BeatSteps != 0 {
		c.fail("beat_steps", "must divide bar_steps (%d), got %d", p.BarSteps, p.BeatSteps)
	}
	if c.err == nil && !(p.High.F()-p.Low.F() >= 12) {
		c.fail("high", "must be at least an octave above low (%v), got %v", p.Low.F(), p.High.F())
	}
	if c.err != nil {
		return Line{}, c.err
	}
	sc := p.Scale.Scale(p.Root.F(), scale.Major)
	events := melody.ChordTonesList(p.chords(sc), p.Bars, melody.Config{
		Scale: sc, BarSteps: p.BarSteps, BeatSteps: p.BeatSteps, MaxLeap: p.MaxLeap, Low: p.Low.F(), High: p.High.F(),
	})
	for i := range events {
		events[i].Velocity *= p.Velocity / 0.7 // melody weights velocity by metric position
	}
	return Line{Events: events, Length: float64(p.Bars * p.BarSteps)}, nil
}

func buildProgression(b *Builder, path string, raw json.RawMessage) (Line, error) {
	p := harmonyParams{Root: P(60), BarSteps: 16, Table: "pop", Size: 3, Velocity: 0.7, Style: "block", Step: 2}
	if err := b.Decode(path, raw, &p); err != nil {
		return Line{}, err
	}
	c := newCheck(path)
	p.check(c, "progression")
	switch p.Style {
	case "block", "root", "arpeggio":
	default:
		c.fail("style", "unknown style %q (known: block, root, arpeggio)", p.Style)
	}
	c.pos("step", p.Step, float64(p.BarSteps))
	c.rangeI("octave", p.Octave, -4, 4)
	if c.err != nil {
		return Line{}, c.err
	}
	sc := p.Scale.Scale(p.Root.F(), scale.Major)
	line := Line{Length: float64(p.Bars * p.BarSteps)}
	shift := float64(12 * p.Octave)
	bar := float64(p.BarSteps)
	for i, ch := range p.chords(sc) {
		start := float64(i) * bar
		switch p.Style {
		case "block":
			for _, pitch := range ch {
				line.Events = append(line.Events, pitchNote(pitch+shift, p.Velocity, start, bar))
			}
		case "root":
			line.Events = append(line.Events, pitchNote(ch[0]+shift, p.Velocity, start, bar))
		case "arpeggio":
			k := 0
			for t := 0.0; t < bar-1e-9; t += p.Step {
				line.Events = append(line.Events, pitchNote(ch[k%len(ch)]+shift, p.Velocity, start+t, p.Step))
				k++
			}
		}
	}
	if len(line.Events) > maxNodeNotes {
		return Line{}, Fail(path+".params.bars", "makes %d notes, at most %d are allowed", len(line.Events), maxNodeNotes)
	}
	return line, nil
}

// ---- euclid

type euclidParams struct {
	grid
	Steps    int   `json:"steps"`
	Pulses   int   `json:"pulses"`
	Rotation int   `json:"rotation"`
	Pitch    Pitch `json:"pitch"`
}

func buildEuclid(b *Builder, path string, raw json.RawMessage) (Line, error) {
	p := euclidParams{grid: defaultGrid(), Pitch: P(36)}
	if err := b.Decode(path, raw, &p); err != nil {
		return Line{}, err
	}
	c := newCheck(path)
	c.rangeI("steps", p.Steps, 1, 256)
	if c.err == nil {
		c.rangeI("pulses", p.Pulses, 0, p.Steps)
	}
	c.rangeI("rotation", p.Rotation, -4096, 4096)
	c.pitch("pitch", p.Pitch)
	if c.err != nil {
		return Line{}, c.err
	}
	n := p.grid.check(c, float64(p.Steps)*p.Step)
	if c.err != nil {
		return Line{}, c.err
	}
	pattern := euclid.Pattern(p.Steps, p.Pulses, p.Rotation)
	line := Line{Length: p.Length}
	for i := range n {
		if pattern[i%len(pattern)] {
			line.Events = append(line.Events, pitchNote(p.Pitch.F(), p.Velocity, float64(i)*p.Step, p.NoteLength))
		}
	}
	return line, nil
}

// ---- sieve

type sieveParams struct {
	grid
	Expr  string     `json:"expr"`
	Lo    int        `json:"lo"`
	Hi    int        `json:"hi"`
	Mode  string     `json:"mode"`
	Pitch Pitch      `json:"pitch"`
	Root  Pitch      `json:"root"`
	Scale *ScaleSpec `json:"scale"`
}

func buildSieve(b *Builder, path string, raw json.RawMessage) (Line, error) {
	p := sieveParams{grid: defaultGrid(), Hi: 15, Mode: "gate", Pitch: P(60), Root: P(60)}
	if err := b.Decode(path, raw, &p); err != nil {
		return Line{}, err
	}
	c := newCheck(path)
	c.scale("scale", p.Scale)
	n := p.grid.check(c, 0)
	c.rangeI("lo", p.Lo, -4096, 4096)
	c.rangeI("hi", p.Hi, -4096, 4096)
	if c.err == nil && p.Hi < p.Lo {
		c.fail("hi", "must be >= lo (%d), got %d", p.Lo, p.Hi)
	}
	if c.err == nil && p.Mode != "gate" && p.Mode != "pitches" {
		c.fail("mode", "unknown mode %q (known: gate, pitches)", p.Mode)
	}
	if c.err == nil && p.Mode == "gate" && p.Lo < 0 {
		c.fail("lo", "must be >= 0 in gate mode (steps count from 0), got %d", p.Lo)
	}
	c.pitch("pitch", p.Pitch)
	c.pitch("root", p.Root)
	if c.err != nil {
		return Line{}, c.err
	}
	if strings.TrimSpace(p.Expr) == "" {
		return Line{}, Fail(path+".params.expr", "is required")
	}
	s, err := ParseSieve(p.Expr, p.Lo, p.Hi)
	if err != nil {
		return Line{}, Fail(path+".params.expr", "%v", err)
	}
	if s.Len() == 0 {
		return Line{}, Fail(path+".params.expr", "selects no values in the range %d..%d", p.Lo, p.Hi)
	}
	line := Line{Length: p.Length}
	if p.Mode == "gate" {
		gate := sieve.NewGate(s, p.Hi+1)
		for i := range n {
			if gate.Generate() {
				line.Events = append(line.Events, pitchNote(p.Pitch.F(), p.Velocity, float64(i)*p.Step, p.NoteLength))
			}
		}
		return line, nil
	}
	gen := sieve.NewGen(s, true)
	sc := scale.New(p.Root.F(), scale.Chromatic)
	if p.Scale != nil {
		sc = p.Scale.Scale(p.Root.F(), scale.Major)
	}
	for i := range n {
		line.Events = append(line.Events, pitchNote(sc.Degree(gen.Generate()), p.Velocity, float64(i)*p.Step, p.NoteLength))
	}
	return line, nil
}

// ---- row

type formSpec struct {
	Op        string  `json:"op"`
	Transpose float64 `json:"transpose"`
}

type rowParams struct {
	grid
	Row   []float64  `json:"row"`
	Size  int        `json:"size"`
	Forms []formSpec `json:"forms"`
	Root  Pitch      `json:"root"`
}

var rowOps = map[string]row.Op{
	"prime": row.Prime, "retrograde": row.Retro, "inversion": row.Invert, "retrograde_inversion": row.RetroInvert,
}

func buildRow(b *Builder, path string, raw json.RawMessage) (Line, error) {
	p := rowParams{grid: defaultGrid(), Size: 12, Root: P(60)}
	if err := b.Decode(path, raw, &p); err != nil {
		return Line{}, err
	}
	c := newCheck(path)
	c.rangeI("size", p.Size, 1, 128)
	if len(p.Row) > 128 {
		c.fail("row", "at most 128 notes, got %d", len(p.Row))
	}
	for i, v := range p.Row {
		if !finite(v) || math.Abs(v) > 127 {
			c.fail(fmt.Sprintf("row[%d]", i), "must be a number in -127..127, got %v", v)
			break
		}
	}
	if len(p.Forms) == 0 {
		p.Forms = []formSpec{{Op: "prime"}}
	}
	if len(p.Forms) > 64 {
		c.fail("forms", "at most 64 forms, got %d", len(p.Forms))
	}
	forms := make([]row.Form, 0, len(p.Forms))
	for i, f := range p.Forms {
		op, ok := rowOps[f.Op]
		if !ok && c.err == nil {
			c.fail(fmt.Sprintf("forms[%d].op", i), "unknown op %q (known: prime, retrograde, inversion, retrograde_inversion)", f.Op)
		}
		c.finite(fmt.Sprintf("forms[%d].transpose", i), f.Transpose)
		forms = append(forms, row.Form{Op: op, Transpose: f.Transpose})
	}
	c.pitch("root", p.Root)
	if c.err != nil {
		return Line{}, c.err
	}
	r := p.Row
	if len(r) == 0 {
		r = row.Random(p.Size)
	}
	// without a length the line is the forms played once
	n := p.grid.check(c, float64(len(r)*len(forms))*p.Step)
	if c.err != nil {
		return Line{}, c.err
	}
	seq := row.NewSerial(r, forms...)
	line := Line{Length: p.Length}
	for i := range n {
		line.Events = append(line.Events, pitchNote(p.Root.F()+seq.Generate(), p.Velocity, float64(i)*p.Step, p.NoteLength))
	}
	return line, nil
}

// ---- numseq

type numseqParams struct {
	grid
	Seq   string     `json:"seq"`
	Mod   int        `json:"mod"`
	Start int        `json:"start"`
	Scale *ScaleSpec `json:"scale"`
	Root  Pitch      `json:"root"`
}

func buildNumseq(b *Builder, path string, raw json.RawMessage) (Line, error) {
	p := numseqParams{grid: defaultGrid(), Mod: 7, Start: 7, Root: P(60)}
	if err := b.Decode(path, raw, &p); err != nil {
		return Line{}, err
	}
	c := newCheck(path)
	c.scale("scale", p.Scale)
	n := p.grid.check(c, 0)
	c.rangeI("mod", p.Mod, 2, 64)
	c.rangeI("start", p.Start, 1, 1<<30)
	c.pitch("root", p.Root)
	var gen genny.Generator[int]
	switch p.Seq {
	case "fibonacci", "collatz", "thue_morse", "primes":
	case "":
		c.fail("seq", "is required (fibonacci, collatz, thue_morse or primes)")
	default:
		c.fail("seq", "unknown sequence %q (known: fibonacci, collatz, thue_morse, primes)", p.Seq)
	}
	if c.err != nil {
		return Line{}, c.err
	}
	switch p.Seq {
	case "fibonacci":
		gen = numseq.NewFibonacci(p.Mod)
	case "collatz":
		gen = numseq.NewCollatzLoop(p.Start)
	case "thue_morse":
		gen = numseq.NewThueMorse()
	case "primes":
		gen = numseq.NewPrimes()
	}
	sc := p.Scale.Scale(p.Root.F(), scale.Major)
	line := Line{Length: p.Length}
	for i := range n {
		line.Events = append(line.Events, pitchNote(sc.Degree(gen.Generate()%p.Mod), p.Velocity, float64(i)*p.Step, p.NoteLength))
	}
	return line, nil
}

// ---- pulse

type pulseParams struct {
	grid
	Divisions   []int   `json:"divisions"`
	Level       int     `json:"level"`
	VelocityLow float64 `json:"velocity_low"`
	Pitch       Pitch   `json:"pitch"`
}

func buildPulse(b *Builder, path string, raw json.RawMessage) (Line, error) {
	p := pulseParams{grid: defaultGrid(), VelocityLow: 0.4, Pitch: P(42)}
	if err := b.Decode(path, raw, &p); err != nil {
		return Line{}, err
	}
	c := newCheck(path)
	total := 1
	if len(p.Divisions) == 0 || len(p.Divisions) > 8 {
		c.fail("divisions", "needs 1..8 divisions, got %d", len(p.Divisions))
	}
	for i, d := range p.Divisions {
		if d < 1 || d > 64 {
			c.fail(fmt.Sprintf("divisions[%d]", i), "must be in 1..64, got %d", d)
			break
		}
		if total *= d; total > 4096 {
			c.fail("divisions", "the meter has more than 4096 pulses")
			break
		}
	}
	c.rangeI("level", p.Level, 0, 4096)
	c.unit("velocity_low", p.VelocityLow)
	c.pitch("pitch", p.Pitch)
	if c.err == nil && p.VelocityLow > p.Velocity {
		c.fail("velocity_low", "must be <= velocity (%v), got %v", p.Velocity, p.VelocityLow)
	}
	if c.err != nil {
		return Line{}, c.err
	}
	n := p.grid.check(c, float64(total)*p.Step)
	if c.err != nil {
		return Line{}, c.err
	}
	ind := pulse.Hierarchy(p.Divisions...)
	play := pulse.Thin(ind, p.Level)
	vel := pulse.Velocity(ind, p.VelocityLow, p.Velocity)
	line := Line{Length: p.Length}
	for i := range n {
		if k := i % len(ind); play[k] {
			line.Events = append(line.Events, pitchNote(p.Pitch.F(), vel[k], float64(i)*p.Step, p.NoteLength))
		}
	}
	return line, nil
}

// ---- tendency

type segmentSpec struct {
	N       int      `json:"n"`
	Low     float64  `json:"low"`
	High    float64  `json:"high"`
	EndLow  *float64 `json:"end_low"`
	EndHigh *float64 `json:"end_high"`
}

type tendencyParams struct {
	grid
	Segments []segmentSpec `json:"segments"`
	Loop     bool          `json:"loop"`
	Scale    *ScaleSpec    `json:"scale"`
	Root     Pitch         `json:"root"`
	Beta     []float64     `json:"beta"`
}

func buildTendency(b *Builder, path string, raw json.RawMessage) (Line, error) {
	p := tendencyParams{grid: defaultGrid(), Root: P(60)}
	if err := b.Decode(path, raw, &p); err != nil {
		return Line{}, err
	}
	c := newCheck(path)
	c.scale("scale", p.Scale)
	n := p.grid.check(c, 0)
	if len(p.Segments) == 0 || len(p.Segments) > 1024 {
		c.fail("segments", "needs 1..1024 segments, got %d", len(p.Segments))
	}
	segs := make([]tendency.Segment, 0, len(p.Segments))
	for i, s := range p.Segments {
		f := fmt.Sprintf("segments[%d]", i)
		c.rangeI(f+".n", s.N, 1, maxNodeNotes)
		c.rangeF(f+".low", s.Low, 0, 127)
		c.rangeF(f+".high", s.High, 0, 127)
		el, eh := s.Low, s.High
		if s.EndLow != nil {
			el = *s.EndLow
		}
		if s.EndHigh != nil {
			eh = *s.EndHigh
		}
		c.rangeF(f+".end_low", el, 0, 127)
		c.rangeF(f+".end_high", eh, 0, 127)
		segs = append(segs, tendency.Segment{N: s.N, A1: s.Low, A2: s.High, Z1: el, Z2: eh})
	}
	if len(p.Beta) != 0 && (len(p.Beta) != 2 || !(p.Beta[0] > 0) || !(p.Beta[1] > 0) || !finite(p.Beta[0]) || !finite(p.Beta[1])) {
		c.fail("beta", "must be [a, b] with both positive")
	}
	c.pitch("root", p.Root)
	if c.err != nil {
		return Line{}, c.err
	}
	mask := tendency.NewMask(p.Loop, segs...)
	var gen genny.Generator[float64] = tendency.NewValue(mask)
	if len(p.Beta) == 2 {
		gen = tendency.NewBetaValue(mask, p.Beta[0], p.Beta[1])
	}
	var sc *scale.Scale
	if p.Scale != nil {
		sc = p.Scale.Scale(p.Root.F(), scale.Major)
	}
	line := Line{Length: p.Length}
	for i := range n {
		v := gen.Generate()
		if sc != nil {
			v = sc.Quantize(v)
		}
		line.Events = append(line.Events, pitchNote(v, p.Velocity, float64(i)*p.Step, p.NoteLength))
	}
	return line, nil
}

// ---- choose

type chooseParams struct {
	grid
	Options []struct {
		Pitch  *Pitch   `json:"pitch"`
		Weight *float64 `json:"weight"`
	} `json:"options"`
}

func buildChoose(b *Builder, path string, raw json.RawMessage) (Line, error) {
	p := chooseParams{grid: defaultGrid()}
	if err := b.Decode(path, raw, &p); err != nil {
		return Line{}, err
	}
	c := newCheck(path)
	n := p.grid.check(c, 0)
	if len(p.Options) == 0 || len(p.Options) > 1024 {
		c.fail("options", "needs 1..1024 options, got %d", len(p.Options))
	}
	opts := make([]choose.Option[int], 0, len(p.Options))
	sum := 0.0
	for i, o := range p.Options {
		f := fmt.Sprintf("options[%d]", i)
		w := 1.0
		if o.Weight != nil {
			w = *o.Weight
		}
		c.rangeF(f+".weight", w, 0, 1e9)
		if o.Pitch != nil {
			c.pitch(f+".pitch", *o.Pitch)
		}
		sum += w
		opts = append(opts, choose.Option[int]{Gen: sequence.NewLoop(i), Weight: w})
	}
	if c.err == nil && !(sum > 0) {
		c.fail("options", "the weights add up to 0: at least one must be > 0")
	}
	if c.err != nil {
		return Line{}, c.err
	}
	pick := choose.New(opts...)
	line := Line{Length: p.Length}
	for i := range n {
		if pitch := p.Options[pick.Generate()].Pitch; pitch != nil {
			line.Events = append(line.Events, pitchNote(pitch.F(), p.Velocity, float64(i)*p.Step, p.NoteLength))
		}
	}
	return line, nil
}

// ---- cloud

type cloudParams struct {
	Length      float64    `json:"length"`
	Density     float64    `json:"density"`
	DensityEnd  *float64   `json:"density_end"`
	Low         Pitch      `json:"low"`
	High        Pitch      `json:"high"`
	Scale       *ScaleSpec `json:"scale"`
	Root        Pitch      `json:"root"`
	DurationMin float64    `json:"duration_min"`
	DurationMax float64    `json:"duration_max"`
	VelocityMin float64    `json:"velocity_min"`
	VelocityMax float64    `json:"velocity_max"`
}

func buildCloud(b *Builder, path string, raw json.RawMessage) (Line, error) {
	p := cloudParams{Low: P(36), High: P(84), Root: P(60), DurationMin: 0.5, DurationMax: 2, VelocityMin: 0.3, VelocityMax: 0.9}
	if err := b.Decode(path, raw, &p); err != nil {
		return Line{}, err
	}
	c := newCheck(path)
	c.scale("scale", p.Scale)
	c.pos("length", p.Length, maxSteps)
	c.rangeF("density", p.Density, 0, 100)
	end := p.Density
	if p.DensityEnd != nil {
		end = *p.DensityEnd
		c.rangeF("density_end", end, 0, 100)
	}
	c.pitch("low", p.Low)
	c.pitch("high", p.High)
	c.pitch("root", p.Root)
	c.pos("duration_min", p.DurationMin, maxSteps)
	c.pos("duration_max", p.DurationMax, maxSteps)
	c.unit("velocity_min", p.VelocityMin)
	c.unit("velocity_max", p.VelocityMax)
	if c.err == nil && p.Low.F() > p.High.F() {
		c.fail("low", "must be <= high (%v), got %v", p.High.F(), p.Low.F())
	}
	if c.err == nil && p.DurationMin > p.DurationMax {
		c.fail("duration_min", "must be <= duration_max (%v), got %v", p.DurationMax, p.DurationMin)
	}
	if c.err == nil && p.VelocityMin > p.VelocityMax {
		c.fail("velocity_min", "must be <= velocity_max (%v), got %v", p.VelocityMax, p.VelocityMin)
	}
	if c.err == nil && max(p.Density, end)*p.Length > maxNodeNotes {
		c.fail("density", "density*length expects more than %d events", maxNodeNotes)
	}
	if c.err != nil {
		return Line{}, c.err
	}
	steps := int(math.Ceil(p.Length))
	cfg := stochastic.Config{
		Density: func(step int) float64 {
			if steps <= 1 {
				return p.Density
			}
			return p.Density + (end-p.Density)*float64(step)/float64(steps-1)
		},
		Pitch:    stochastic.Uniform(p.Low.F(), p.High.F()),
		Velocity: stochastic.Uniform(p.VelocityMin, p.VelocityMax),
		Duration: stochastic.Uniform(p.DurationMin, p.DurationMax),
	}
	if p.Scale != nil {
		cfg.Scale = p.Scale.Scale(p.Root.F(), scale.Major)
	}
	events := stochastic.NewCloud(cfg).Render(steps)
	if len(events) > 0 && p.Length < float64(steps) { // the last step may be partial
		kept := events[:0]
		for _, e := range events {
			if e.Start < p.Length {
				kept = append(kept, e)
			}
		}
		events = kept
	}
	return Line{Events: events, Length: p.Length}, nil
}
