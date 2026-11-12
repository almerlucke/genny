package tendency_test

import (
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/float/tendency"
)

func benchMask() *tendency.Mask {
	return tendency.NewMask(true,
		tendency.Segment{N: 16, A1: 0, A2: 1, Z1: 5, Z2: 9},
		tendency.Segment{N: 16, A1: 3, A2: 4, Z1: 4, Z2: 5})
}

func BenchmarkValue(b *testing.B) {
	genny.Seed(1)
	v := tendency.NewValue(benchMask())
	b.ReportAllocs()
	for range b.N {
		v.Generate()
	}
}

func BenchmarkBetaValue(b *testing.B) {
	genny.Seed(1)
	v := tendency.NewBetaValue(benchMask(), 2, 3)
	b.ReportAllocs()
	for range b.N {
		v.Generate()
	}
}

func BenchmarkChoice(b *testing.B) {
	genny.Seed(1)
	m := tendency.NewMask(true, tendency.Segment{N: 16, A1: 0, A2: 50, Z1: 50, Z2: 100})
	c := tendency.NewChoice(m, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	b.ReportAllocs()
	for range b.N {
		c.Generate()
	}
}

func BenchmarkTwoMasks(b *testing.B) {
	genny.Seed(1)
	g := tendency.NewTwoMasks(benchMask(), tendency.NewMask(true, tendency.Segment{N: 1, A1: 2, A2: 6}), 10)
	b.ReportAllocs()
	for range b.N {
		g.Generate()
	}
}
