package spec

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/almerlucke/genny/midi"
)

// one valid node per kind; TestEveryKindHasACase makes sure that new kinds get one
var validNodes = map[string]string{
	"notes":       `{"kind":"notes","params":{"notes":[{"pitch":"C4","start":0},{"pitch":62,"start":2,"duration":3,"velocity":0.5}],"length":8}}`,
	"melody":      `{"kind":"melody","params":{"scale":"minor","root":"D3","length":32}}`,
	"chordmelody": `{"kind":"chordmelody","params":{"bars":4,"degrees":[1,4,5,1]}}`,
	"progression": `{"kind":"progression","params":{"bars":4,"style":"arpeggio","table":"jazz"}}`,
	"euclid":      `{"kind":"euclid","params":{"pulses":5,"steps":8,"pitch":36}}`,
	"sieve":       `{"kind":"sieve","params":{"expr":"(3@0|5@1)&!2@0","hi":31,"length":32}}`,
	"row":         `{"kind":"row","params":{"forms":[{"op":"prime"},{"op":"retrograde_inversion","transpose":3}]}}`,
	"numseq":      `{"kind":"numseq","params":{"seq":"collatz","length":16,"mod":7}}`,
	"pulse":       `{"kind":"pulse","params":{"divisions":[3,2,2],"level":3}}`,
	"tendency":    `{"kind":"tendency","params":{"length":16,"segments":[{"n":8,"low":60,"high":64,"end_low":50,"end_high":80}],"scale":"major"}}`,
	"choose":      `{"kind":"choose","params":{"length":16,"options":[{"pitch":60},{"pitch":"E4","weight":2},{"pitch":null}]}}`,
	"cloud":       `{"kind":"cloud","params":{"length":16,"density":1.5,"density_end":0.5}}`,
	"concat":      `{"kind":"concat","params":{"items":[` + leaf + `,` + leaf + `]}}`,
	"layer":       `{"kind":"layer","params":{"items":[` + leaf + `,` + leaf + `]}}`,
	"merge":       `{"kind":"merge","params":{"items":[` + leaf + `]}}`,
	"repeat":      wrap("repeat", `"n":3`),
	"transpose":   wrap("transpose", `"semitones":7`),
	"invert":      wrap("invert", `"axis":"C4","scale":"major"`),
	"retrograde":  wrap("retrograde", ``),
	"stretch":     wrap("stretch", `"factor":2`),
	"shift":       wrap("shift", `"steps":4`),
	"slice":       wrap("slice", `"from":1,"to":6`),
	"quantize":    wrap("quantize", `"scale":[0,2,4,7,9],"root":0`),
	"canon":       wrap("canon", `"delay":4,"interval":7,"voices":3`),
	"swing":       wrap("swing", `"amount":0.3,"humanize":0.05`),
	"humanize":    wrap("humanize", `"timing":0.1,"velocity":0.1`),
	"gain":        wrap("gain", `"factor":0.5`),
	"thin":        wrap("thin", `"keep":0.5`),
}

const leaf = `{"kind":"melody","params":{"length":8,"root":60}}`

func wrap(kind, params string) string {
	if params != "" {
		params += ","
	}
	return `{"kind":"` + kind + `","params":{` + params + `"item":` + leaf + `}}`
}

func piece(source string) string {
	return `{"seed":5,"tracks":[{"name":"t","channel":0,"program":1,"source":` + source + `}]}`
}

func render(t *testing.T, doc string, seed *uint64) (Result, error) {
	t.Helper()
	s, err := ParseBytes([]byte(doc))
	if err != nil {
		return Result{}, err
	}
	return s.Render(seed)
}

func TestEveryKindHasACase(t *testing.T) {
	for _, k := range Kinds() {
		if _, ok := validNodes[k.Name]; !ok {
			t.Errorf("kind %q has no valid case in validNodes", k.Name)
		}
		if k.Doc == "" || k.Category == "" {
			t.Errorf("kind %q lacks documentation", k.Name)
		}
	}
	for name := range validNodes {
		if _, ok := Lookup(name); !ok {
			t.Errorf("case for unknown kind %q", name)
		}
	}
}

