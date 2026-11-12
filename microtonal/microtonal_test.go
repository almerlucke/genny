package microtonal

import (
	"bytes"
	"math"
	"slices"
	"strings"
	"testing"

	"github.com/almerlucke/genny/midi"
	"github.com/almerlucke/genny/note"
	"github.com/almerlucke/genny/scale"
)

func TestTuning(t *testing.T) {
	e31 := EDO(60, 31)
	cases := []struct {
		name string
		got  float64
		want float64
	}{
		{"edo31 degree 0", e31.Pitch(0), 60},
		{"edo31 degree 31 is the octave", e31.Pitch(31), 72},
		{"edo31 degree -31", e31.Pitch(-31), 48},
		{"edo31 degree 18 is about a fifth", e31.Pitch(18), 60 + 1200*18.0/31/100},
		{"note", e31.Note(3, 2), 60 + 24 + 1200*3.0/31/100},
		{"just fifth", Just(60, 2, 1, 3.0/2).Pitch(1), 60 + 12*math.Log2(1.5)},
		{"bohlen-pierce period", EqualDivision(60, 13, 1200*math.Log2(3)).Pitch(13), 60 + 12*math.Log2(3)},
		{"snap", e31.Snap(60.01), 60},
		{"snap up", e31.Snap(60.35), 60 + 1200*1.0/31/100},
	}
	for _, c := range cases {
		if math.Abs(c.got-c.want) > 1e-9 {
			t.Errorf("%s: %v, want %v", c.name, c.got, c.want)
		}
	}
	if d := e31.Nearest(60 + 12.0/31*100/100*3 + 0.01); d != 3 {
		t.Errorf("nearest %d", d)
	}
}

func TestFromScl(t *testing.T) {
	s, err := scale.ParseScl(strings.NewReader(scale.SclBohlenPierce))
	if err != nil {
		t.Fatal(err)
	}
	tu := FromScl(60, s)
	if tu.Len() != 13 || math.Abs(tu.Period-1901.955) > 0.001 {
		t.Fatalf("len %d period %v", tu.Len(), tu.Period)
	}
	want := s.Scale(60)
	for i := -14; i < 30; i++ {
		if math.Abs(tu.Pitch(i)-want.Degree(i)) > 1e-9 {
			t.Errorf("degree %d: %v vs scale %v", i, tu.Pitch(i), want.Degree(i))
		}
	}
}

func TestPanics(t *testing.T) {
	for name, f := range map[string]func(){
		"edo 0":      func() { EDO(60, 0) },
		"no cents":   func() { FromCents(60, 1200) },
		"not zero":   func() { FromCents(60, 1200, 5) },
		"descending": func() { FromCents(60, 1200, 0, 700, 300) },
		"past":       func() { FromCents(60, 1200, 0, 1200) },
		"period":     func() { FromCents(60, 0, 0) },
		"empty scl":  func() { FromScl(60, scale.Scl{}) },
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: no panic", name)
				}
			}()
			f()
		}()
	}
}

func TestLocateExact(t *testing.T) {
	cases := []struct {
		pitch float64
		rng   float64
		want  Bend
	}{
		{60, 2, Bend{60, 8192}},
		{60.25, 2, Bend{60, 8192 + 1024}},
		{60.5, 2, Bend{61, 8192 - 2048}}, // a quarter tone: +-50 cents is 2048 steps at range 2
		{60.4998, 2, Bend{60, 8192 + 2047}},
		{59.75, 2, Bend{60, 8192 - 1024}},
		{59.5, 2, Bend{60, 8192 - 2048}},
		{60.25, 4, Bend{60, 8192 + 512}},
	}
	for _, c := range cases {
		got, ok := Locate(c.pitch, c.rng)
		if !ok || got != c.want {
			t.Errorf("Locate(%v, %v) = %v %v, want %v", c.pitch, c.rng, got, ok, c.want)
		}
	}
	for _, p := range []float64{-1, 128, math.NaN(), math.Inf(1)} {
		if _, ok := Locate(p, 2); ok {
			t.Errorf("Locate(%v) ok", p)
		}
	}
	if c := (Bend{60, 8192 + 2048}).Cents(2); c != 50 {
		t.Errorf("cents %v", c)
	}
}

