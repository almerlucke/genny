// Package iter provides Iterator, a FrameGenerator that repeatedly applies an Updater (such as an
// iterated map from the chaos package) to a vector of values.
package iter

// Updater advances a vector of values by one step
type Updater interface {
	Update([]float64)
}

// Iterator generates the successive states of an Updater
type Iterator struct {
	updater    Updater
	initValues []float64
	values     []float64
	outVector  []float64
}

// New creates an Iterator that starts from initValues
func New(initValues []float64, updater Updater) *Iterator {
	iter := &Iterator{
		initValues: initValues,
		values:     make([]float64, len(initValues)),
		updater:    updater,
		outVector:  make([]float64, len(initValues)),
	}

	copy(iter.values, initValues)

	return iter
}

func (iter *Iterator) SetValues(vs []float64) {
	copy(iter.values, vs)
}

func (iter *Iterator) Dimensions() int {
	return len(iter.outVector)
}

func (iter *Iterator) Generate() []float64 {
	copy(iter.outVector, iter.values)
	iter.updater.Update(iter.values)
	return iter.outVector
}

func (iter *Iterator) Continuous() bool {
	return true
}

func (iter *Iterator) Done() bool {
	return false
}

func (iter *Iterator) Reset() {
	copy(iter.values, iter.initValues)
	for i := 0; i < len(iter.outVector); i++ {
		iter.outVector[i] = 0.0
	}
}
