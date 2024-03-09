package mult

type Mult struct {
	M float64
}

func New(m float64) *Mult {
	return &Mult{M: m}
}

func (m *Mult) Shape(x float64) float64 {
	return m.M * x
}
