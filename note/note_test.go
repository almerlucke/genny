package note

import (
	"slices"
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/constant"
	"github.com/almerlucke/genny/euclid"
	"github.com/almerlucke/genny/ratchet"
	"github.com/almerlucke/genny/rest"
	"github.com/almerlucke/genny/sequence"
)

func gated(pitches []float64, steps, pulses int) genny.Generator[rest.Optional[float64]] {
	return rest.NewGated[float64](sequence.NewLoop(pitches...), euclid.New(steps, pulses, 0))
}

func TestRenderDefaults(t *testing.T) {
	ev := Voice{Pitch: Always[float64](sequence.NewLoop(60.0, 62.0))}.Render(4)
	want := []Event{{60, 0.8, 0, 1}, {62, 0.8, 1, 1}, {60, 0.8, 2, 1}, {62, 0.8, 3, 1}}
	if !slices.Equal(ev, want) {
		t.Errorf("%v", ev)
	}
}

func TestRestsKeepEverythingAligned(t *testing.T) {
	// 3 pulses in 8 steps: steps 0, 3, 6. The pitch is consumed on rests, so are velocity and length
	v := Voice{
		Pitch:    gated([]float64{60, 61, 62, 63, 64, 65, 66, 67}, 8, 3),
		Velocity: sequence.NewLoop(0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8),
		Length:   sequence.NewLoop(1.0, 2, 3, 4, 5, 6, 7, 8),
	}
	got := v.Render(8)
	want := []Event{{60, 0.1, 0, 1}, {63, 0.4, 3, 4}, {66, 0.7, 6, 7}}
	if !slices.Equal(got, want) {
		t.Errorf("%v, want %v", got, want)
	}
}

func TestOffsetsAndRatchets(t *testing.T) {
	v := Voice{
		Pitch:   Always[float64](constant.New(60.0)),
		Length:  constant.New(1.0),
		Offset:  sequence.NewLoop(-0.5, 0.25),
		Ratchet: ratchet.NewPattern(1, 2),
	}
	got := v.Render(2)
	want := []Event{
		{60, 0.8, 0, 1},      // -0.5 cannot start before 0
		{60, 0.8, 1.25, 0.5}, // two repeats at 1.25 and 1.75, half the length each
		{60, 0.8, 1.75, 0.5},
	}
	if !slices.Equal(got, want) {
		t.Errorf("%v, want %v", got, want)
	}
}

func TestChordsAndOrder(t *testing.T) {
	c1, c2 := []float64{60, 64, 67}, []float64{62, 65, 69}
	v := Voice{
		Chord:    rest.NewProbability[[]float64](sequence.NewLoop(c1, c2), 1),
		Velocity: constant.New(2.0), // clamped to 1
	}
	got := v.Render(2)
	if len(got) != 6 || got[0].Start != 0 || got[3].Start != 1 || got[0].Velocity != 1 {
		t.Errorf("%v", got)
	}
	if !slices.IsSortedFunc(got, func(a, b Event) int {
		if a.Start < b.Start {
			return -1
		} else if a.Start > b.Start {
			return 1
		}
		return 0
	}) {
		t.Error("not sorted by start")
	}
}

func TestVoiceNeedsExactlyOneSource(t *testing.T) {
	for name, v := range map[string]Voice{
		"none": {},
		"both": {
			Pitch: Always[float64](constant.New(60.0)),
			Chord: Always[[]float64](constant.New([]float64{60})),
		},
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s should panic", name)
				}
			}()
			v.Render(1)
		}()
	}
}
