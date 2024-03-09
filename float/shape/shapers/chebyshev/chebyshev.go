package chebyshev

type Chebyshev struct {
	harmonics   map[int]float64
	maxHarmonic int
}

func New(harmonics map[int]float64) *Chebyshev {
	c := &Chebyshev{}

	c.SetHarmonics(harmonics)

	return c
}

func (c *Chebyshev) SetHarmonics(harmonics map[int]float64) {
	c.harmonics = harmonics

	maxHarmonic := 0

	for k := range c.harmonics {
		if k > maxHarmonic {
			maxHarmonic = k
		}
	}

	c.maxHarmonic = maxHarmonic
}

func (c *Chebyshev) Shape(signal float64) float64 {
	var t0, t1, t2, mix float64

	t0 = 1
	t1 = signal

	for harmonic := 1; harmonic <= c.maxHarmonic; harmonic++ {
		if magnitude, ok := c.harmonics[harmonic]; ok {
			mix += magnitude * t1
		}
		t2 = 2.0*signal*t1 - t0
		t0 = t1
		t1 = t2
	}

	return mix
}
