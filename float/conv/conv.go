package conv

import (
	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/conv"
)

type vecConvert struct{}

func (c *vecConvert) Convert(v float64) []float64 {
	return []float64{v}
}

// ToVec creates a conversion from a float64 generator to []float64
func ToVec(g genny.Generator[float64]) *conv.Convert[float64, []float64] {
	return conv.New[float64, []float64](g, &vecConvert{})
}

type floatConvert struct {
	channel int
}

func (c *floatConvert) Convert(v []float64) float64 {
	return v[c.channel]
}

// FromVec creates a conversion from a []float64 generator to float64
func FromVec(g genny.Generator[[]float64], channel int) *conv.Convert[[]float64, float64] {
	return conv.New[[]float64, float64](g, &floatConvert{channel: channel})
}