func TestValidKinds(t *testing.T) {
	for name, node := range validNodes {
		t.Run(name, func(t *testing.T) {
			res, err := render(t, piece(node), nil)
			if err != nil {
				t.Fatalf("%v\n%s", err, node)
			}
			if len(res.File.Tracks) != 1 || len(res.File.Tracks[0].Events) == 0 {
				t.Fatalf("expected one track with notes, got %+v", res)
			}
			var buf bytes.Buffer
			if err := midi.Write(&buf, res.File); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestInvalid(t *testing.T) {
	cases := []struct {
		name, doc, want string
	}{
		{"not json", `{"tracks": [`, "line 1"},
		{"syntax position", "{\n  \"tempo\": 1,,\n}", "line 2"},
		{"not an object", `[]`, "expected an object"},
		{"unknown top field", `{"tracks":[],"bogus":1}`, `bogus: unknown field`},
		{"no tracks", `{"tracks":[]}`, "tracks: must have at least one track"},
		{"bad tempo", `{"tempo":-5,"tracks":[{}]}`, "tempo: must be in"},
		{"bad spb", `{"steps_per_beat":100,"tracks":[{}]}`, "steps_per_beat"},
		{"bad time signature", `{"time_signature":{"num":3,"den":3},"tracks":[{}]}`, "time_signature.den"},
		{"seed negative", `{"seed":-1,"tracks":[{}]}`, "seed"},
		{"unknown track field", `{"tracks":[{"nme":"x"}]}`, "tracks[0].nme: unknown field"},
		{"bad channel", `{"tracks":[{"channel":16,"source":` + leaf + `}]}`, "tracks[0].channel: must be in 0..15"},
		{"bad program", `{"tracks":[{"program":128,"source":` + leaf + `}]}`, "tracks[0].program: must be in 0..127"},
		{"missing source", `{"tracks":[{"name":"x"}]}`, "tracks[0].source: is required"},
		{"channel type", `{"tracks":[{"channel":"a","source":` + leaf + `}]}`, "tracks[0].channel: expected a number, got string"},
		{"missing kind", piece(`{"params":{}}`), "tracks[0].source.kind: is required"},
		{"unknown kind", piece(`{"kind":"melodi"}`), `tracks[0].source.kind: unknown kind "melodi" (did you mean "melody"?)`},
		{"unknown node field", piece(`{"kind":"melody","length":4}`), "tracks[0].source.length: unknown field"},
		{"unknown param", piece(`{"kind":"melody","params":{"length":4,"lenght":3}}`), "tracks[0].source.params.lenght: unknown field"},
		{"length zero", piece(`{"kind":"melody","params":{"length":0}}`), "tracks[0].source.params.length: must be > 0"},
		{"length missing", piece(`{"kind":"melody"}`), "tracks[0].source.params.length: must be > 0"},
		{"length huge", piece(`{"kind":"melody","params":{"length":1e9}}`), "params.length: must be <="},
		{"too many notes", piece(`{"kind":"melody","params":{"length":100000,"step":0.0001}}`), "params.step"},
		{"param type", piece(`{"kind":"melody","params":{"length":"long"}}`), "params.length: expected a number, got string"},
		{"bad scale name", piece(`{"kind":"melody","params":{"length":4,"scale":"klingon"}}`), `params.scale: unknown scale "klingon"`},
		{"bad scale steps", piece(`{"kind":"melody","params":{"length":4,"scale":[0,4,2]}}`), "params.scale: scale steps must ascend"},
		{"bad scale type", piece(`{"kind":"melody","params":{"length":4,"scale":5}}`), "params.scale: expected a scale name"},
		{"bad note name", piece(`{"kind":"melody","params":{"length":4,"root":"H4"}}`), `params.root: invalid note name "H4"`},
		{"bad octave", piece(`{"kind":"melody","params":{"length":4,"root":"C"}}`), "params.root: invalid note name"},
		{"pitch type", piece(`{"kind":"melody","params":{"length":4,"root":[1]}}`), "params.root: expected a number or a note name, got array"},
		{"pitch range", piece(`{"kind":"melody","params":{"length":4,"root":200}}`), "params.root: pitch 200 is outside"},
		{"start outside", piece(`{"kind":"melody","params":{"length":4,"start":99}}`), "params.start"},
		{"low above high", piece(`{"kind":"melody","params":{"length":4,"low":5,"high":2}}`), "params.low: must be <= high"},
		{"velocity", piece(`{"kind":"melody","params":{"length":4,"velocity":2}}`), "params.velocity: must be in [0, 1]"},
		{"euclid pulses", piece(`{"kind":"euclid","params":{"steps":8,"pulses":9}}`), "params.pulses: must be in 0..8"},
		{"euclid steps", piece(`{"kind":"euclid","params":{"pulses":1}}`), "params.steps: must be in 1..256"},
		{"sieve empty", piece(`{"kind":"sieve","params":{"length":8}}`), "params.expr: is required"},
		{"sieve syntax", piece(`{"kind":"sieve","params":{"length":8,"expr":"3@0 |"}}`), "params.expr: position"},
		{"sieve mod", piece(`{"kind":"sieve","params":{"length":8,"expr":"0@1"}}`), "modulus must be at least 1"},
		{"sieve paren", piece(`{"kind":"sieve","params":{"length":8,"expr":"(3@0"}}`), "expected ')'"},
		{"sieve none", piece(`{"kind":"sieve","params":{"length":8,"expr":"3@0&3@1"}}`), "selects no values"},
		{"sieve mode", piece(`{"kind":"sieve","params":{"length":8,"expr":"3@0","mode":"x"}}`), "params.mode: unknown mode"},
		{"row op", piece(`{"kind":"row","params":{"forms":[{"op":"prime"},{"op":"flip"}]}}`), "params.forms[1].op: unknown op"},
		{"row value", piece(`{"kind":"row","params":{"row":[0,1,500]}}`), "params.row[2]"},
		{"numseq missing", piece(`{"kind":"numseq","params":{"length":4}}`), "params.seq: is required"},
		{"numseq unknown", piece(`{"kind":"numseq","params":{"length":4,"seq":"pi"}}`), `params.seq: unknown sequence "pi"`},
		{"pulse division", piece(`{"kind":"pulse","params":{"divisions":[2,0]}}`), "params.divisions[1]"},
		{"pulse none", piece(`{"kind":"pulse","params":{}}`), "params.divisions: needs 1..8"},
		{"tendency segs", piece(`{"kind":"tendency","params":{"length":8}}`), "params.segments: needs 1..1024"},
		{"tendency n", piece(`{"kind":"tendency","params":{"length":8,"segments":[{"n":0,"low":1,"high":2}]}}`), "params.segments[0].n"},
		{"tendency beta", piece(`{"kind":"tendency","params":{"length":8,"beta":[0,1],"segments":[{"n":1,"low":1,"high":2}]}}`), "params.beta"},
		{"choose weights", piece(`{"kind":"choose","params":{"length":8,"options":[{"pitch":60,"weight":0}]}}`), "weights add up to 0"},
		{"choose negative", piece(`{"kind":"choose","params":{"length":8,"options":[{"pitch":60,"weight":-1}]}}`), "params.options[0].weight"},
		{"cloud density", piece(`{"kind":"cloud","params":{"length":8,"density":-1}}`), "params.density"},
		{"cloud range", piece(`{"kind":"cloud","params":{"length":8,"density":1,"low":80,"high":40}}`), "params.low: must be <= high"},
		{"cloud flood", piece(`{"kind":"cloud","params":{"length":100000,"density":100}}`), "params.density"},
		{"harmony table", piece(`{"kind":"progression","params":{"bars":2,"table":"nope"}}`), "params.table: unknown table"},
		{"harmony degree", piece(`{"kind":"progression","params":{"bars":2,"degrees":[1,0]}}`), "params.degrees[1]"},
		{"harmony style", piece(`{"kind":"progression","params":{"bars":2,"style":"x"}}`), "params.style: unknown style"},
		{"chordmelody range", piece(`{"kind":"chordmelody","params":{"bars":2,"low":60,"high":65}}`), "params.high: must be at least an octave"},
		{"chordmelody beat", piece(`{"kind":"chordmelody","params":{"bars":2,"beat_steps":5}}`), "params.beat_steps: must divide"},
		{"notes empty", piece(`{"kind":"notes","params":{"notes":[]}}`), "params.notes: needs at least one"},
		{"notes bad", piece(`{"kind":"notes","params":{"notes":[{"pitch":60,"start":0},{"pitch":60,"start":-1}]}}`), "params.notes[1].start"},
		{"concat empty", piece(`{"kind":"concat","params":{"items":[]}}`), "params.items: needs at least one item"},
		{"nested error path", piece(`{"kind":"concat","params":{"items":[` + leaf + `,{"kind":"repeat","params":{"n":2,"item":{"kind":"euclid","params":{"steps":4,"pulses":9}}}}]}}`),
			"tracks[0].source.params.items[1].params.item.params.pulses: must be in 0..4"},
		{"item missing", piece(`{"kind":"transpose","params":{"semitones":1}}`), "tracks[0].source.params.item: is required"},
		{"transpose missing", piece(`{"kind":"transpose","params":{"item":` + leaf + `}}`), "params.semitones: is required"},
		{"repeat n", piece(`{"kind":"repeat","params":{"n":0,"item":` + leaf + `}}`), "params.n: must be in 1..1000"},
		{"repeat too long", piece(`{"kind":"repeat","params":{"n":1000,"item":{"kind":"melody","params":{"length":99000}}}}`), "params.n"},
		{"stretch", piece(`{"kind":"stretch","params":{"factor":0,"item":` + leaf + `}}`), "params.factor: must be > 0"},
		{"slice", piece(`{"kind":"slice","params":{"from":5,"to":2,"item":` + leaf + `}}`), "params.to: must be > from"},
		{"quantize scale", piece(`{"kind":"quantize","params":{"item":` + leaf + `}}`), "params.scale: is required"},
		{"canon voices", piece(`{"kind":"canon","params":{"delay":1,"voices":99,"item":` + leaf + `}}`), "params.voices"},
		{"swing amount", piece(`{"kind":"swing","params":{"amount":2,"item":` + leaf + `}}`), "params.amount"},
		{"thin keep", piece(`{"kind":"thin","params":{"keep":2,"item":` + leaf + `}}`), "params.keep"},
		{"out of midi range", piece(`{"kind":"transpose","params":{"semitones":100,"item":` + leaf + `}}`), "outside the MIDI range"},
		{"too deep", piece(strings.Repeat(`{"kind":"gain","params":{"factor":1,"item":`, 40) + leaf + strings.Repeat(`}}`, 40)), "nested deeper"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := render(t, c.doc, nil)
			if err == nil {
				t.Fatalf("expected an error containing %q", c.want)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("error %q does not contain %q", err, c.want)
			}
		})
	}
}

func TestErrorType(t *testing.T) {
	_, err := render(t, piece(`{"kind":"melody","params":{"length":-1}}`), nil)
	var e *Error
	if !errors.As(err, &e) {
		t.Fatalf("expected *Error, got %T", err)
	}
	if e.Path != "tracks[0].source.params.length" || e.Msg != "must be > 0, got -1" {
		t.Errorf("got path %q msg %q", e.Path, e.Msg)
	}
}

func TestAllTrackErrorsReported(t *testing.T) {
	doc := `{"tracks":[{"source":{"kind":"x"}},{"source":` + leaf + `},{"channel":99,"source":` + leaf + `}]}`
	_, err := render(t, doc, nil)
	if err == nil {
		t.Fatal("expected errors")
	}
	for _, want := range []string{"tracks[0].source.kind", "tracks[2].channel"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("%q missing from %q", want, err)
		}
	}
	if strings.Contains(err.Error(), "tracks[1]") {
		t.Errorf("valid track reported: %v", err)
	}
}

func TestTrailingData(t *testing.T) {
	if _, err := ParseBytes([]byte(`{"tracks":[{}]} {}`)); err == nil {
		t.Error("expected an error for trailing data")
	}
}

// Library constructors panic on bad arguments: whatever slips through validation must become an error.
func TestPanicsBecomeErrors(t *testing.T) {
	Register(Kind{Name: "test_panic", Category: "generator", Doc: "panics",
		Build: func(*Builder, string, json.RawMessage) (Line, error) { panic("boom") }})
	defer delete(kinds, "test_panic")
	_, err := render(t, piece(`{"kind":"test_panic"}`), nil)
	if err == nil || !strings.Contains(err.Error(), "tracks[0].source: kind \"test_panic\" failed: boom") {
		t.Errorf("got %v", err)
	}
}

func TestRegisterRejectsDuplicates(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected a panic")
		}
	}()
	Register(kinds["melody"])
}

