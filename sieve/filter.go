package sieve

import (
	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/rest"
)

// NewFilter keeps the values of gen at the steps that are members of the sieve (the others are
// rests), like the sieve filter of AC Toolbox. After length steps the sieve starts again at step 0.
// It panics if length is less than 1
func NewFilter[T any](gen genny.Generator[T], s Sieve, length int) *rest.Gated[T] {
	return rest.NewGated(gen, genny.Generator[bool](NewGate(s, length)))
}
