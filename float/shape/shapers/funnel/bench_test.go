package funnel_test

import (
	"testing"

	"github.com/almerlucke/genny/float/shape/shapers/funnel"
)

func BenchmarkShape(b *testing.B) {
	s := funnel.New([]float64{0, 2, 4, 5, 7, 9, 11, 12, 14, 16})
	b.ReportAllocs()
	x, sum := -1.0, 0.0
	for range b.N {
		sum += s.Shape(x)
		x += 0.001
		if x > 1 {
			x = -1
		}
	}
	_ = sum
}