func TestNoteNames(t *testing.T) {
	good := map[string]float64{"C4": 60, "c4": 60, "A4": 69, "C#4": 61, "Db4": 61, "Bb2": 46, "C-1": 0, "G9": 127, "E+3": 52}
	for s, want := range good {
		if got, err := ParseNoteName(s); err != nil || got != want {
			t.Errorf("%q: got %v, %v want %v", s, got, err, want)
		}
	}
	for _, s := range []string{"", "H4", "C", "C4x", "4", "C99", "#4"} {
		if _, err := ParseNoteName(s); err == nil {
			t.Errorf("%q: expected an error", s)
		}
	}
}

func TestSieveExpr(t *testing.T) {
	cases := []struct {
		expr string
		want []int
	}{
		{"3@0", []int{0, 3, 6, 9}},
		{"3@1|5@0", []int{0, 1, 4, 5, 7, 10}},
		{"3@0&2@0", []int{0, 6}},
		{"3@0-2@0", []int{3, 9}},
		{"!3@0", []int{1, 2, 4, 5, 7, 8, 10, 11}},
		{" ( 3@0 | 2@1 ) & !3@0 ", []int{1, 5, 7, 11}},
		{"3@0|2@1&5@1", []int{0, 1, 3, 6, 9, 11}},
	}
	for _, c := range cases {
		s, err := ParseSieve(c.expr, 0, 11)
		if err != nil {
			t.Errorf("%q: %v", c.expr, err)
			continue
		}
		if fmt.Sprint(s.Values()) != fmt.Sprint(c.want) {
			t.Errorf("%q: got %v want %v", c.expr, s.Values(), c.want)
		}
	}
	for _, bad := range []string{"", "3", "3@", "@1", "3@0)", "(3@0", "3@0 &", "3@0 x", "99999999999@1", strings.Repeat("(", 100) + "3@0"} {
		if _, err := ParseSieve(bad, 0, 11); err == nil {
			t.Errorf("%q: expected an error", bad)
		}
	}
}

