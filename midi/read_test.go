package midi

import (
	"bytes"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/almerlucke/genny/note"
)

func roundTrip(t *testing.T, f File) File {
	t.Helper()
	var buf bytes.Buffer
	if err := Write(&buf, f); err != nil {
		t.Fatal(err)
	}
	got, err := Read(&buf)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func near(a, b float64) bool { return math.Abs(a-b) < 0.01 }

func sameEvents(t *testing.T, got, want []note.Event) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%d events, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		g, w := got[i], want[i]
		if !near(g.Pitch, w.Pitch) || !near(g.Start, w.Start) || !near(g.Duration, w.Duration) || !near(g.Velocity, w.Velocity) {
			t.Errorf("event %d: %+v, want %+v", i, g, w)
		}
	}
}

func TestRoundTripNotes(t *testing.T) {
	events := []note.Event{
		{Pitch: 60, Velocity: 1, Start: 0, Duration: 1},
		{Pitch: 64, Velocity: 64.0 / 127, Start: 1, Duration: 2.5}, // a fractional step
		{Pitch: 67, Velocity: 102.0 / 127, Start: 1, Duration: 1},
		{Pitch: 60, Velocity: 1, Start: 3.5, Duration: 4},
	}
	got := roundTrip(t, File{Tempo: 90, Tracks: []Track{{Name: "lead", Channel: 2, Program: 5, Events: events}}})
	if !near(got.Tempo, 90) || got.StepsPerBeat != 4 || len(got.Tracks) != 1 || len(got.TempoChanges) != 0 {
		t.Fatalf("%+v", got)
	}
	tr := got.Tracks[0]
	if tr.Name != "lead" || tr.Channel != 2 || tr.Program != 5 || tr.PitchBend {
		t.Errorf("%+v", tr)
	}
	sameEvents(t, tr.Events, events)
}

func TestStepsPerBeat(t *testing.T) {
	var buf bytes.Buffer
	if err := Write(&buf, File{StepsPerBeat: 2, Tracks: []Track{{Events: []note.Event{{Pitch: 60, Velocity: 1, Start: 2, Duration: 1}}}}}); err != nil {
		t.Fatal(err)
	}
	// a beat is 2 steps in the file, read with 8 steps in a beat everything is 4 times as long
	got, err := ReadWith(&buf, 8)
	if err != nil {
		t.Fatal(err)
	}
	sameEvents(t, got.Tracks[0].Events, []note.Event{{Pitch: 60, Velocity: 1, Start: 8, Duration: 4}})
	if got.StepsPerBeat != 8 {
		t.Error(got.StepsPerBeat)
	}
	if _, err := ReadWith(bytes.NewReader(nil), 0); err == nil {
		t.Error("zero steps per beat")
	}
}

func TestRoundTripPitchBend(t *testing.T) {
	events := []note.Event{
		{Pitch: 60.5, Velocity: 1, Start: 0, Duration: 1},
		{Pitch: 62, Velocity: 1, Start: 1, Duration: 1},
		{Pitch: 63.25, Velocity: 1, Start: 2, Duration: 1},
		{Pitch: 58.75, Velocity: 1, Start: 3, Duration: 1},
	}
	got := roundTrip(t, File{Tracks: []Track{{PitchBend: true, Events: events}}})
	if !got.Tracks[0].PitchBend {
		t.Error("no pitch bend")
	}
	sameEvents(t, got.Tracks[0].Events, events)
}

func TestRoundTripTempoAndMeter(t *testing.T) {
	f := File{
		Tempo:        66,
		TempoChanges: []TempoChange{{Step: 16, BPM: 100}, {Step: 8.5, BPM: 80}},
		TimeSignatures: []TimeSignature{
			{Step: 0, Num: 12, Den: 8}, {Step: 24, Num: 7, Den: 8}, {Step: 48, Num: 4, Den: 4},
		},
		Tracks: []Track{{Events: []note.Event{{Pitch: 60, Velocity: 1, Duration: 1}}}},
	}
	got := roundTrip(t, f)
	if !near(got.Tempo, 66) || len(got.TempoChanges) != 2 {
		t.Fatalf("%+v", got)
	}
	// they come back in order of time
	if c := got.TempoChanges; !near(c[0].Step, 8.5) || !near(c[0].BPM, 80) || !near(c[1].Step, 16) || !near(c[1].BPM, 100) {
		t.Errorf("%+v", c)
	}
	if len(got.TimeSignatures) != 3 {
		t.Fatalf("%+v", got.TimeSignatures)
	}
	for i, w := range f.TimeSignatures {
		if g := got.TimeSignatures[i]; g != w {
			t.Errorf("time signature %d: %+v, want %+v", i, g, w)
		}
	}
}

