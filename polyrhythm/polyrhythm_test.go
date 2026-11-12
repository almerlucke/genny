package polyrhythm

import (
	"math"

	"github.com/almerlucke/genny/note"
	"reflect"
	"testing"
)

func TestFrac(t *testing.T) {
	tests := []struct {
		name string
		got  Frac
		want Frac
	}{
		{"normalize", F(6, -4), F(-3, 2)},
		{"add", F(1, 3).Add(F(1, 6)), F(1, 2)},
		{"sub", F(1, 2).Sub(F(2, 3)), F(-1, 6)},
		{"mul", F(2, 3).Mul(F(9, 4)), F(3, 2)},
		{"div", F(3, 2).Div(F(3, 4)), Int(2)},
		{"zero value", Frac{}.Add(F(1, 5)), F(1, 5)},
		{"gcd", GCD(F(2, 3), F(4, 5)), F(2, 15)},
		{"lcm", LCM(F(2, 3), F(4, 5)), F(4, 1)},
		{"lcm halves", LCM(F(7, 2), Int(4)), Int(28)},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, tt.got, tt.want)
		}
	}
	if F(-7, 2).Floor() != -4 || F(-7, 2).Ceil() != -3 || F(7, 2).Floor() != 3 || F(7, 2).Ceil() != 4 || Int(3).Ceil() != 3 {
		t.Error("floor/ceil")
	}
	if F(1, 3).Cmp(F(2, 6)) != 0 || F(1, 3).Cmp(F(1, 2)) != -1 || F(2, 3).Sign() != 1 {
		t.Error("cmp")
	}
	if F(7, 4).String() != "7/4" || Int(2).String() != "2" {
		t.Error("string")
	}
}

