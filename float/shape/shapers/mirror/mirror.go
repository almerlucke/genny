package mirror

type Mirror struct {
	Bottom float64
	Top    float64
}

func New(bottom float64, top float64) *Mirror {
	return &Mirror{
		Bottom: bottom,
		Top:    top,
	}
}

func (mirror *Mirror) Shape(x float64) float64 {
	for {
		if x > mirror.Top {
			x = mirror.Top*2 - x
		} else if x < mirror.Bottom {
			x = mirror.Bottom*2 - x
		} else {
			break
		}
	}

	return x
}
