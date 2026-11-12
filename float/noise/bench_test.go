package noise_test

import (
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/float/noise"
)

func BenchmarkNoise1Octave(b *testing.B) {
	genny.Seed(1)
	n := noise.New(0.01, 1)
	b.ReportAllocs()
	for range b.N {
		n.Generate()
	}
}

func BenchmarkNoise4Octaves(b *testing.B) {
	genny.Seed(1)
	n := noise.New(0.01, 4)
	b.ReportAllocs()
	for range b.N {
		n.Generate()
	}
}
