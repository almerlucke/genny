// Package choose picks between generators at random
package choose

import "github.com/almerlucke/genny"

// Option is a generator with a relative weight
type Option[T any] struct {
	Gen    genny.Generator[T]
	Weight float64
}

// Weighted picks on every Generate one of its generators with a chance proportional to the weight,
// and generates from that one. Generators that are not picked do not advance. Generators that are done
// are not picked anymore
type Weighted[T any] struct {
	options []Option[T]
}

// New creates a weighted choice
func New[T any](options ...Option[T]) *Weighted[T] {
	return &Weighted[T]{options: options}
}

// SetWeight changes the weight of option i
func (w *Weighted[T]) SetWeight(i int, weight float64) { w.options[i].Weight = weight }

func (w *Weighted[T]) Generate() (v T) {
	total := 0.0
	last := -1
	for i, o := range w.options {
		if o.Weight > 0 && !o.Gen.Done() {
			total += o.Weight
			last = i
		}
	}
	if last < 0 {
		// nothing to pick, give the last option the chance to repeat its end value
		if len(w.options) == 0 {
			return
		}
		return w.options[len(w.options)-1].Gen.Generate()
	}

	r := genny.Float64() * total
	pick := last
	acc := 0.0
	for i, o := range w.options {
		if o.Weight > 0 && !o.Gen.Done() {
			acc += o.Weight
			if r < acc {
				pick = i
				break
			}
		}
	}
	return w.options[pick].Gen.Generate()
}

// Continuous is true if all generators are continuous
func (w *Weighted[T]) Continuous() bool {
	for _, o := range w.options {
		if !o.Gen.Continuous() {
			return false
		}
	}
	return true
}

// Done is true if all generators are done
func (w *Weighted[T]) Done() bool {
	for _, o := range w.options {
		if !o.Gen.Done() {
			return false
		}
	}
	return true
}

func (w *Weighted[T]) Reset() {
	for _, o := range w.options {
		o.Gen.Reset()
	}
}
