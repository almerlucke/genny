package chaos_test

import (
	"testing"

	"github.com/almerlucke/genny/float/iter"
	"github.com/almerlucke/genny/float/iter/updaters/chaos"
)

func benchIter(b *testing.B, init []float64, u iter.Updater) {
	it := iter.New(init, u)
	b.ReportAllocs()
	for range b.N {
		it.Generate()
	}
}

func BenchmarkVerhulst(b *testing.B) { benchIter(b, []float64{0.3}, chaos.NewVerhulst(3.9)) }
func BenchmarkHenon(b *testing.B)    { benchIter(b, []float64{0.1, 0.1}, chaos.NewHenon(1.4, 0.3)) }
func BenchmarkAronson(b *testing.B)  { benchIter(b, []float64{0.1, 0.1}, chaos.NewAronson(1.5)) }
func BenchmarkLorenz(b *testing.B) {
	benchIter(b, []float64{1, 1, 1}, chaos.NewLorenz(10, 28, 8.0/3, 0.005))
}
func BenchmarkRossler(b *testing.B) {
	benchIter(b, []float64{1, 1, 1}, chaos.NewRossler(0.2, 0.2, 5.7, 0.01))
}
func BenchmarkStandardMap(b *testing.B) { benchIter(b, []float64{0.1, 0.2}, chaos.NewStandardMap(1.2)) }
