package phasedist_test

import (
	"testing"

	"github.com/almerlucke/genny/float/shape/shapers/phasedist"
)

func BenchmarkShape(b *testing.B) {
	s := phasedist.NewResonant(0.5, 3)
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
