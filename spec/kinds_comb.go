package spec

import (
	"encoding/json"
	"fmt"
	"math"
	"slices"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/groove"
	"github.com/almerlucke/genny/line"
	"github.com/almerlucke/genny/note"
	"github.com/almerlucke/genny/scale"
)

func init() {
	itemDoc := Param{"item", "node", "required", "the node to work on"}
	itemsDoc := Param{"items", "array of nodes", "required", "the nodes"}
	Register(Kind{Name: "concat", Category: "combinator", Build: buildConcat,
		Doc: "play the items one after the other (each starts where the one before ends)", Params: []Param{itemsDoc}})
	Register(Kind{Name: "layer", Category: "combinator", Build: buildLayer,
		Doc: "play the items together (merge is the same)", Params: []Param{itemsDoc}})
	Register(Kind{Name: "merge", Category: "combinator", Build: buildLayer,
		Doc: "same as layer", Params: []Param{itemsDoc}})
	Register(Kind{Name: "repeat", Category: "combinator", Build: buildRepeat,
		Doc:    "repeat the item n times (package line Loop); the item is built once, so random items repeat exactly",
		Params: []Param{{"n", "int", "required", "number of times, 1..1000"}, itemDoc}})
	Register(Kind{Name: "transpose", Category: "combinator", Build: buildTranspose,
		Doc: "move all pitches by semitones", Params: []Param{{"semitones", "number", "required", "may be negative or fractional"}, itemDoc}})
	Register(Kind{Name: "invert", Category: "combinator", Build: buildInvert,
		Doc: "mirror the pitches around an axis, chromatically or within a scale",
		Params: []Param{
			{"axis", "pitch", "first note", "the axis of the mirror"},
			{"scale", "scale", "none", "with a scale the mirror is diatonic"},
			{"root", "pitch", "0", "root of the scale"}, itemDoc,
		}})
	Register(Kind{Name: "retrograde", Category: "combinator", Build: buildRetrograde,
		Doc: "play the item backwards within its length", Params: []Param{itemDoc}})
	Register(Kind{Name: "stretch", Category: "combinator", Build: buildStretch,
		Doc: "multiply times and durations: 2 is twice as slow", Params: []Param{{"factor", "number", "required", "in (0, 64]"}, itemDoc}})
	Register(Kind{Name: "shift", Category: "combinator", Build: buildShift,
		Doc: "move the item later by steps (rest before it)", Params: []Param{{"steps", "number", "required", ">= 0"}, itemDoc}})
	Register(Kind{Name: "slice", Category: "combinator", Build: buildSlice,
		Doc:    "keep the notes that start in [from, to) and move them to start at 0",
		Params: []Param{{"from", "number", "0", "first step"}, {"to", "number", "required", "end step"}, itemDoc}})
	Register(Kind{Name: "quantize", Category: "combinator", Build: buildQuantize,
		Doc:    "move every pitch to the nearest pitch of a scale",
		Params: []Param{{"scale", "scale", "required", "the scale"}, {"root", "pitch", "0", "root of the scale"}, itemDoc}})
	Register(Kind{Name: "canon", Category: "combinator", Build: buildCanon,
		Doc: "the item as a canon: voice i enters i*delay steps later and i*interval semitones higher",
		Params: []Param{
			{"delay", "number", "required", "steps between entries"},
			{"interval", "number", "0", "semitones between voices"},
			{"voices", "int", "2", "number of voices, 1..16"}, itemDoc,
		}})
	Register(Kind{Name: "swing", Category: "combinator", Build: buildSwing,
		Doc: "delay the notes on odd steps (package groove)",
		Params: []Param{
			{"amount", "number", "required", "delay in fractions of a step; 1/3 is a triplet feel"},
			{"humanize", "number", "0", "random timing jitter in steps"}, itemDoc,
		}})
	Register(Kind{Name: "humanize", Category: "combinator", Build: buildHumanize,
		Doc: "random jitter of start time and velocity (package groove)",
		Params: []Param{
			{"timing", "number", "0", "maximum shift of the start in steps"},
			{"velocity", "number", "0", "maximum velocity change"}, itemDoc,
		}})
	Register(Kind{Name: "gain", Category: "combinator", Build: buildGain,
		Doc: "scale the velocities (clamped to [0, 1])", Params: []Param{{"factor", "number", "required", ">= 0"}, itemDoc}})
	Register(Kind{Name: "thin", Category: "combinator", Build: buildThin,
		Doc:    "keep each note with a probability",
		Params: []Param{{"keep", "number", "required", "probability in [0, 1]"}, itemDoc}})
}

// ---- list combinators

type itemsParams struct {
	Items []json.RawMessage `json:"items"`
}