func TestLocateRoundTrip(t *testing.T) {
	for _, r := range []float64{2, 4, 12} {
		half := r * 100 / 8192 / 2 // half a bend step in cents
		worst := 0.0
		for p := 36.0; p < 96; p += 0.00731 {
			b, ok := Locate(p, r)
			if !ok {
				t.Fatalf("not ok at %v", p)
			}
			if b.Value < 0 || b.Value > 16383 {
				t.Fatalf("bend %d out of 14 bit", b.Value)
			}
			// the offset stays within the nearest note (or its neighbour at the tie)
			if math.Abs(b.Cents(r)) > 50.0001 {
				t.Fatalf("offset %v cents at %v", b.Cents(r), p)
			}
			worst = max(worst, math.Abs(b.Pitch(r)-p)*100)
		}
		if worst > half+1e-6 {
			t.Errorf("range %v: worst error %v cents, half a step is %v", r, worst, half)
		}
	}
	// at the default range a step is 0.0244 cents, so the error is at most 0.0122
	if h := 2 * 100 / 8192.0 / 2; h >= 0.0123 {
		t.Errorf("half step %v", h)
	}
}

func TestSingle(t *testing.T) {
	e31 := EDO(60, 31)
	var evs []note.Event
	for i := range 6 {
		evs = append(evs, note.Event{Pitch: e31.Pitch(i), Velocity: 0.8, Start: float64(i), Duration: 1})
	}
	tracks := Render(evs, &e31, Strategy{Channel: 3, Program: 5, Name: "lead", Range: 2})
	if len(tracks) != 1 || tracks[0].Channel != 3 || !tracks[0].PitchBend || tracks[0].Program != 5 {
		t.Fatalf("tracks %+v", tracks)
	}
	want := []midi.Controller{{Number: 101, Value: 0}, {Number: 100, Value: 0}, {Number: 6, Value: 2}, {Number: 38, Value: 0}, {Number: 101, Value: 127}, {Number: 100, Value: 127}}
	if !slices.Equal(tracks[0].Controllers, want) {
		t.Errorf("rpn %v", tracks[0].Controllers)
	}
	if len(tracks[0].Events) != 6 {
		t.Errorf("events %d", len(tracks[0].Events))
	}
}

func TestRangeSetup(t *testing.T) {
	c := Strategy{Range: 12.5}.rangeSetup()
	if c[2].Value != 12 || c[3].Value != 50 {
		t.Errorf("%v", c)
	}
	if (Strategy{Range: 1}).rangeOrDefault() != 2 || (Strategy{Range: 500}).rangeOrDefault() != 96 {
		t.Error("range clamp")
	}
}

func chord(start float64, pitches ...float64) []note.Event {
	var out []note.Event
	for _, p := range pitches {
		out = append(out, note.Event{Pitch: p, Velocity: 0.7, Start: start, Duration: 2})
	}
	return out
}

func TestMPEChannels(t *testing.T) {
	e31 := EDO(60, 31)
	// two overlapping chords of 3 notes: 6 notes sound at 1..2
	evs := append(chord(0, e31.Pitch(0), e31.Pitch(10), e31.Pitch(18)), chord(1, e31.Pitch(3), e31.Pitch(13), e31.Pitch(21))...)
	evs = append(evs, chord(2, e31.Pitch(5))...) // starts when the first chord ends
	tracks, skipped := Plan(evs, &e31, Strategy{MPE: true})
	if skipped != 0 {
		t.Fatalf("skipped %d", skipped)
	}
	if tracks[0].Channel != 0 || len(tracks[0].Events) != 0 {
		t.Fatalf("manager %+v", tracks[0])
	}
	if c := tracks[0].Controllers; c[0].Number != 101 || c[1].Number != 100 || c[1].Value != 6 || c[2].Value != 15 {
		t.Errorf("mpe config %v", c)
	}
	// no channel is used twice at the same time, and none is the manager channel
	type iv struct{ s, e float64 }
	for _, tr := range tracks[1:] {
		if tr.Channel < 1 {
			t.Errorf("member on channel %d", tr.Channel)
		}
		var last iv
		for _, e := range tr.Events {
			if e.Start < last.e {
				t.Errorf("channel %d overlaps: %v starts before %v", tr.Channel, e.Start, last.e)
			}
			last = iv{e.Start, e.Start + e.Duration}
		}
	}
	used := map[int]bool{}
	n := 0
	for _, tr := range tracks[1:] {
		used[tr.Channel] = true
		n += len(tr.Events)
	}
	if n != 7 || len(used) != 7 {
		t.Errorf("%d notes on %d channels", n, len(used))
	}
}

