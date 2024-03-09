package shape

import "github.com/almerlucke/genny"

// Shaper shapes a signal
type Shaper interface {
	Shape(float64) float64
}

// Shape takes a []float64 generator and shaper as input, all slice values generated are shaped with the same shaper
type Shape struct {
	gen    genny.Generator[[]float64]
	shaper Shaper
	outVec []float64
}

// New creates a new shape from a []float64 generator and a shaper
func New(gen genny.Generator[[]float64], dim int, shaper Shaper) *Shape {
	return &Shape{
		gen:    gen,
		shaper: shaper,
		outVec: make([]float64, dim),
	}
}

func (s *Shape) Generate() []float64 {
	for index, value := range s.gen.Generate() {
		s.outVec[index] = s.shaper.Shape(value)
	}

	return s.outVec
}

func (s *Shape) Continuous() bool {
	return s.gen.Continuous()
}

func (s *Shape) Done() bool {
	return s.gen.Done()
}

func (s *Shape) Reset() {
	s.gen.Reset()
}