func buildItems(b *Builder, path string, raw json.RawMessage) ([]Line, error) {
	var p itemsParams
	if err := b.Decode(path, raw, &p); err != nil {
		return nil, err
	}
	if len(p.Items) == 0 {
		return nil, Fail(path+".params.items", "needs at least one item")
	}
	if len(p.Items) > 1024 {
		return nil, Fail(path+".params.items", "at most 1024 items, got %d", len(p.Items))
	}
	lines := make([]Line, len(p.Items))
	for i, it := range p.Items {
		l, err := b.Child(fmt.Sprintf("%s.params.items[%d]", path, i), it)
		if err != nil {
			return nil, err
		}
		lines[i] = l
	}
	return lines, nil
}

func buildConcat(b *Builder, path string, raw json.RawMessage) (Line, error) {
	lines, err := buildItems(b, path, raw)
	if err != nil {
		return Line{}, err
	}
	var out Line
	for _, l := range lines {
		out.Events = append(out.Events, line.Shift(l.Events, out.Length)...)
		out.Length += l.Length
	}
	return out, nil
}

func buildLayer(b *Builder, path string, raw json.RawMessage) (Line, error) {
	lines, err := buildItems(b, path, raw)
	if err != nil {
		return Line{}, err
	}
	var out Line
	all := make([][]note.Event, len(lines))
	for i, l := range lines {
		all[i] = l.Events
		out.Length = max(out.Length, l.Length)
	}
	out.Events = line.Merge(all...)
	return out, nil
}

// ---- single item combinators

// unary decodes params into p (which holds the item as `item` through get) and builds the item
func unary[P any](b *Builder, path string, raw json.RawMessage, p *P, item func(*P) json.RawMessage) (Line, error) {
	if err := b.Decode(path, raw, p); err != nil {
		return Line{}, err
	}
	return b.Child(path+".params.item", item(p))
}

type repeatParams struct {
	N    int             `json:"n"`
	Item json.RawMessage `json:"item"`
}

func buildRepeat(b *Builder, path string, raw json.RawMessage) (Line, error) {
	var p repeatParams
	if err := b.Decode(path, raw, &p); err != nil {
		return Line{}, err
	}
	c := newCheck(path)
	c.rangeI("n", p.N, 1, 1000)
	if c.err != nil {
		return Line{}, c.err
	}
	l, err := b.Child(path+".params.item", p.Item)
	if err != nil {
		return Line{}, err
	}
	if !(l.Length > 0) {
		return Line{}, Fail(path+".params.item", "has length 0, cannot be repeated")
	}
	if float64(p.N)*l.Length > maxSteps {
		return Line{}, Fail(path+".params.n", "%d repeats of %v steps is longer than %d steps", p.N, l.Length, maxSteps)
	}
	if len(l.Events)*p.N > maxNodeNotes*5 {
		return Line{}, Fail(path+".params.n", "%d repeats make too many notes", p.N)
	}
	return Line{Events: line.Loop(l.Events, p.N, l.Length), Length: l.Length * float64(p.N)}, nil
}

type transposeParams struct {
	Semitones *float64        `json:"semitones"`
	Item      json.RawMessage `json:"item"`
}

func buildTranspose(b *Builder, path string, raw json.RawMessage) (Line, error) {
	var p transposeParams
	if err := b.Decode(path, raw, &p); err != nil {
		return Line{}, err
	}
	c := newCheck(path)
	if p.Semitones == nil {
		c.fail("semitones", "is required")
	} else {
		c.finite("semitones", *p.Semitones)
	}
	if c.err != nil {
		return Line{}, c.err
	}
	l, err := b.Child(path+".params.item", p.Item)
	if err != nil {
		return Line{}, err
	}
	l.Events = line.Transpose(l.Events, *p.Semitones)
	return l, nil
}

type invertParams struct {
	Axis  *Pitch          `json:"axis"`
	Scale *ScaleSpec      `json:"scale"`
	Root  Pitch           `json:"root"`
	Item  json.RawMessage `json:"item"`
}

func buildInvert(b *Builder, path string, raw json.RawMessage) (Line, error) {
	var p invertParams
	if err := b.Decode(path, raw, &p); err != nil {
		return Line{}, err
	}
	c := newCheck(path)
	c.scale("scale", p.Scale)
	if p.Axis != nil {
		c.pitch("axis", *p.Axis)
	}
	c.pitch("root", p.Root)
	if c.err != nil {
		return Line{}, c.err
	}
	l, err := b.Child(path+".params.item", p.Item)
	if err != nil {
		return Line{}, err
	}
	if len(l.Events) == 0 {
		return l, nil
	}
	axis := l.Events[0].Pitch
	if p.Axis != nil {
		axis = p.Axis.F()
	}
	var sc *scale.Scale
	if p.Scale != nil {
		sc = p.Scale.Scale(p.Root.F(), scale.Major)
	}
	l.Events = line.Invert(l.Events, axis, sc)
	return l, nil
}

