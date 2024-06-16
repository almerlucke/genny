package quantize

import "math"

type Quantizer struct {
	quantum float64
}

func New(quantum float64) *Quantizer {
	return &Quantizer{quantum}
}

func (q *Quantizer) SetQuantum(quantum float64) {
	q.quantum = quantum
}

func (q *Quantizer) Quantum() float64 {
	return q.quantum
}

func (q *Quantizer) Shape(x float64) float64 {
	return math.Round(x/q.quantum) * q.quantum
}
