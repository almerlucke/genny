package fvec

import (
	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/cast"
)

type vecCaster struct{}

func (t *vecCaster) Cast(v float64) []float64 {
	return []float64{v}
}

// New creates a cast for a float64 output generator to float64 vec
func New(g genny.Generator[float64]) *cast.Cast[float64, []float64] {
	return cast.New[float64, []float64](g, &vecCaster{})
}
