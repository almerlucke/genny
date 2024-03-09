package add

type Add struct {
	A float64
}

func New(a float64) *Add {
	return &Add{A: a}
}

func (a *Add) Shape(x float64) float64 {
	return x + a.A
}