func TestFracPanics(t *testing.T) {
	for name, f := range map[string]func(){
		"den0":     func() { F(1, 0) },
		"div0":     func() { Int(1).Div(Frac{}) },
		"overflow": func() { Int(math.MaxInt64).Add(Int(1)) },
		"layer":    func() { NewLayer(0, Int(1)) },
		"span":     func() { NewLayer(3, Frac{}) },
		"additive": func() { Additive(Int(1)) },
		"group":    func() { Additive(Int(1), 3, 0) },
		"equate":   func() { Equate(Int(1), Frac{}) },
		"plan":     func() { NewPlan(Frac{}) },
		"bars":     func() { NewPlan(Int(120), Section{Bars: 0}) },
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

func TestCycleAndRealign(t *testing.T) {
	tests := []struct {
		name   string
		layers []Layer
		cycle  Frac
		meet   []Frac
	}{
		{"3:2", []Layer{NewLayer(3, Int(2)), NewLayer(2, Int(2))}, Int(2), []Frac{Int(0)}},
		{"5:4", []Layer{NewLayer(5, Int(4)), NewLayer(4, Int(4))}, Int(4), []Frac{Int(0)}},
		{"7:4", []Layer{NewLayer(7, Int(4)), NewLayer(4, Int(4))}, Int(4), []Frac{Int(0)}},
		{"cross 3 beats in 4", []Layer{Every(Int(3)), Every(Int(4))}, Int(12), []Frac{Int(0)}},
		{"7/2 vs 4", []Layer{Every(F(7, 2)), Every(Int(4))}, Int(28), []Frac{Int(0)}},
		{"6:4 meet twice", []Layer{NewLayer(6, Int(4)), NewLayer(4, Int(4))}, Int(2), []Frac{Int(0)}},
		{"nested 3:(2:5)", []Layer{NewLayer(2, Int(5)).Nest(3)}, F(5, 6), []Frac{Int(0)}},
		{"none", nil, Frac{}, nil},
	}
	for _, tt := range tests {
		if got := Cycle(tt.layers...); got != tt.cycle {
			t.Errorf("%s: cycle %v, want %v", tt.name, got, tt.cycle)
		}
		if got := Realign(tt.layers...); !reflect.DeepEqual(got, tt.meet) {
			t.Errorf("%s: realign %v, want %v", tt.name, got, tt.meet)
		}
	}
}

func TestRealignOffset(t *testing.T) {
	a, b := NewLayer(2, Int(2)), NewLayer(3, Int(3)) // periods 1 and 1
	b.Offset = F(1, 2)
	if got := Realign(a, b); len(got) != 0 {
		t.Errorf("offset layers should never meet, got %v", got)
	}
	b.Offset = Int(1) // a whole period later: same grid
	if got := Realign(a, b); !reflect.DeepEqual(got, []Frac{Int(0)}) {
		t.Errorf("got %v", got)
	}
}

func TestNestedPeriod(t *testing.T) {
	l := NewLayer(2, Int(5)).Nest(3)
	if l.Period() != F(5, 6) {
		t.Errorf("period %v", l.Period())
	}
	if p := Tuplet(5, 4, F(1, 2)).Period(); p != F(2, 5) {
		t.Errorf("tuplet period %v", p)
	}
}

func TestGrid(t *testing.T) {
	g := NewGrid(NewLayer(3, Int(2)), NewLayer(2, Int(2)))
	if g.Unit != F(1, 3) {
		t.Fatalf("unit %v", g.Unit)
	}
	// periods 2/3 and 1: gcd 1/3, cycle 2 -> 6 cells
	want := "x.x.x.\nx..x.."
	if g.Steps != 6 || g.String() != want {
		t.Errorf("steps %d\n%s", g.Steps, g)
	}
	g = NewGrid(NewLayer(5, Int(4)), NewLayer(4, Int(4)))
	// periods 4/5 and 1: gcd 1/5, cycle 4 -> 20 cells
	if g.Unit != F(1, 5) || g.Steps != 20 {
		t.Errorf("5:4 grid %v %d", g.Unit, g.Steps)
	}
	if n := countTrue(g.Cells[0]); n != 5 {
		t.Errorf("5 pulses expected, got %d", n)
	}
	// pattern thins the grid
	l := NewLayer(4, Int(4)).WithEuclid(4, 3, 0)
	g = NewGrid(l)
	if g.Steps != 4 || g.String() != "xxx." {
		t.Errorf("pattern grid %q", g)
	}
	if got := NewGrid(); got.Steps != 0 {
		t.Error("empty grid")
	}
	if got := NewGrid(NewLayer(1000003, Int(1)), NewLayer(999983, Int(1))); got.Steps != 0 {
		t.Error("huge grid should be empty")
	}
}

func countTrue(b []bool) int {
	n := 0
	for _, v := range b {
		if v {
			n++
		}
	}
	return n
}

func TestEvents(t *testing.T) {
	l := NewLayer(3, Int(2))
	l.Pitch = 64
	ev := l.Events(Int(0), Int(4), 12)
	if len(ev) != 6 {
		t.Fatalf("got %d events", len(ev))
	}
	// period 2/3 beat, 12 steps per beat: 8 steps apart, exactly
	for i, e := range ev {
		if e.Start != float64(8*i) || e.Duration != 4 || e.Pitch != 64 {
			t.Errorf("event %d: %+v", i, e)
		}
	}
	// window not starting at 0, exclusive end
	ev = l.Events(F(1, 2), F(4, 3), 3)
	if len(ev) != 1 || ev[0].Start != 2 {
		t.Errorf("window: %+v", ev)
	}
	if got := (Layer{}).Events(Int(0), Int(4), 4); got != nil {
		t.Error("invalid layer")
	}
	// no drift: a thousand bars of 7:4 land on exact steps
	l7 := NewLayer(7, Int(4))
	ev = l7.Events(Int(4000), Int(4004), 7)
	if ev[0].Start != 28000 || ev[3].Start != 28000+12 {
		t.Errorf("drift: %v", ev[3].Start)
	}
}

func TestAccents(t *testing.T) {
	l := NewLayer(8, Int(4)).WithEuclidAccent(8, 3, 0, 0.2)
	ev := l.Events(Int(0), Int(4), 4)
	var hi []int
	for i, e := range ev {
		if e.Velocity == 110 {
			hi = append(hi, i)
		} else if e.Velocity != 60+50*0.2 {
			t.Errorf("soft velocity %v", e.Velocity)
		}
	}
	if !reflect.DeepEqual(hi, []int{0, 3, 6}) {
		t.Errorf("accented %v", hi)
	}
	b := NewLayer(4, Int(4)).WithBarlow(0, 0, 2, 2)
	if b.Accent[0] != 1 || b.Accent[2] <= b.Accent[1] {
		t.Errorf("barlow %v", b.Accent)
	}
	th := NewLayer(4, Int(4)).WithBarlow(0, 2, 2, 2) // only the 2 most important pulses
	if got := len(th.Events(Int(0), Int(4), 4)); got != 2 {
		t.Errorf("thinned to %d", got)
	}
}

func TestAdditive(t *testing.T) {
	m := Additive(F(1, 2), 3, 3, 2)
	if m.Length() != Int(4) || m.Units() != 8 {
		t.Errorf("length %v", m.Length())
	}
	if !reflect.DeepEqual(m.Accents(), []float64{1, .3, .3, .7, .3, .3, .7, .3}) {
		t.Errorf("accents %v", m.Accents())
	}
	if n, d, ok := m.TimeSignature(); n != 8 || d != 8 || !ok {
		t.Errorf("time signature %d/%d %v", n, d, ok)
	}
	if _, _, ok := Additive(F(2, 3), 3).TimeSignature(); ok {
		t.Error("triplet unit has no time signature")
	}
	if on := Additive(F(1, 2), 2, 2, 3).Onsets(); len(on) != 7 || on[6] != Int(3) {
		t.Errorf("onsets %v", on)
	}
	// 2+2+3 eighths = 7/2 beats against a 4 beat bar: realign after 28 beats
	if c := Cycle(Additive(F(1, 2), 2, 2, 3).Layer(), Every(Int(4))); c != Int(28) {
		t.Errorf("cycle %v", c)
	}
}

func TestModulation(t *testing.T) {
	tests := []struct {
		name string
		m    Modulation
		from Frac
		want Frac
	}{
		{"3:2", Ratio(3, 2), Int(120), Int(180)},
		{"dotted quarter = quarter", Equate(F(3, 2), Int(1)), Int(120), Int(80)},
		{"quarter = dotted quarter", Equate(Int(1), F(3, 2)), Int(120), Int(180)},
		{"quintuplet eighth = eighth", Equate(F(2, 5), F(1, 2)), Int(100), Int(125)},
		{"eighth = quintuplet eighth", Equate(F(1, 2), F(2, 5)), Int(100), Int(80)},
		{"triplet quarter = quarter", Equate(F(2, 3), Int(1)), Int(90), Int(135)},
		{"non integer", Ratio(7, 4), Int(90), F(315, 2)},
		{"non integer start", Ratio(5, 4), F(126, 5), F(63, 2)},
		{"identity", Modulation{}, Int(77), Int(77)},
	}
	for _, tt := range tests {
		got := tt.m.Tempo(tt.from)
		if got != tt.want {
			t.Errorf("%s: %v, want %v", tt.name, got, tt.want)
		}
		// round trip
		if back := tt.m.Inverse().Tempo(got); back != tt.from {
			t.Errorf("%s: round trip %v, want %v", tt.name, back, tt.from)
		}
		// seconds kept: d old beats at old tempo == Carry(d) new beats at new tempo
		d := F(3, 7)
		if tt.m.Carry(d).Mul(Int(60)).Div(got) != d.Mul(Int(60)).Div(tt.from) {
			t.Errorf("%s: carried duration differs in seconds", tt.name)
		}
	}
	if Equate(F(3, 2), Int(1)).Carry(F(3, 2)) != Int(1) {
		t.Error("dotted quarter should carry to exactly 1 beat")
	}
	chain := Ratio(3, 2).Then(Ratio(4, 5)).Then(Ratio(5, 6))
	if chain.Tempo(Int(120)) != Int(120) {
		t.Errorf("chain %v", chain.Tempo(Int(120)))
	}
}

func TestCarryEvents(t *testing.T) {
	in := []note.Event{{Pitch: 60, Velocity: 90, Start: 4, Duration: 2}}
	got := Equate(F(3, 2), Int(1)).CarryEvents(in) // slower tempo: fewer beats for the same seconds
	if got[0].Start != 8.0/3 || got[0].Duration != 4.0/3 || in[0].Start != 4 {
		t.Errorf("%+v %+v", got, in)
	}
	if got := (Modulation{}).CarryEvents(in); got[0] != in[0] {
		t.Error("identity")
	}
}

func testPlan() Plan {
	return NewPlan(Int(90),
		Section{Name: "A", Bars: 4, Num: 2, Den: 4},
		Section{Name: "B", Bars: 2, Into: Equate(F(2, 3), Int(1))},
		Section{Name: "C", Bars: 3, Num: 7, Den: 8, Into: Equate(F(1, 2), F(2, 5))},
	)
}

func TestPlan(t *testing.T) {
	p := testPlan()
	wantTempo := []Frac{Int(90), Int(135), Int(108)}
	wantStart := []Frac{Int(0), Int(8), Int(16)}
	wantBar := []int{0, 4, 6}
	for i, s := range p {
		if s.Tempo != wantTempo[i] || s.Start != wantStart[i] || s.FirstBar != wantBar[i] {
			t.Errorf("section %d: %+v", i, s)
		}
	}
	// A: 8 beats at 90 = 16/3 s; B: 8 beats at 135 = 32/9 s; C: 3 bars of 3.5 beats at 108
	if p[1].Time != F(16, 3) || p[2].Time != F(16, 3).Add(F(32, 9)) {
		t.Errorf("times %v %v", p[1].Time, p[2].Time)
	}
	if p.Beats() != F(53, 2) {
		t.Errorf("beats %v", p.Beats())
	}
	if p.Seconds() != p[2].Time.Add(F(21, 2).MulInt(60).Div(Int(108))) {
		t.Errorf("seconds %v", p.Seconds())
	}
}

func TestPlanRoundTrip(t *testing.T) {
	p := testPlan()
	for _, b := range []Frac{Int(0), F(1, 3), Int(8), F(79, 7), Int(16), F(53, 2), Int(40)} {
		if got := p.BeatsAt(p.SecondsAt(b)); got != b {
			t.Errorf("beats %v -> %v", b, got)
		}
	}
	for _, s := range []Frac{Int(0), F(5, 3), Int(9), F(1234, 100)} {
		if got := p.SecondsAt(p.BeatsAt(s)); got != s {
			t.Errorf("seconds %v -> %v", s, got)
		}
	}
	// a note of 2 seconds started 1 beat before the end of A lasts longer in beats once it
	// crosses into B (135 bpm): 1 beat of A (2/3 s), then 4/3 s = 3 beats of B
	if got := p.DurationAt(Int(7), Int(2)); got != Int(4) {
		t.Errorf("duration %v", got)
	}
	// within a section it is just seconds*bpm/60
	if got := p.DurationAt(Int(2), Int(2)); got != Int(3) {
		t.Errorf("duration %v", got)
	}
}

func TestPlanMidi(t *testing.T) {
	p := testPlan()
	tc := p.TempoChanges(4)
	if len(tc) != 2 || tc[0].Step != 32 || tc[0].BPM != 135 || tc[1].Step != 64 || tc[1].BPM != 108 {
		t.Errorf("tempo changes %+v", tc)
	}
	if p.Start() != 90 {
		t.Error("start")
	}
	ts := p.TimeSignatures(4)
	if len(ts) != 3 || ts[0].Num != 2 || ts[0].Den != 4 || ts[1].Num != 4 || ts[1].Step != 32 || ts[2].Den != 8 || ts[2].Step != 64 {
		t.Errorf("time signatures %+v", ts)
	}
	same := NewPlan(Int(100), Section{Bars: 1}, Section{Bars: 1}, Section{Bars: 1, Into: Ratio(1, 1)})
	if len(same.TempoChanges(4)) != 0 || len(same.TimeSignatures(4)) != 1 {
		t.Error("unchanged tempo and meter should give no entries")
	}
	if len(NewPlan(Int(100)).TempoChanges(4)) != 0 || NewPlan(Int(100)).Start() != 120 || NewPlan(Int(100)).Seconds() != (Frac{}) {
		t.Error("empty plan")
	}
}
