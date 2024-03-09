package lookup

import "math"

// Table for shaping, make sure table always contains one more sample, so we do not need to wrap
type Table []float64

func newPhaseTable(n int, f func(float64) float64) Table {
	table := make(Table, n+1)
	phase := 0.0
	inc := 2.0 * math.Pi / float64(n)

	for i := 0; i < n; i++ {
		table[i] = f(phase)
		phase += inc
	}

	table[n] = table[0]

	return table
}

func NewSineTable(n int) Table {
	return newPhaseTable(n, math.Sin)
}

func NewCosineTable(n int) Table {
	return newPhaseTable(n, math.Cos)
}

func NewNormalizedSineTable(n int) Table {
	return newPhaseTable(n, func(phase float64) float64 { return math.Sin(phase)*0.5 + 0.5 })
}

func NewNormalizedCosineTable(n int) Table {
	return newPhaseTable(n, func(phase float64) float64 { return math.Cos(phase)*0.5 + 0.5 })
}

func (t Table) Shape(x float64) float64 {
	nf := x * float64(len(t)-1)
	n1 := int(nf)
	t1 := t[n1]
	return t1 + (nf-float64(n1))*(t[n1+1]-t1)
}
