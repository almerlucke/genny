package lookup

import (
	"log"
	"math"
)

// Table for shaping, make sure table always contains one more sample, so we do not need to wrap
type Table []float64

type FillMode int

const (
	First FillMode = iota
	Last
	Custom
)

func NewWithFunc(n int, fillMode FillMode, f func(int, int) float64, last ...float64) Table {
	table := make(Table, n+1)

	for i := 0; i < n; i++ {
		table[i] = f(i, n)
	}

	if fillMode == First {
		table[n] = table[0]
	} else if fillMode == Last {
		table[n] = table[n-1]
	} else if len(last) > 0 {
		table[n] = last[0]
	} else {
		table[n] = 0
	}

	return table
}

func NewWithRampFunc(n int, fillMode FillMode, f func(float64) float64, last ...float64) Table {
	return NewWithFunc(n, fillMode, func(i int, n int) float64 {
		return f(float64(i) / float64(n))
	}, last...)
}

func NewWithPhaseFunc(n int, fillMode FillMode, f func(float64) float64, last ...float64) Table {
	return NewWithRampFunc(n, fillMode, func(v float64) float64 {
		return f(v * 2.0 * math.Pi)
	}, last...)
}

func NewSineTable(n int) Table {
	return NewWithPhaseFunc(n, First, math.Sin)
}

func NewUniSineTable(n int) Table {
	return NewWithPhaseFunc(n, First, unipolar(math.Sin))
}

func NewCosineTable(n int) Table {
	return NewWithPhaseFunc(n, First, math.Cos)
}

func NewUniCosineTable(n int) Table {
	return NewWithPhaseFunc(n, First, unipolar(math.Cos))
}

func NewSawtoothTable(n int) Table {
	return NewWithRampFunc(n, First, func(x float64) float64 { return x*2.0 - 1.0 })
}

func NewUniSawtoothTable(n int) Table {
	return NewWithRampFunc(n, First, func(x float64) float64 { return x })
}

func NewTriangleTable(n int) Table {
	return NewWithRampFunc(n, First, tri)
}

func NewUniTriangleTable(n int) Table {
	return NewWithRampFunc(n, First, unipolar(tri))
}

func NewSquareTable(n int) Table {
	return NewWithRampFunc(n, First, func(x float64) float64 {
		if x < 0.5 {
			return 1.0
		}
		return -1.0
	})
}

func NewUniSquareTable(n int) Table {
	return NewWithRampFunc(n, First, func(x float64) float64 {
		if x < 0.5 {
			return 1.0
		}
		return 0.0
	})
}

func (t Table) Shape(x float64) float64 {
	nf := x * float64(len(t)-1)
	n1 := int(nf)
	t1 := t[n1]
	return t1 + (nf-float64(n1))*(t[n1+1]-t1)
}

func (t Table) Print() {
	for _, v := range t {
		log.Printf("%f", v)
	}
}

func tri(x float64) float64 {
	x *= 4.0
	if x >= 3.0 {
		x -= 4.0
	} else if x > 1.0 {
		x = 2.0 - x
	}
	return x
}

func unipolar(f func(float64) float64) func(float64) float64 {
	return func(x float64) float64 {
		return f(x)*0.5 + 0.5
	}
}
