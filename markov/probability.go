package markov

import (
	"math/rand"
)

type ProbabilityState[T any] struct {
	value         T
	states        []*ProbabilityState[T]
	probabilities []float64
	total         float64
}

func NewProbabilityState[T any](v T) *ProbabilityState[T] {
	return &ProbabilityState[T]{
		value: v,
	}
}

func (ps *ProbabilityState[T]) SetProbabilities(statesAndProbs ...any) {
	numStates := len(statesAndProbs) / 2
	states := make([]*ProbabilityState[T], numStates)
	probabilities := make([]float64, numStates)
	total := 0.0

	for i, j := 0, 0; j < numStates; i, j = i+2, j+1 {
		var state *ProbabilityState[T] = nil
		state, _ = statesAndProbs[i].(*ProbabilityState[T])
		states[j] = state
		probabilities[j] = statesAndProbs[i+1].(float64)
		total += probabilities[j]
	}

	ps.total = total
	ps.probabilities = probabilities
	ps.states = states
}

func (ps *ProbabilityState[T]) Value() T {
	return ps.value
}

func (ps *ProbabilityState[T]) Next(history []State[T]) (s State[T]) {
	r := rand.Float64() * ps.total
	acc := 0.0

	for index, probability := range ps.probabilities {
		acc += probability
		if r < acc {
			s = ps.states[index]
			break
		}
	}

	return
}

func (ps *ProbabilityState[T]) Start(history []State[T]) State[T] {
	return ps
}
