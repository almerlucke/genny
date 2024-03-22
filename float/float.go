package float

import (
	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/conv"
)

type FrameGenerator interface {
	genny.Generator[[]float64]
	Dimensions() int
}

type ToFrameConvert struct {
	gen genny.Generator[float64]
	out [1]float64
}

func (c *ToFrameConvert) Generate() []float64 {
	c.out[0] = c.gen.Generate()
	return c.out[:]
}

func (c *ToFrameConvert) Dimensions() int {
	return 1
}

func (c *ToFrameConvert) Continuous() bool {
	return c.gen.Continuous()
}

func (c *ToFrameConvert) Done() bool {
	return c.gen.Done()
}

func (c *ToFrameConvert) Reset() {
	c.gen.Reset()
}

// ToFrame creates a converter from a Generator[float64] to a FrameGenerator
func ToFrame(gen genny.Generator[float64]) *ToFrameConvert {
	return &ToFrameConvert{gen: gen}
}

type floatConvert struct {
	dimension int
}

func (c *floatConvert) Convert(v []float64) float64 {
	return v[c.dimension]
}

// FromFrame creates a conversion from a FrameGenerator to Generator[float64]
func FromFrame(g FrameGenerator, dimension int) *conv.Convert[[]float64, float64] {
	return conv.New[[]float64, float64](g, &floatConvert{dimension: dimension})
}
