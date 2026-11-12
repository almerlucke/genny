package style

import (
	"math"
	"slices"
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/note"
)

// tune is a short melody: pitches and durations in steps
func tune(transpose, speed float64) []note.Event {
	pitches := []float64{60, 62, 64, 60, 60, 62, 64, 60, 64, 65, 67, 64, 65, 67, 67, 69, 67, 65, 64, 60}
	durs := []float64{4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 8, 4, 4, 8, 2, 2, 2, 2, 4, 8}
	var out []note.Event
	t := 0.0
	for i, p := range pitches {
		out = append(out, note.Event{Pitch: p + transpose, Velocity: 0.7, Start: t * speed, Duration: durs[i] * speed})
		t += durs[i]
	}
	return out
}

func TestMelody(t *testing.T) {
	events := []note.Event{
		{Pitch: 64, Start: 2, Duration: 1},
		{Pitch: 60, Start: 0, Duration: 1}, // a chord: the top note counts
		{Pitch: 67, Start: 0, Duration: 2},
		{Pitch: 62, Start: 3, Duration: 0.5},
	}
	got := Melody(events)
	want := []Note{{67, 2}, {64, 1}, {62, 0.5}} // the rest after 64 belongs to it, the last keeps its length
	if !slices.Equal(got, want) {
		t.Errorf("%v, want %v", got, want)
	}
	if len(Melody(nil)) != 0 {
		t.Error("empty melody")
	}
}

func TestEventsAndLength(t *testing.T) {
	phrase := []Note{{60, 1}, {62, 2.5}}
	ev := Events(phrase, 4, 0.5)
	if len(ev) != 2 || ev[0].Start != 4 || ev[1].Start != 5 || ev[1].Duration != 2.5 || ev[1].Pitch != 62 || ev[0].Velocity != 0.5 {
		t.Errorf("%+v", ev)
	}
	if Length(phrase) != 3.5 {
		t.Error(Length(phrase))
	}
	// what Events makes comes back as the same melody
	if !slices.Equal(Melody(ev), phrase) {
		t.Errorf("%v", Melody(ev))
	}
}

func TestAbsoluteStyleOnlyUsesTheExample(t *testing.T) {
	genny.Seed(1)
	s := Learn(tune(0, 1), 2)
	pitches, durs := map[float64]bool{}, map[float64]bool{}
	for _, n := range Melody(tune(0, 1)) {
		pitches[n.Pitch], durs[n.Duration] = true, true
	}
	for range 200 {
		n := s.Generate()
		if !pitches[n.Pitch] || !durs[n.Duration] {
			t.Fatalf("%+v is not in the example", n)
		}
	}
}

func TestIntervalsAndRatiosTranspose(t *testing.T) {
	// the same melody a fourth higher and at half the speed has the same style: with the same seed
	// the new music is the same, a fourth higher and twice as long
	genny.Seed(4)
	a := Learn(tune(0, 1), 2, Intervals(), Ratios()).Phrase(60)
	genny.Seed(4)
	b := Learn(tune(5, 2), 2, Intervals(), Ratios()).Phrase(60)
	for i := range a {
		if math.Abs(b[i].Pitch-a[i].Pitch-5) > 1e-9 || math.Abs(b[i].Duration-2*a[i].Duration) > 1e-9 {
			t.Fatalf("note %d: %+v and %+v", i, a[i], b[i])
		}
	}

	// the same style transposed after learning
	genny.Seed(4)
	c := Learn(tune(0, 1), 2, Intervals(), Ratios()).Transpose(5).Scale(2).Phrase(60)
	if !slices.Equal(b, c) {
		t.Errorf("%v\n%v", b[:5], c[:5])
	}
}

func TestIntervalsStayInRange(t *testing.T) {
	genny.Seed(2)
	s := Learn(tune(0, 1), 1, Intervals(), Ratios())
	for range 2000 {
		n := s.Generate()
		if n.Pitch < 48 || n.Pitch > 81 || n.Duration < 2-1e-9 || n.Duration > 8+1e-9 {
			t.Fatalf("%+v out of range", n)
		}
	}
}

func TestSeededAndReset(t *testing.T) {
	genny.Seed(9)
	s := Learn(tune(0, 1), 2, Intervals())
	a := s.Phrase(30)
	genny.Seed(9)
	s.Reset()
	if b := s.Phrase(30); !slices.Equal(a, b) {
		t.Error("not repeatable")
	}
	genny.Seed(10)
	s.Reset()
	if b := s.Phrase(30); slices.Equal(a, b) {
		t.Error("another seed gives the same")
	}
	if !s.Continuous() || s.Done() {
		t.Error("a style never ends")
	}
}

func TestFill(t *testing.T) {
	genny.Seed(3)
	s := Learn(tune(0, 1), 2, Intervals(), Ratios())
	for _, length := range []float64{1, 10, 33.5, 200} {
		if got := Length(s.Fill(length)); math.Abs(got-length) > 1e-9 {
			t.Errorf("length %v, want %v", got, length)
		}
	}
	if len(s.Fill(0)) != 0 {
		t.Error("filled nothing")
	}
}

func TestVelocity(t *testing.T) {
	if v := Learn(tune(0, 1), 1).Velocity; math.Abs(v-0.7) > 1e-9 {
		t.Error(v)
	}
}

func TestInvalid(t *testing.T) {
	for name, f := range map[string]func(){
		"order":  func() { Learn(tune(0, 1), 0) },
		"short":  func() { Learn(tune(0, 1)[:3], 3) },
		"empty":  func() { Learn(nil, 1) },
		"factor": func() { Learn(tune(0, 1), 1).Scale(0) },
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
