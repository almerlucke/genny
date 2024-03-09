package thru

type Thru struct {
}

func New() *Thru {
	return &Thru{}
}

func (t *Thru) Shape(x float64) float64 {
	return x
}
