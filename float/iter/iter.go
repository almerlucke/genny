package iter

type Updater interface {
	Update([]float64)
}

type Iterator struct {
	updater    Updater
	initValues []float64
	values     []float64
	outVector  []float64
}

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
