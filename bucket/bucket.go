package bucket

import "math/rand"

type Mode int

const (
	Random Mode = iota
	Indexed
)

// Bucket holds values which can be fetched randomly or one by one. The values are shuffled random initially
type Bucket[T any] struct {
	values     []T
	mode       Mode
	index      int
	done       bool
	continuous bool
}

// New creates a bucket from the given values and shuffles them randomly.
// NextValue can pick an element at random (Random mode) or one by one until the bucket is depleted (Indexed mode)
func New[T any](mode Mode, values ...T) *Bucket[T] {
	b := &Bucket[T]{
		values: values,
		mode:   mode,
	}

	b.shuffle()

	return b
}

// NewContinuous creates a continuous bucket from the given values and shuffles them randomly.
// NextValue can pick an element at random (Random mode) or one by one until the bucket is depleted (Indexed mode)
func NewContinuous[T any](mode Mode, values ...T) *Bucket[T] {
	b := &Bucket[T]{
		values:     values,
		mode:       mode,
		continuous: true,
	}

	b.shuffle()

	return b
}

// NextValue gets the next value from the bucket. If all values are fetched from the bucket return nil
func (b *Bucket[T]) NextValue() T {
	if b.mode == Random {
		return b.values[rand.Intn(len(b.values))]
	}

	if b.done {
		return b.values[len(b.values)-1]
	}

	v := b.values[b.index]

	b.index++

	if b.index >= len(b.values) {
		if b.continuous {
			b.Reset()
		} else {
			b.done = true
		}
	}

	return v
}

// Done checks if bucket is done in non-continuous mode
func (b *Bucket[T]) Done() bool {
	return b.done
}

// Continuous checks if the bucket is continuous
func (b *Bucket[T]) Continuous() bool {
	return b.mode == Random || (b.mode == Indexed && b.continuous)
}

// Reset the bucket
func (b *Bucket[T]) Reset() {
	b.index = 0
	b.shuffle()
}

// Values returns bucket internal values slice
func (b *Bucket[T]) Values() []T {
	return b.values
}

// shuffle all values randomly
func (b *Bucket[T]) shuffle() {
	rand.Shuffle(len(b.values), func(i, j int) {
		b.values[i], b.values[j] = b.values[j], b.values[i]
	})
}
