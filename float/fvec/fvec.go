package fvec

import (
	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/convert"
)

type vecConvert struct{}

func (t *vecConvert) Convert(v float64) []float64 {
	return []float64{v}
}

// New creates a cast for a float64 output generator to float64 vec
func New(g genny.Generator[float64]) *convert.Convert[float64, []float64] {
	return convert.New[float64, []float64](g, &vecConvert{})
}