func TestTempoMapInTempoTrack(t *testing.T) {
	// ticks are beats, a tempo change only changes the clock: notes stay where they are
	p := write(t, File{TempoChanges: []TempoChange{{Step: 4, BPM: 60}}, Tracks: []Track{{Events: []note.Event{{Pitch: 60, Velocity: 1, Start: 8, Duration: 1}}}}})
	if len(p.tracks) != 2 || !p.tracks[0].ended {
		t.Fatalf("%+v", p.tracks)
	}
	if e := p.tracks[1].events[1]; e.tick != 960 {
		t.Errorf("%+v", e)
	}
}

func TestRoundTripControllersAndPrograms(t *testing.T) {
	tr := Track{
		Channel: 3, Program: 10,
		Controllers:    []Controller{{Step: 0, Number: 7, Value: 100}, {Step: 2, Number: 10, Value: 0}, {Step: 6.5, Number: 1, Value: 127}},
		ProgramChanges: []ProgramChange{{Step: 4, Program: 20}, {Step: 8, Program: 30}},
		Events: []note.Event{
			{Pitch: 60, Velocity: 1, Start: 0, Duration: 2},
			{Pitch: 62, Velocity: 1, Start: 5, Duration: 2},
			{Pitch: 64, Velocity: 1, Start: 9, Duration: 2},
		},
	}
	got := roundTrip(t, File{Tracks: []Track{tr}}).Tracks[0]
	if got.Program != 10 || got.Channel != 3 {
		t.Errorf("%+v", got)
	}
	if len(got.Controllers) != 3 {
		t.Fatalf("%+v", got.Controllers)
	}
	for i, w := range tr.Controllers {
		g := got.Controllers[i]
		if !near(g.Step, w.Step) || g.Number != w.Number || g.Value != w.Value {
			t.Errorf("controller %d: %+v, want %+v", i, g, w)
		}
	}
	if len(got.ProgramChanges) != 2 || got.ProgramChanges[0].Program != 20 || !near(got.ProgramChanges[1].Step, 8) {
		t.Errorf("%+v", got.ProgramChanges)
	}
	sameEvents(t, got.Events, tr.Events)
}

func TestRoundTripSeveralTracks(t *testing.T) {
	f := File{Tracks: []Track{
		{Name: "a", Channel: 0, Events: []note.Event{{Pitch: 60, Velocity: 1, Duration: 1}}},
		{Name: "drums", Channel: 9, Events: []note.Event{{Pitch: 36, Velocity: 1, Start: 1, Duration: 1}}},
	}}
	got := roundTrip(t, f)
	if len(got.Tracks) != 2 || got.Tracks[1].Name != "drums" || got.Tracks[1].Channel != 9 || got.Tracks[1].Program != 0 {
		t.Fatalf("%+v", got.Tracks)
	}
}

// smf builds a file from track bodies
func smf(format, division int, tracks ...[]byte) []byte {
	out := []byte("MThd")
	out = append(out, 0, 0, 0, 6, byte(format>>8), byte(format), byte(len(tracks)>>8), byte(len(tracks)), byte(division>>8), byte(division))
	for _, tr := range tracks {
		out = append(out, chunk("MTrk", tr)...)
	}
	return out
}

