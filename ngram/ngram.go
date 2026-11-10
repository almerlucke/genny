// Package ngram learns a Markov chain from an example sequence (a melody, a rhythm, a chord list) and
// generates new sequences in the same style: the next value is chosen by what followed the last
// few values in the example
package ngram

import (
	"fmt"

	"github.com/almerlucke/genny"
)

type followers[T comparable] struct {
	values []T // in order of first appearance, so seeded runs repeat
	counts []float64
	total  float64
}

func (f *followers[T]) add(v T) {
	for i, x := range f.values {
		if x == v {
			f.counts[i]++
			f.total++
			return
		}
	}
	f.values = append(f.values, v)
	f.counts = append(f.counts, 1)
	f.total++
}

func (f *followers[T]) pick() T {
	r := genny.Float64() * f.total
	acc := 0.0
	for i, c := range f.counts {
		acc += c
		if r < acc {
			return f.values[i]
		}
	}
	return f.values[len(f.values)-1]
}

// Chain generates values with the statistics of the sequence it learned
type Chain[T comparable] struct {
	order int
	seq   []T
	table map[string]*followers[T]
	ctx   []T
}

// the context key, values are compared by their printed form
func key[T any](ctx []T) string { return fmt.Sprint(ctx) }

// Learn builds a chain from seq. The next value depends on the last order values (1 is the simplest,
// 2 or 3 follow the example more closely). When the last values were never followed by anything in the
// example the chain backs off to a shorter context. It starts with the first values of the example.
// It panics if order is less than 1 or seq is not longer than order
func Learn[T comparable](seq []T, order int) *Chain[T] {
	if order < 1 || len(seq) <= order {
		panic("ngram: need order >= 1 and a sequence longer than the order")
	}
	c := &Chain[T]{order: order, seq: append([]T{}, seq...), table: map[string]*followers[T]{}}

	for l := 1; l <= order; l++ {
		for i := 0; i+l < len(seq); i++ {
			k := key(seq[i : i+l])
			f := c.table[k]
			if f == nil {
				f = &followers[T]{}
				c.table[k] = f
			}
			f.add(seq[i+l])
		}
	}

	c.Reset()
	return c
}

func (c *Chain[T]) Generate() T {
	for l := min(len(c.ctx), c.order); l >= 1; l-- {
		if f := c.table[key(c.ctx[len(c.ctx)-l:])]; f != nil {
			v := f.pick()
			c.push(v)
			return v
		}
	}

	// a dead end (the context only occurs at the end of the example): start somewhere else
	i := genny.Intn(len(c.seq) - c.order)
	c.ctx = append(c.ctx[:0], c.seq[i:i+c.order]...)
	v := c.seq[i+c.order]
	c.push(v)
	return v
}

func (c *Chain[T]) push(v T) {
	c.ctx = append(c.ctx, v)
	if len(c.ctx) > c.order {
		c.ctx = c.ctx[1:]
	}
}

func (c *Chain[T]) Continuous() bool { return true }

func (c *Chain[T]) Done() bool { return false }

func (c *Chain[T]) Reset() { c.ctx = append([]T{}, c.seq[:c.order]...) }