type itemParams struct {
	Item json.RawMessage `json:"item"`
}

func buildRetrograde(b *Builder, path string, raw json.RawMessage) (Line, error) {
	var p itemParams
	l, err := unary(b, path, raw, &p, func(p *itemParams) json.RawMessage { return p.Item })
	if err != nil || len(l.Events) == 0 {
		return l, err
	}
	// line.Retrograde mirrors within the span of the notes; mirror within the length of the line instead
	first, last := math.Inf(1), math.Inf(-1)
	for _, e := range l.Events {
		first = min(first, e.Start)
		last = max(last, e.Start+e.Duration)
	}
	l.Events = line.Shift(line.Retrograde(l.Events), l.Length-first-last)
	return l, nil
}

type factorParams struct {
	Factor *float64        `json:"factor"`
	Item   json.RawMessage `json:"item"`
}

func buildStretch(b *Builder, path string, raw json.RawMessage) (Line, error) {
	var p factorParams
	if err := b.Decode(path, raw, &p); err != nil {
		return Line{}, err
	}
	c := newCheck(path)
	if p.Factor == nil {
		c.fail("factor", "is required")
	} else {
		c.pos("factor", *p.Factor, 64)
	}
	if c.err != nil {
		return Line{}, c.err
	}
	l, err := b.Child(path+".params.item", p.Item)
	if err != nil {
		return Line{}, err
	}
	return Line{Events: line.Stretch(l.Events, *p.Factor), Length: l.Length * *p.Factor}, nil
}

func buildGain(b *Builder, path string, raw json.RawMessage) (Line, error) {
	var p factorParams
	if err := b.Decode(path, raw, &p); err != nil {
		return Line{}, err
	}
	c := newCheck(path)
	if p.Factor == nil {
		c.fail("factor", "is required")
	} else {
		c.rangeF("factor", *p.Factor, 0, 100)
	}
	if c.err != nil {
		return Line{}, c.err
	}
	l, err := b.Child(path+".params.item", p.Item)
	if err != nil {
		return Line{}, err
	}
	l.Events = slices.Clone(l.Events)
	for i := range l.Events {
		l.Events[i].Velocity = min(l.Events[i].Velocity**p.Factor, 1)
	}
	return l, nil
}

type shiftParams struct {
	Steps *float64        `json:"steps"`
	Item  json.RawMessage `json:"item"`
}

func buildShift(b *Builder, path string, raw json.RawMessage) (Line, error) {
	var p shiftParams
	if err := b.Decode(path, raw, &p); err != nil {
		return Line{}, err
	}
	c := newCheck(path)
	if p.Steps == nil {
		c.fail("steps", "is required")
	} else {
		c.rangeF("steps", *p.Steps, 0, maxSteps)
	}
	if c.err != nil {
		return Line{}, c.err
	}
	l, err := b.Child(path+".params.item", p.Item)
	if err != nil {
		return Line{}, err
	}
	return Line{Events: line.Shift(l.Events, *p.Steps), Length: l.Length + *p.Steps}, nil
}

type sliceParams struct {
	From float64         `json:"from"`
	To   *float64        `json:"to"`
	Item json.RawMessage `json:"item"`
}

func buildSlice(b *Builder, path string, raw json.RawMessage) (Line, error) {
	var p sliceParams
	if err := b.Decode(path, raw, &p); err != nil {
		return Line{}, err
	}
	c := newCheck(path)
	c.rangeF("from", p.From, 0, maxSteps)
	if p.To == nil {
		c.fail("to", "is required")
	} else if c.err == nil && !(*p.To > p.From) {
		c.fail("to", "must be > from (%v), got %v", p.From, *p.To)
	}
	if c.err != nil {
		return Line{}, c.err
	}
	l, err := b.Child(path+".params.item", p.Item)
	if err != nil {
		return Line{}, err
	}
	return Line{Events: line.Slice(l.Events, p.From, *p.To), Length: *p.To - p.From}, nil
}

type quantizeParams struct {
	Scale *ScaleSpec      `json:"scale"`
	Root  Pitch           `json:"root"`
	Item  json.RawMessage `json:"item"`
}

func buildQuantize(b *Builder, path string, raw json.RawMessage) (Line, error) {
	var p quantizeParams
	if err := b.Decode(path, raw, &p); err != nil {
		return Line{}, err
	}
	c := newCheck(path)
	c.scale("scale", p.Scale)
	if p.Scale == nil {
		c.fail("scale", "is required")
	}
	c.pitch("root", p.Root)
	if c.err != nil {
		return Line{}, c.err
	}
	l, err := b.Child(path+".params.item", p.Item)
	if err != nil {
		return Line{}, err
	}
	l.Events = line.Quantize(l.Events, p.Scale.Scale(p.Root.F(), scale.Major))
	return l, nil
}

