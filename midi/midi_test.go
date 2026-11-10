package midi

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/almerlucke/genny/note"
)

// parsed is what a reader of the file sees
type parsed struct {
	format, division int
	tracks           []track
}

type track struct {
	name   string
	tempo  int // microseconds per beat
	events []ev
	ended  bool
}

type ev struct {
	tick     int
	kind     string // on, off, bend, program
	ch, a, b int
}

func readVLQ(t *testing.T, b []byte, i *int) int {
	v := 0
	for {
		c := b[*i]
		*i++
		v = v<<7 | int(c&0x7f)
		if c&0x80 == 0 {
			return v
		}
	}
}

func parse(t *testing.T, data []byte) parsed {
	t.Helper()
	if string(data[:4]) != "MThd" || binary.BigEndian.Uint32(data[4:8]) != 6 {
		t.Fatal("bad header")
	}
	p := parsed{format: int(binary.BigEndian.Uint16(data[8:])), division: int(binary.BigEndian.Uint16(data[12:]))}
	n := int(binary.BigEndian.Uint16(data[10:]))
	pos := 14
	for range n {
		if string(data[pos:pos+4]) != "MTrk" {
			t.Fatalf("no track chunk at %d", pos)
		}
		size := int(binary.BigEndian.Uint32(data[pos+4:]))
		body := data[pos+8 : pos+8+size]
		pos += 8 + size

		var tr track
		tick := 0
		for i := 0; i < len(body); {
			tick += readVLQ(t, body, &i)
			status := body[i]
			i++
			switch {
			case status == 0xff:
				typ := body[i]
				i++
				l := readVLQ(t, body, &i)
				d := body[i : i+l]
				i += l
				switch typ {
				case 0x03:
					tr.name = string(d)
				case 0x51:
					tr.tempo = int(d[0])<<16 | int(d[1])<<8 | int(d[2])
				case 0x2f:
					tr.ended = i == len(body)
				}
			case status&0xf0 == 0x90:
				tr.events = append(tr.events, ev{tick, "on", int(status & 0xf), int(body[i]), int(body[i+1])})
				i += 2
			case status&0xf0 == 0x80:
				tr.events = append(tr.events, ev{tick, "off", int(status & 0xf), int(body[i]), int(body[i+1])})
				i += 2
			case status&0xf0 == 0xe0:
				tr.events = append(tr.events, ev{tick, "bend", int(status & 0xf), int(body[i]) | int(body[i+1])<<7, 0})
				i += 2
			case status&0xf0 == 0xc0:
				tr.events = append(tr.events, ev{tick, "program", int(status & 0xf), int(body[i]), 0})
				i++
			default:
				t.Fatalf("unknown status %x", status)
			}
		}
		p.tracks = append(p.tracks, tr)
	}
	if pos != len(data) {
		t.Fatal("trailing bytes")
	}
	return p
}

func write(t *testing.T, f File) parsed {
	t.Helper()
	var buf bytes.Buffer
	if err := Write(&buf, f); err != nil {
		t.Fatal(err)
	}
	return parse(t, buf.Bytes())
}

func TestNotesTempoAndTiming(t *testing.T) {
	p := write(t, File{
		Tempo: 90,
		Tracks: []Track{{
			Name: "lead", Channel: 2, Program: 5,
			Events: []note.Event{
				{Pitch: 60, Velocity: 1, Start: 0, Duration: 1},
				{Pitch: 64.4, Velocity: 0.5, Start: 4, Duration: 2}, // a beat later, at 4 steps per beat
			},
		}},
	})
	if p.format != 1 || p.division != 480 || len(p.tracks) != 2 {
		t.Fatalf("%+v", p)
	}
	if want := 666667; p.tracks[0].tempo != want || !p.tracks[0].ended {
		t.Errorf("tempo %d, want %d", p.tracks[0].tempo, want)
	}
	tr := p.tracks[1]
	want := []ev{
		{0, "program", 2, 5, 0},
		{0, "on", 2, 60, 127},
		{120, "off", 2, 60, 0}, // 1 step is a quarter of a beat
		{480, "on", 2, 64, 64}, // pitch rounded, velocity 0.5
		{720, "off", 2, 64, 0}, // 2 steps are half a beat
	}
	if tr.name != "lead" || !tr.ended || len(tr.events) != len(want) {
		t.Fatalf("%+v", tr)
	}
	for i, w := range want {
		if tr.events[i] != w {
			t.Errorf("event %d: %+v, want %+v", i, tr.events[i], w)
		}
	}
}

func TestBackToBackNotesAndLongDeltas(t *testing.T) {
	// the second note starts exactly when the first one stops: the off must come first.
	// A far away note needs a delta time of several bytes
	p := write(t, File{StepsPerBeat: 1, Tracks: []Track{{Channel: 9, Events: []note.Event{
		{Pitch: 36, Velocity: 1, Start: 0, Duration: 1},
		{Pitch: 36, Velocity: 1, Start: 1, Duration: 1},
		{Pitch: 38, Velocity: 1, Start: 1000, Duration: 1},
	}}}})
	var got []string
	last := 0
	for _, e := range p.tracks[1].events {
		got = append(got, e.kind)
		last = e.tick
	}
	want := []string{"on", "off", "on", "off", "on", "off"} // no program change on the drum channel
	if len(got) != len(want) {
		t.Fatalf("%v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%v, want %v", got, want)
		}
	}
	if last != 1001*480 {
		t.Errorf("last tick %d, want %d", last, 1001*480)
	}
}

func TestPitchBend(t *testing.T) {
	p := write(t, File{Tracks: []Track{{PitchBend: true, Events: []note.Event{
		{Pitch: 60.5, Velocity: 1, Start: 0, Duration: 1},
		{Pitch: 62, Velocity: 1, Start: 1, Duration: 1},
	}}}})
	var bends []int
	for _, e := range p.tracks[1].events {
		if e.kind == "bend" {
			bends = append(bends, e.a)
		}
	}
	// 60.5 is rounded to 61 and bent half a semitone down: 2048 below the center of 8192 (bend range 2 semitones)
	if len(bends) != 2 || bends[0] != 6144 || bends[1] != 8192 {
		t.Fatalf("bends %v, want [6144 8192]", bends)
	}
}

func TestInvalidInput(t *testing.T) {
	for name, f := range map[string]File{
		"channel":  {Tracks: []Track{{Channel: 16}}},
		"program":  {Tracks: []Track{{Program: 128}}},
		"tempo":    {Tempo: -1},
		"duration": {Tracks: []Track{{Events: []note.Event{{Pitch: 60, Velocity: 1, Duration: 0}}}}},
		"start":    {Tracks: []Track{{Events: []note.Event{{Pitch: 60, Velocity: 1, Start: -1, Duration: 1}}}}},
	} {
		if err := Write(&bytes.Buffer{}, f); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
}

func TestWriteFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "x.mid")
	err := WriteFile(path, File{Tracks: []Track{{Events: []note.Event{{Pitch: 60, Velocity: 1, Duration: 1}}}}})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil || parse(t, data).tracks[1].events[1].kind != "on" {
		t.Errorf("%v", err)
	}
	if err := WriteFile(path, File{Tempo: -1}); err == nil {
		t.Error("an invalid file was written")
	}
}
