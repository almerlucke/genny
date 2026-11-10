// Package or provides Or, a generator that each cycle chooses one of its generators, at random or one
// by one.
package or

import (
	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/bucket"
)

// Mode is the way Or picks a generator
type Mode bucket.Mode

const (
	Random  = Mode(bucket.Random)
	Indexed = Mode(bucket.Indexed)
)

// Or chooses one of the generators each cycle randomly or in bucket random mode (one by one).
// An Or without generators is done at once and generates the zero value
type Or[T any] struct {
	bucket  *bucket.Bucket[genny.Generator[T]]
	current genny.Generator[T]
	done    bool
}

// New creates an Or that runs through its generators once (Indexed) or picks at random (Random)
func New[T any](mode Mode, gens ...genny.Generator[T]) *Or[T] {
	o := &Or[T]{
		bucket: bucket.New(gens...).SetMode(bucket.Mode(mode)),
	}

	o.pick()

	return o
}

// NewLoop creates an Or that keeps going after all generators are used
func NewLoop[T any](mode Mode, gens ...genny.Generator[T]) *Or[T] {
	o := &Or[T]{
		bucket: bucket.New(gens...).Loop(true).SetMode(bucket.Mode(mode)),
	}

	o.pick()

	return o
}

// pick takes the next generator from the bucket, with an empty bucket the Or is done
func (o *Or[T]) pick() {
	if len(o.bucket.Values()) == 0 {
		o.current = nil
		o.done = true
		return
	}
	o.done = false
	o.current = o.bucket.Generate()
}

func (o *Or[T]) Generate() (value T) {
	if o.current == nil {
		return
	}

	gotoNext := false

	value = o.current.Generate()

	if o.current.Continuous() {
		gotoNext = true
	} else if o.current.Done() {
		gotoNext = true
		o.current.Reset()
	}

	if gotoNext {
		if !o.bucket.Continuous() && o.bucket.Done() {
			o.done = true
		} else {
			o.current = o.bucket.Generate()
		}
	}

	return
}

func (o *Or[T]) Continuous() bool {
	return o.bucket.Continuous()
}

func (o *Or[T]) Reset() {
	o.bucket.Reset()

	for _, g := range o.bucket.Values() {
		g.Reset()
	}

	o.pick()
}

func (o *Or[T]) Done() bool {
	return o.done
}
