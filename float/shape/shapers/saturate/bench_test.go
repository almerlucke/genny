package saturate_test

import (
	"testing"

	"github.com/almerlucke/genny/float/shape/shapers/saturate"
)

func BenchmarkShape(b *testing.B) {
	s := saturate.New(saturate.Tanh, 4)
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
