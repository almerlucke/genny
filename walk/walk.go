// Package walk provides Walk, a random walk through the cells of an N dimensional Container (such as a
// Matrix), moving one step in one dimension at a time.
package walk

import "github.com/almerlucke/genny"

// Container returns a value for an index and has dimensions
type Container[T any] interface {
	ValueAtIndex(index []int) T
	Dimensions() []int
}

// Matrix container for any number of dimensions
type Matrix[T any] struct {
	values     []T
	dimensions []int
}

// NewMatrix with dimensions and values, dimensions are laid out in a single values slice
func NewMatrix[T any](dimensions []int, values []T) *Matrix[T] {
	return &Matrix[T]{
		dimensions: dimensions,
		values:     values,
	}
}

// ValueAtIndex for N dimensional matrix
func (m *Matrix[T]) ValueAtIndex(index []int) T {
	sliceIndex := 0
	dimensionMultiply := 1

	for dim, i := range index {
		sliceIndex += i * dimensionMultiply
		dimensionMultiply *= m.dimensions[dim]
	}

	return m.values[sliceIndex]
}

// Dimensions of the container
func (m *Matrix[T]) Dimensions() []int {
	return m.dimensions
}

// Walk for doing a random walk inside a Container
type Walk[T any] struct {
	container Container[T]
	index     []int
}

// New creates a new random walk
func New[T any](container Container[T]) *Walk[T] {
	w := &Walk[T]{
		container: container,
	}

	dims := container.Dimensions()
	index := make([]int, len(dims))

	for i, d := range dims {
		if d > 0 {
			index[i] = genny.Intn(d)
		}
	}

	w.index = index

	return w
}

// Generate for random walk
func (w *Walk[T]) Generate() T {
	dimensions := w.container.Dimensions()

	// only dimensions with room to move can be walked
	var walkable []int
	for dim, size := range dimensions {
		if size > 1 {
			walkable = append(walkable, dim)
		}
	}

	if len(walkable) == 0 {
		return w.container.ValueAtIndex(w.index)
	}

	selectedDimension := walkable[genny.Intn(len(walkable))]
	selectedDimensionSize := dimensions[selectedDimension]
	selectedIndex := w.index[selectedDimension]

	if selectedIndex < 1 {
		selectedIndex++
	} else if selectedIndex > (selectedDimensionSize - 2) {
		selectedIndex--
	} else {
		if genny.Intn(2) == 1 {
			selectedIndex++
		} else {
			selectedIndex--
		}
	}

	w.index[selectedDimension] = selectedIndex

	return w.container.ValueAtIndex(w.index)
}

// Continuous will always return true
func (w *Walk[T]) Continuous() bool {
	return true
}

// Done will always return false
func (w *Walk[T]) Done() bool {
	return false
}

// Reset does nothing
func (w *Walk[T]) Reset() {
}
