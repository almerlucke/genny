package onepole_test

import (
	"testing"

	"github.com/almerlucke/genny/float/shape/shapers/onepole"
)

func BenchmarkShape(b *testing.B) {
	s := onepole.New(0.1)
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