func TestMPEFull(t *testing.T) {
	var evs []note.Event
	for i := range 20 {
		evs = append(evs, note.Event{Pitch: 60 + float64(i)*0.1, Velocity: 1, Start: 0, Duration: 4})
	}
	tracks, skipped := Plan(evs, nil, Strategy{MPE: true, Members: 4})
	if skipped != 16 || len(tracks) != 5 {
		t.Errorf("skipped %d tracks %d", skipped, len(tracks))
	}
	if c := tracks[0].Controllers[2].Value; c != 4 {
		t.Errorf("members %d", c)
	}
	// invalid events are skipped
	_, skipped = Plan([]note.Event{{Pitch: 200, Start: 0, Duration: 1}, {Pitch: 60, Start: -1, Duration: 1}, {Pitch: 60, Duration: 0}}, nil, Strategy{})
	if skipped != 3 {
		t.Errorf("skipped %d", skipped)
	}
	if tr := Render(nil, nil, Strategy{}); len(tr) != 0 {
		t.Error("empty input")
	}
}

func TestRoundTripFile(t *testing.T) {
	e31 := EDO(60, 31)
	var evs []note.Event
	evs = append(evs, chord(0, e31.Pitch(0), e31.Pitch(10), e31.Pitch(18))...)
	evs = append(evs, chord(1, e31.Pitch(4), e31.Pitch(14))...)
	evs = append(evs, note.Event{Pitch: 60.5, Velocity: 0.5, Start: 4, Duration: 1}) // exact quarter tone
	tracks := Render(evs, nil, Strategy{MPE: true})

	var buf bytes.Buffer
	if err := midi.Write(&buf, midi.File{Tracks: tracks}); err != nil {
		t.Fatal(err)
	}
	f, err := midi.Read(&buf)
	if err != nil {
		t.Fatal(err)
	}
	var got []float64
	for _, tr := range f.Tracks {
		for _, e := range tr.Events {
			got = append(got, e.Pitch)
		}
	}
	var want []float64
	for _, e := range evs {
		want = append(want, e.Pitch)
	}
	slices.Sort(got)
	slices.Sort(want)
	if len(got) != len(want) {
		t.Fatalf("%d notes back, want %d", len(got), len(want))
	}
	for i := range got {
		if math.Abs(got[i]-want[i])*100 > 0.0123 {
			t.Errorf("note %d: %v, want %v (%v cents off)", i, got[i], want[i], (got[i]-want[i])*100)
		}
	}
}

func TestDeterministic(t *testing.T) {
	e := EDO(60, 19)
	evs := chord(0, e.Pitch(1), e.Pitch(8), e.Pitch(11))
	a, b := Render(evs, &e, Strategy{MPE: true}), Render(evs, &e, Strategy{MPE: true})
	var x, y bytes.Buffer
	midi.Write(&x, midi.File{Tracks: a})
	midi.Write(&y, midi.File{Tracks: b})
	if !bytes.Equal(x.Bytes(), y.Bytes()) {
		t.Error("not deterministic")
	}
}

func TestMPELeavesTheDrumChannelAlone(t *testing.T) {
	var evs []note.Event
	for i := 0; i < 14; i++ {
		evs = append(evs, note.Event{Pitch: 60 + float64(i), Velocity: 0.7, Start: 0, Duration: 4})
	}
	has := func(tracks []midi.Track, ch int) bool {
		for _, tr := range tracks {
			if tr.Channel == ch && len(tr.Events) > 0 {
				return true
			}
		}
		return false
	}
	tracks, skipped := Plan(evs, nil, Strategy{MPE: true, Members: 14})
	if has(tracks, 9) {
		t.Error("channel 10 (index 9) must not carry notes by default")
	}
	if skipped != 1 {
		t.Errorf("14 simultaneous notes on 13 usable channels: skipped %d, want 1", skipped)
	}
	tracks, skipped = Plan(evs, nil, Strategy{MPE: true, Members: 14, DrumChannel: true})
	if !has(tracks, 9) || skipped != 0 {
		t.Errorf("DrumChannel should use channel 10: has=%v skipped=%d", has(tracks, 9), skipped)
	}
}
