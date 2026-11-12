package dist_test

import (
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/float/dist"
)

func BenchmarkGaussian(b *testing.B) {
	genny.Seed(1)
	g := dist.NewGaussian(0, 1)
	b.ReportAllocs()
	for range b.N {
		g.Generate()
	}
}

func BenchmarkExponential(b *testing.B) {
	genny.Seed(1)
	g := dist.NewExponential(1)
	b.ReportAllocs()
	for range b.N {
		g.Generate()
	}
}

func BenchmarkBeta(b *testing.B) {
	genny.Seed(1)
	g := dist.NewBeta(2, 3)
	b.ReportAllocs()
	for range b.N {
		g.Generate()
	}
}

func BenchmarkCauchy(b *testing.B) {
	genny.Seed(1)
	g := dist.NewCauchy(0, 1)
	b.ReportAllocs()
	for range b.N {
		g.Generate()
	}
}
