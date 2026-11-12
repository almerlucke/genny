package line

import (
	"slices"
	"testing"

	"github.com/almerlucke/genny/note"
	"github.com/almerlucke/genny/scale"
	"github.com/almerlucke/genny/sequence"
)

func mk(pitches ...float64) []note.Event {
	var out []note.Event
	for i, p := range pitches {
		out = append(out, note.Event{Pitch: p, Velocity: 0.5, Start: float64(i) * 1.5, Duration: 1 + float64(i%2)*0.5})
	}
	return out
}

func pitches(ev []note.Event) []float64 {
	var out []float64
	for _, e := range ev {
		out = append(out, e.Pitch)
	}
	return out
}

func mustPanic(t *testing.T, name string, f func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Errorf("%s: expected a panic", name)
		}
	}()
	f()
}

func TestNoMutation(t *testing.T) {
	in := mk(60, 62, 64, 65)
	orig := slices.Clone(in)
	c := scale.New(60, scale.Major)
	Transpose(in, 3)
	Stretch(in, 2)
	Invert(in, 62, c)
	Invert(in, 62, nil)
	Retrograde(in)
	Shift(in, 4)
	Slice(in, 1, 3)
	Loop(in, 2, 8)
	Quantize(in, c)
	Canon(in, 4, 7, 3)
	Merge(in, in)
	if !slices.Equal(in, orig) {
		t.Error("input was changed")
	}
}

func TestTransposeStretch(t *testing.T) {
	got := Transpose(mk(60, 62), -12)
	if !slices.Equal(pitches(got), []float64{48, 50}) {
		t.Error(got)
	}
	s := Stretch(mk(60, 62), 2)
	if s[1].Start != 3 || s[1].Duration != 3 || s[0].Duration != 2 {
		t.Error(s)
	}
	d := Diminish(mk(60, 62), 3)
	if d[1].Start != 0.5 {
		t.Error(d)
	}
	mustPanic(t, "stretch 0", func() { Stretch(nil, 0) })
	mustPanic(t, "diminish -1", func() { Diminish(nil, -1) })
}

func TestInvert(t *testing.T) {
	c := scale.New(60, scale.Major)
	tests := []struct {
		name string
		axis float64
		s    *scale.Scale
		in   []float64
		want []float64
	}{
		{"chromatic", 60, nil, []float64{60, 62, 67}, []float64{60, 58, 53}},
		{"diatonic", 64, c, []float64{60, 62, 64, 67}, []float64{67, 65, 64, 60}},
		{"diatonic stays in scale", 60, c, []float64{62, 64, 65}, []float64{59, 57, 55}},
	}
	for _, tt := range tests {
		if got := pitches(Invert(mk(tt.in...), tt.axis, tt.s)); !slices.Equal(got, tt.want) {
			t.Errorf("%s: got %v want %v", tt.name, got, tt.want)
		}
	}
	// inverting twice is the identity
	in := mk(60, 62, 65, 69)
	if got := Invert(Invert(in, 63, nil), 63, nil); !slices.Equal(got, in) {
		t.Error("chromatic invert twice")
	}
}

func TestRetrograde(t *testing.T) {
	in := mk(60, 62, 64, 65, 67)
	r := Retrograde(in)
	if !slices.Equal(pitches(r), []float64{67, 65, 64, 62, 60}) {
		t.Error(pitches(r))
	}
	if !slices.Equal(Retrograde(r), in) {
		t.Error("retrograde twice is not the identity")
	}
	// the line keeps its length
	if r[0].Start != 0 {
		t.Error(r[0])
	}
	if Retrograde(nil) != nil {
		t.Error("empty")
	}
	// a line that does not start at 0
	off := Shift(in, 10)
	if !slices.Equal(Retrograde(Retrograde(off)), off) {
		t.Error("offset line")
	}
}

func TestShiftSliceLoopMerge(t *testing.T) {
	in := mk(60, 62, 64)
	if s := Shift(in, -2); s[0].Start != 0 || s[2].Start != 1 {
		t.Error(s)
	}
	if got := Slice(in, 1, 3.1); len(got) != 2 || got[0].Start != 0.5 {
		t.Error(got)
	}
	if len(Slice(in, 10, 20)) != 0 {
		t.Error("slice outside")
	}
	lp := Loop(in, 3, 10)
	if len(lp) != 9 || lp[8].Start != 23 {
		t.Error(lp)
	}
	if len(Loop(in, 0, 10)) != 0 || len(Loop(nil, 4, 2)) != 0 {
		t.Error("empty loop")
	}
	mustPanic(t, "loop length 0", func() { Loop(in, 2, 0) })
	m := Merge(in, Shift(mk(70, 72), 0.5))
	if len(m) != 5 || !slices.IsSortedFunc(m, func(a, b note.Event) int {
		if a.Start < b.Start {
			return -1
		} else if a.Start > b.Start {
			return 1
		}
		return 0
	}) {
		t.Error(m)
	}
	if Merge() != nil {
		t.Error("empty merge")
	}
}

func TestQuantize(t *testing.T) {
	c := scale.New(60, scale.Major)
	got := pitches(Quantize(mk(60.9, 61.4, 63.2), c))
	if !slices.Equal(got, []float64{62, 62, 64}) && !slices.Equal(got, []float64{60, 62, 64}) {
		// 60.9 is closest to 60 (0.9) or 62 (1.1): 60
		t.Error(got)
	}
	if got[0] != 60 {
		t.Error(got)
	}
}

func TestCanon(t *testing.T) {
	in := mk(60, 62, 64)
	for voices := 1; voices <= 4; voices++ {
		c := Canon(in, 8, 7, voices)
		if len(c) != voices {
			t.Fatalf("voices %d got %d", voices, len(c))
		}
		for i, v := range c {
			if len(v) != len(in) {
				t.Errorf("voice %d has %d notes", i, len(v))
			}
			if v[0].Start != in[0].Start+float64(i)*8 || v[0].Pitch != 60+float64(i)*7 {
				t.Errorf("voice %d starts wrong: %v", i, v[0])
			}
		}
	}
	mustPanic(t, "no voices", func() { Canon(in, 1, 1, 0) })
}

func TestThin(t *testing.T) {
	in := mk(60, 62, 64, 65)
	got := Thin(in, sequence.NewLoop(true, false))
	if !slices.Equal(pitches(got), []float64{60, 64}) {
		t.Error(got)
	}
	if len(Thin(nil, sequence.NewLoop(true))) != 0 {
		t.Error("empty")
	}
	mustPanic(t, "nil gate", func() { Thin(in, nil) })
}