// ---- structure and reproducibility

func TestGoldenStructure(t *testing.T) {
	doc := `{"tempo":100,"steps_per_beat":2,"seed":9,"time_signature":{"num":3,"den":4},"tracks":[
	  {"name":"a","channel":2,"program":19,"source":{"kind":"euclid","params":{"steps":8,"pulses":3,"pitch":"C2"}}},
	  {"name":"b","channel":5,"program":40,"source":{"kind":"concat","params":{"items":[
	     {"kind":"notes","params":{"notes":[{"pitch":60,"start":0,"duration":2}],"length":4}},
	     {"kind":"notes","params":{"notes":[{"pitch":64,"start":1,"duration":1}],"length":4}}]}}}]}`
	res, err := render(t, doc, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Notes != 3+2 || res.Steps != 8 || res.Seed != 9 {
		t.Fatalf("got %+v", res)
	}
	var buf bytes.Buffer
	if err := midi.Write(&buf, res.File); err != nil {
		t.Fatal(err)
	}
	back, err := midi.ReadWith(&buf, 2)
	if err != nil {
		t.Fatal(err)
	}
	if back.Tempo != 100 || len(back.Tracks) != 2 {
		t.Fatalf("read back %+v", back)
	}
	byName := map[string]midi.Track{}
	for _, tr := range back.Tracks {
		byName[tr.Name] = tr
	}
	a, b := byName["a"], byName["b"]
	if a.Channel != 2 || a.Program != 19 || len(a.Events) != 3 || a.Events[0].Pitch != 36 {
		t.Errorf("track a: %+v", a)
	}
	if b.Channel != 5 || b.Program != 40 || len(b.Events) != 2 {
		t.Fatalf("track b: %+v", b)
	}
	if b.Events[0].Pitch != 60 || b.Events[0].Start != 0 || b.Events[0].Duration != 2 ||
		b.Events[1].Pitch != 64 || b.Events[1].Start != 5 {
		t.Errorf("track b events: %+v", b.Events)
	}
	if len(back.TimeSignatures) != 1 || back.TimeSignatures[0].Num != 3 || back.TimeSignatures[0].Den != 4 {
		t.Errorf("time signature: %+v", back.TimeSignatures)
	}
}

func TestCombinators(t *testing.T) {
	notes := `{"kind":"notes","params":{"length":4,"notes":[{"pitch":60,"start":0,"duration":1},{"pitch":64,"start":2,"duration":1}]}}`
	wrapN := func(kind, params string) string {
		return `{"kind":"` + kind + `","params":{` + params + `"item":` + notes + `}}`
	}
	type ev struct{ p, s float64 }
	cases := []struct {
		name, src string
		want      []ev
	}{
		{"transpose", wrapN("transpose", `"semitones":-12,`), []ev{{48, 0}, {52, 2}}},
		{"retrograde", wrapN("retrograde", ``), []ev{{64, 1}, {60, 3}}},
		{"shift", wrapN("shift", `"steps":3,`), []ev{{60, 3}, {64, 5}}},
		{"stretch", wrapN("stretch", `"factor":2,`), []ev{{60, 0}, {64, 4}}},
		{"slice", wrapN("slice", `"from":1,"to":4,`), []ev{{64, 1}}},
		{"invert", wrapN("invert", `"axis":60,`), []ev{{60, 0}, {56, 2}}},
		{"repeat", wrapN("repeat", `"n":2,`), []ev{{60, 0}, {64, 2}, {60, 4}, {64, 6}}},
		{"gain", wrapN("gain", `"factor":0,`), []ev{{60, 0}, {64, 2}}},
		{"canon", wrapN("canon", `"delay":1,"interval":12,"voices":2,`), []ev{{60, 0}, {72, 1}, {64, 2}, {76, 3}}},
		{"concat", `{"kind":"concat","params":{"items":[` + notes + `,` + notes + `]}}`, []ev{{60, 0}, {64, 2}, {60, 4}, {64, 6}}},
		{"layer", `{"kind":"layer","params":{"items":[` + notes + `,` + wrapN("transpose", `"semitones":1,`) + `]}}`, []ev{{60, 0}, {61, 0}, {64, 2}, {65, 2}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, err := render(t, piece(c.src), nil)
			if err != nil {
				t.Fatal(err)
			}
			var got []ev
			for _, e := range res.File.Tracks[0].Events {
				got = append(got, ev{e.Pitch, e.Start})
			}
			if fmt.Sprint(got) != fmt.Sprint(c.want) {
				t.Errorf("got %v want %v", got, c.want)
			}
		})
	}
	// the lengths add up: concat of two 4-step lines then repeat 3 gives 24 steps
	res, err := render(t, piece(`{"kind":"repeat","params":{"n":3,"item":{"kind":"concat","params":{"items":[`+notes+`,`+notes+`]}}}}`), nil)
	if err != nil || res.Steps != 24 {
		t.Errorf("steps: %v %v", res.Steps, err)
	}
}

func TestEuclidKnownPattern(t *testing.T) {
	res, err := render(t, piece(`{"kind":"euclid","params":{"steps":8,"pulses":3,"length":16,"pitch":36}}`), nil)
	if err != nil {
		t.Fatal(err)
	}
	var starts []float64
	for _, e := range res.File.Tracks[0].Events {
		starts = append(starts, e.Start)
	}
	if fmt.Sprint(starts) != "[0 3 6 8 11 14]" {
		t.Errorf("starts %v", starts)
	}
}

func TestSeedReproducible(t *testing.T) {
	doc := piece(`{"kind":"layer","params":{"items":[` + validNodes["melody"] + `,` + validNodes["cloud"] + `,` + validNodes["row"] + `,` +
		validNodes["progression"] + `,` + validNodes["thin"] + `,` + validNodes["humanize"] + `]}}`)
	bytesOf := func(seed *uint64) []byte {
		res, err := render(t, doc, seed)
		if err != nil {
			t.Fatal(err)
		}
		var buf bytes.Buffer
		if err := midi.Write(&buf, res.File); err != nil {
			t.Fatal(err)
		}
		return buf.Bytes()
	}
	three, four := uint64(3), uint64(4)
	if !bytes.Equal(bytesOf(nil), bytesOf(nil)) {
		t.Error("the same spec gave different files")
	}
	if !bytes.Equal(bytesOf(&three), bytesOf(&three)) {
		t.Error("the same seed gave different files")
	}
	if bytes.Equal(bytesOf(&three), bytesOf(&four)) {
		t.Error("different seeds gave the same file")
	}
	if bytes.Equal(bytesOf(nil), bytesOf(&three)) {
		t.Error("the seed override had no effect")
	}
}

func TestDefaultSeedIsFixed(t *testing.T) {
	doc := `{"tracks":[{"source":` + validNodes["melody"] + `}]}`
	a, err1 := render(t, doc, nil)
	b, err2 := render(t, doc, nil)
	if err1 != nil || err2 != nil || a.Seed != 1 || fmt.Sprint(a.File.Tracks) != fmt.Sprint(b.File.Tracks) {
		t.Errorf("seed %d, %v %v", a.Seed, err1, err2)
	}
}

func TestSamplesRender(t *testing.T) {
	files, err := filepath.Glob("../examples/spec/*.json")
	if err != nil || len(files) < 3 {
		t.Fatalf("sample specs: %v %v", files, err)
	}
	for _, f := range files {
		t.Run(filepath.Base(f), func(t *testing.T) {
			data, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			res, err := render(t, string(data), nil)
			if err != nil {
				t.Fatal(err)
			}
			var buf bytes.Buffer
			if err := midi.Write(&buf, res.File); err != nil {
				t.Fatal(err)
			}
			back, err := midi.Read(&buf)
			if err != nil || len(back.Tracks) != len(res.File.Tracks) {
				t.Fatalf("read back: %v, %d tracks", err, len(back.Tracks))
			}
		})
	}
}

// Mutating every single value of every valid spec must give errors or valid files, never panics.
func TestNoPanicsOnMutations(t *testing.T) {
	mutations := []any{0.0, -1.0, 1e30, nil, "x", []any{}, map[string]any{}, true, []any{1.0, 2.0, 3.0}, 1.5, []any{[]any{}}, "C4", 0.5, 3.0, 64.0}
	for name, node := range validNodes {
		var base any
		if err := json.Unmarshal([]byte(piece(node)), &base); err != nil {
			t.Fatal(err)
		}
		var paths [][]any
		var walk func(v any, path []any)
		walk = func(v any, path []any) {
			switch x := v.(type) {
			case map[string]any:
				for k, c := range x {
					p := append(slices.Clone(path), k)
					paths = append(paths, p)
					walk(c, p)
				}
			case []any:
				for i, c := range x {
					p := append(slices.Clone(path), i)
					paths = append(paths, p)
					walk(c, p)
				}
			}
		}
		walk(base, nil)
		orig, _ := json.Marshal(base)
		for _, path := range paths {
			for _, m := range mutations {
				var doc any
				json.Unmarshal(orig, &doc)
				cur := doc
				for _, step := range path[:len(path)-1] {
					switch k := step.(type) {
					case string:
						cur = cur.(map[string]any)[k]
					case int:
						cur = cur.([]any)[k]
					}
				}
				switch k := path[len(path)-1].(type) {
				case string:
					cur.(map[string]any)[k] = m
				case int:
					cur.([]any)[k] = m
				}
				data, _ := json.Marshal(doc)
				func() {
					defer func() {
						if r := recover(); r != nil {
							t.Errorf("%s, path %v = %v panicked: %v\n%s", name, path, m, r, data)
						}
					}()
					if _, err := render(t, string(data), nil); err != nil && strings.Contains(err.Error(), "\" failed: ") {
						t.Errorf("%s, path %v = %v: a library panic was caught instead of validated: %v", name, path, m, err)
					}
				}()
			}
		}
	}
}
