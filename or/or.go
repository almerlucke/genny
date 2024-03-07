package or

import (
	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/bucket"
)

type Mode bucket.Mode

const (
	Random  = Mode(bucket.Random)
	Indexed = Mode(bucket.Indexed)
)

// Or chooses one of the generators each cycle randomly or in bucket random mode (one by one)
type Or[T any] struct {
	bucket  *bucket.Bucket[genny.Generator[T]]
	current genny.Generator[T]
	done    bool
}

func New[T any](mode Mode, gens ...genny.Generator[T]) *Or[T] {
	o := &Or[T]{
		bucket: bucket.New(bucket.Mode(mode), gens...),
	}

	o.current = o.bucket.Generate()

	return o
}

func NewContinuous[T any](mode Mode, gens ...genny.Generator[T]) *Or[T] {
	o := &Or[T]{
		bucket: bucket.NewContinuous(bucket.Mode(mode), gens...),
	}

	o.current = o.bucket.Generate()

	return o
}

func (o *Or[T]) Generate() (value T) {
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

	o.current = o.bucket.Generate()
}

func (o *Or[T]) Done() bool {
	return o.done
}
