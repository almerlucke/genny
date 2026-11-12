package sieve_test

import (
	"testing"

	"github.com/almerlucke/genny/sieve"
)

func BenchmarkNew(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		sieve.New(7, 3, 0, 1000)
	}
}

func BenchmarkAlgebra(b *testing.B) {
	x, y := sieve.New(3, 0, 0, 1000), sieve.New(5, 2, 0, 1000)
	b.ReportAllocs()
	for range b.N {
		x.Union(y).Intersection(x).Difference(y).Complement()
	}
}

func BenchmarkGen(b *testing.B) {
	g := sieve.NewGen(sieve.New(3, 0, 0, 1000), true)
	b.ReportAllocs()
	for range b.N {
		g.Generate()
	}
}

func BenchmarkContains(b *testing.B) {
	s := sieve.New(3, 0, 0, 1000)
	b.ReportAllocs()
	for i := range b.N {
		s.Contains(i % 1000)
	}
}

func BenchmarkFunnel(b *testing.B) {
	allowed := sieve.Scale(sieve.New(2, 0, 0, 48), 36)
	vals := make([]float64, 64)
	for i := range vals {
		vals[i] = 30 + float64(i)
	}
	b.ReportAllocs()
	for range b.N {
		sieve.Funnel(vals, allowed)
	}
}
