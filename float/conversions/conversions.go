package conversions

import (
	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/convert"
)

type vecConvert struct{}

func (c *vecConvert) Convert(v float64) []float64 {
	return []float64{v}
}

// ToVec creates a conversion from a float64 generator to []float64
func ToVec(g genny.Generator[float64]) *convert.Convert[float64, []float64] {
	return convert.New[float64, []float64](g, &vecConvert{})
}

type floatConvert struct {
	channel int
}

func (c *floatConvert) Convert(v []float64) float64 {
	return v[c.channel]
}

// FromVec creates a conversion from a []float64 generator to float64
func FromVec(g genny.Generator[[]float64], channel int) *convert.Convert[[]float64, float64] {
	return convert.New[[]float64, float64](g, &floatConvert{channel: channel})
}
