package tests

import (
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/chord"
	"github.com/almerlucke/genny/euclid"
	"github.com/almerlucke/genny/float/noise"
	"github.com/almerlucke/genny/float/shape"
	"github.com/almerlucke/genny/groove"
	"github.com/almerlucke/genny/mutate"
	"github.com/almerlucke/genny/progression"
	"github.com/almerlucke/genny/rest"
	"github.com/almerlucke/genny/scale"
	"github.com/almerlucke/genny/transform"
)

// the new generators work together: a mutating melody on a euclidean rhythm, chords from a
// progression with smooth voice leading, smooth noise quantized to a scale
func TestMusicPieces(t *testing.T) {
	genny.Seed(42)
	s := scale.New(60, scale.Minor)

	melody := mutate.New([]float64{60, 63, 67, 70}, mutate.Config[float64]{SwapProb: 0.5, ReverseProb: 0.2})
	line := rest.NewGated[float64](melody, euclid.New(8, 5, 0))
	swing := groove.NewSwing(0.2).Humanize(0.02)
	played := 0
	for range 64 {
		off := swing.Generate()
		if n := line.Generate(); n.Played() {
			played++
			if n.Value < 60 || n.Value > 70 || off < -0.02 || off > 0.22 {
				t.Fatal(n, off)
			}
		}
	}
	if played != 40 {
		t.Errorf("played %d of 64 steps, want 40 (5 of 8)", played)
	}

	degrees := progression.New(progression.Pop, 0)
	chords := transform.New(degrees, func(d int) int { return d }) // type check: degrees is a Generator[int]
	leader := chord.NewVoiceLeader(48, 84)
	for range 32 {
		c := leader.Transform(chord.Diatonic(s, chords.Generate(), 3))
		if len(c) != 3 || c[0] < 48 || c[2] > 84 {
			t.Fatal(c)
		}
	}

	pitch := shape.NewSingle(transform.New(noise.New(0.05, 2), func(v float64) float64 { return 60 + v*12 }), s)
	for range 500 {
		if p := pitch.Generate(); s.Quantize(p) != p {
			t.Fatalf("%v is not in the scale", p)
		}
	}
}
