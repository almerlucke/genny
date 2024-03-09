package linear

type Linear struct {
	Scale float64
	Shift float64
}

func New(scale float64, shift float64) *Linear {
	return &Linear{
		Scale: scale,
		Shift: shift,
	}
}

func NewBipolar() *Linear {
	return &Linear{Scale: 2.0, Shift: -1.0}
}

func NewUnipolar() *Linear {
	return &Linear{Scale: 0.5, Shift: 0.5}
}

func (l *Linear) Shape(x float64) float64 {
	return x*l.Scale + l.Shift
}