func TestReadHandMade(t *testing.T) {
	// format 0: running status, note on with velocity 0 as a note off, a sysex and an unknown meta event,
	// division 96 ticks per beat
	tr := []byte{
		0, 0xff, 0x03, 2, 'h', 'i',
		0, 0xff, 0x51, 3, 0x07, 0xa1, 0x20, // 500000 us: 120 bpm
		0, 0xf0, 2, 1, 0xf7, // sysex
		0, 0xff, 0x7f, 1, 9, // sequencer specific
		0, 0x91, 60, 100, // running status from here
		24, 62, 100, // a second note on, a quarter beat later (one step)
		24, 60, 0, // velocity 0 ends the first
		24, 62, 0,
		0, 0xff, 0x2f, 0,
	}
	got, err := Read(bytes.NewReader(smf(0, 96, tr)))
	if err != nil {
		t.Fatal(err)
	}
	if !near(got.Tempo, 120) || len(got.Tracks) != 1 {
		t.Fatalf("%+v", got)
	}
	tk := got.Tracks[0]
	if tk.Name != "hi" || tk.Channel != 1 {
		t.Errorf("%+v", tk)
	}
	sameEvents(t, tk.Events, []note.Event{
		{Pitch: 60, Velocity: 100.0 / 127, Start: 0, Duration: 2},
		{Pitch: 62, Velocity: 100.0 / 127, Start: 1, Duration: 2},
	})
}

func TestReadOpenNotesAndRetrigger(t *testing.T) {
	tr := []byte{
		0, 0x90, 60, 90,
		48, 60, 90, // the same pitch again while it sounds: the first one ends
		48, 64, 90, // never switched off
		48, 0xff, 0x2f, 0,
	}
	got, err := Read(bytes.NewReader(smf(1, 96, tr)))
	if err != nil {
		t.Fatal(err)
	}
	sameEvents(t, got.Tracks[0].Events, []note.Event{
		{Pitch: 60, Velocity: 90.0 / 127, Start: 0, Duration: 2},
		{Pitch: 60, Velocity: 90.0 / 127, Start: 2, Duration: 4},
		{Pitch: 64, Velocity: 90.0 / 127, Start: 4, Duration: 2},
	})
}

func TestReadSkipsUnknownChunk(t *testing.T) {
	data := smf(1, 96, []byte{0, 0x90, 60, 90, 24, 0x80, 60, 0})
	data = append(data, chunk("XXXX", []byte{1, 2, 3})...)
	data = append(data, chunk("MTrk", []byte{0, 0x90, 62, 90, 24, 0x80, 62, 0})...)
	data[11] = 2 // two tracks; the unknown chunk between them does not count
	got, err := Read(bytes.NewReader(data))
	if err != nil || len(got.Tracks) != 2 {
		t.Fatalf("%v %+v", err, got)
	}
}