type canonParams struct {
	Delay    *float64        `json:"delay"`
	Interval float64         `json:"interval"`
	Voices   int             `json:"voices"`
	Item     json.RawMessage `json:"item"`
}

func buildCanon(b *Builder, path string, raw json.RawMessage) (Line, error) {
	p := canonParams{Voices: 2}
	if err := b.Decode(path, raw, &p); err != nil {
		return Line{}, err
	}
	c := newCheck(path)
	if p.Delay == nil {
		c.fail("delay", "is required")
	} else {
		c.rangeF("delay", *p.Delay, 0, maxSteps)
	}
	c.finite("interval", p.Interval)
	c.rangeI("voices", p.Voices, 1, 16)
	if c.err != nil {
		return Line{}, c.err
	}
	l, err := b.Child(path+".params.item", p.Item)
	if err != nil {
		return Line{}, err
	}
	voices := line.Canon(l.Events, *p.Delay, p.Interval, p.Voices)
	return Line{Events: line.Merge(voices...), Length: l.Length + float64(p.Voices-1)**p.Delay}, nil
}

type swingParams struct {
	Amount   *float64        `json:"amount"`
	Humanize float64         `json:"humanize"`
	Item     json.RawMessage `json:"item"`
}

func buildSwing(b *Builder, path string, raw json.RawMessage) (Line, error) {
	var p swingParams
	if err := b.Decode(path, raw, &p); err != nil {
		return Line{}, err
	}
	c := newCheck(path)
	if p.Amount == nil {
		c.fail("amount", "is required")
	} else {
		c.rangeF("amount", *p.Amount, 0, 1)
	}
	c.rangeF("humanize", p.Humanize, 0, 1)
	if c.err != nil {
		return Line{}, c.err
	}
	l, err := b.Child(path+".params.item", p.Item)
	if err != nil {
		return Line{}, err
	}
	steps := int(math.Ceil(l.Length)) + 1
	sw := groove.NewSwing(*p.Amount).Humanize(p.Humanize)
	offsets := make([]float64, steps)
	for i := range offsets {
		offsets[i] = sw.Generate()
	}
	l.Events = slices.Clone(l.Events)
	for i := range l.Events {
		k := min(int(l.Events[i].Start), steps-1)
		l.Events[i].Start = max(l.Events[i].Start+offsets[k], 0)
	}
	return l, nil
}

type humanizeParams struct {
	Timing   float64         `json:"timing"`
	Velocity float64         `json:"velocity"`
	Item     json.RawMessage `json:"item"`
}

func buildHumanize(b *Builder, path string, raw json.RawMessage) (Line, error) {
	var p humanizeParams
	if err := b.Decode(path, raw, &p); err != nil {
		return Line{}, err
	}
	c := newCheck(path)
	c.rangeF("timing", p.Timing, 0, 16)
	c.rangeF("velocity", p.Velocity, 0, 1)
	if c.err != nil {
		return Line{}, c.err
	}
	l, err := b.Child(path+".params.item", p.Item)
	if err != nil {
		return Line{}, err
	}
	t, v := groove.NewHumanize(p.Timing), groove.NewHumanize(p.Velocity)
	l.Events = slices.Clone(l.Events)
	for i := range l.Events {
		l.Events[i].Start = max(l.Events[i].Start+t.Generate(), 0)
		l.Events[i].Velocity = min(max(l.Events[i].Velocity+v.Generate(), 0), 1)
	}
	return l, nil
}

type thinParams struct {
	Keep *float64        `json:"keep"`
	Item json.RawMessage `json:"item"`
}

func buildThin(b *Builder, path string, raw json.RawMessage) (Line, error) {
	var p thinParams
	if err := b.Decode(path, raw, &p); err != nil {
		return Line{}, err
	}
	c := newCheck(path)
	if p.Keep == nil {
		c.fail("keep", "is required")
	} else {
		c.unit("keep", *p.Keep)
	}
	if c.err != nil {
		return Line{}, c.err
	}
	l, err := b.Child(path+".params.item", p.Item)
	if err != nil {
		return Line{}, err
	}
	var gate genny.Generator[bool] = keepGate(*p.Keep)
	l.Events = line.Thin(l.Events, gate)
	return l, nil
}

// keepGate is true with a probability, from the seeded stream
type keepGate float64

func (k keepGate) Generate() bool   { return genny.Float64() < float64(k) }
func (k keepGate) Continuous() bool { return true }
func (k keepGate) Done() bool       { return false }
func (k keepGate) Reset()           {}