func TestReadErrors(t *testing.T) {
	good := smf(1, 96, []byte{0, 0x90, 60, 90, 24, 0x80, 60, 0})
	for name, data := range map[string][]byte{
		"empty":            nil,
		"no header":        []byte("RIFFxxxxxxxxxxxxxxxxxx"),
		"short header":     good[:10],
		"short header len": []byte("MThd\x00\x00\x00\x02\x00\x00"),
		"format 7":         smf(7, 96),
		"smpte":            smf(1, 0xe728),
		"zero division":    smf(1, 0),
		"truncated track":  good[:len(good)-3],
		"missing track":    good[:14],
		"data first":       smf(1, 96, []byte{0, 60, 90}),
		"high data bit":    smf(1, 96, []byte{0, 0x90, 60, 0x90}),
		"long vlq":         smf(1, 96, []byte{0x80, 0x80, 0x80, 0x80, 0x01, 0x90, 60, 90}),
		"meta too long":    smf(1, 96, []byte{0, 0xff, 0x03, 9, 'a'}),
		"sysex too long":   smf(1, 96, []byte{0, 0xf0, 5, 1}),
		"bad status":       smf(1, 96, []byte{0, 0xf1, 0}),
		"cut message":      smf(1, 96, []byte{0, 0x90, 60}),
		"huge chunk":       append(bytes.Clone(good[:14]), 'M', 'T', 'r', 'k', 0xff, 0xff, 0xff, 0xff),
	} {
		if _, err := Read(bytes.NewReader(data)); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
	if _, err := ReadFile(filepath.Join(t.TempDir(), "nothing.mid")); err == nil {
		t.Error("a missing file was read")
	}
}

func TestReadFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "x.mid")
	f := File{Tempo: 100, Tracks: []Track{{Name: "x", Events: []note.Event{{Pitch: 60, Velocity: 1, Duration: 1}}}}}
	if err := WriteFile(path, f); err != nil {
		t.Fatal(err)
	}
	got, err := ReadFile(path)
	if err != nil || len(got.Tracks) != 1 || !near(got.Tempo, 100) {
		t.Errorf("%v %+v", err, got)
	}
	if _, err := ReadFileWith(path, 2); err != nil {
		t.Error(err)
	}
	if err := os.WriteFile(path, []byte("garbage"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadFile(path); err == nil {
		t.Error("garbage was read")
	}
}

func TestWriteExtensionsInvalid(t *testing.T) {
	ev := []note.Event{{Pitch: 60, Velocity: 1, Duration: 1}}
	for name, f := range map[string]File{
		"controller number": {Tracks: []Track{{Events: ev, Controllers: []Controller{{Number: 128}}}}},
		"controller value":  {Tracks: []Track{{Events: ev, Controllers: []Controller{{Value: -1}}}}},
		"controller step":   {Tracks: []Track{{Events: ev, Controllers: []Controller{{Step: -1}}}}},
		"program change":    {Tracks: []Track{{Events: ev, ProgramChanges: []ProgramChange{{Program: 128}}}}},
		"tempo change":      {TempoChanges: []TempoChange{{BPM: 0}}},
		"tempo step":        {TempoChanges: []TempoChange{{Step: math.NaN(), BPM: 100}}},
		"signature den":     {TimeSignatures: []TimeSignature{{Num: 4, Den: 3}}},
		"signature num":     {TimeSignatures: []TimeSignature{{Num: 0, Den: 4}}},
		"signature zero":    {TimeSignatures: []TimeSignature{{}}},
	} {
		if err := Write(&bytes.Buffer{}, f); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
}

func TestReadNeverPanics(t *testing.T) {
	var buf bytes.Buffer
	err := Write(&buf, File{
		TempoChanges: []TempoChange{{Step: 4, BPM: 90}}, TimeSignatures: []TimeSignature{{Num: 3, Den: 4}},
		Tracks: []Track{{
			PitchBend: true, Controllers: []Controller{{Step: 1, Number: 7, Value: 9}},
			ProgramChanges: []ProgramChange{{Step: 3, Program: 4}},
			Events:         []note.Event{{Pitch: 60.5, Velocity: 1, Duration: 1}, {Pitch: 62, Velocity: 1, Start: 2, Duration: 2}},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	good := buf.Bytes()
	rng := rand.New(rand.NewSource(1))

	// every truncation
	for n := range len(good) {
		Read(bytes.NewReader(good[:n]))
	}
	// random byte changes, and garbage
	for range 3000 {
		data := bytes.Clone(good)
		for range 1 + rng.Intn(4) {
			data[rng.Intn(len(data))] = byte(rng.Intn(256))
		}
		Read(bytes.NewReader(data))
	}
	for range 500 {
		data := make([]byte, rng.Intn(200))
		rng.Read(data)
		Read(bytes.NewReader(data))
		Read(strings.NewReader("MThd\x00\x00\x00\x06\x00\x01\x00\x01\x01\xe0MTrk" + string(data)))
	}
}

func FuzzRead(f *testing.F) {
	var buf bytes.Buffer
	Write(&buf, File{
		Tempo: 100, TimeSignatures: []TimeSignature{{Num: 6, Den: 8}},
		Tracks: []Track{{
			Name: "x", PitchBend: true, Controllers: []Controller{{Step: 1, Number: 7, Value: 9}},
			Events: []note.Event{{Pitch: 60.5, Velocity: 1, Duration: 1}},
		}},
	})
	f.Add(buf.Bytes())
	f.Add(smf(0, 96, []byte{0, 0x91, 60, 100, 24, 62, 100, 24, 60, 0, 24, 62, 0, 0, 0xff, 0x2f, 0}))
	f.Add(smf(1, 480, []byte{0, 0xff, 0x51, 3, 7, 0xa1, 0x20, 0, 0xf0, 1, 0xf7, 0, 0xe0, 0, 64, 0, 0xc0, 5}))
	f.Add([]byte("MThd\x00\x00\x00\x06\x00\x01\x00\x01\x01\xe0MTrk\x00\x00\x00\x01\xff"))
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, data []byte) {
		got, err := Read(bytes.NewReader(data))
		if err != nil {
			return
		}
		for _, tr := range got.Tracks {
			for _, e := range tr.Events {
				if !(e.Duration > 0) || e.Start < 0 || math.IsNaN(e.Pitch) {
					t.Fatalf("bad event %+v", e)
				}
			}
		}
	})
}
