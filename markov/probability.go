package markov

import (
	"github.com/almerlucke/genny"
)

// ProbabilityState is a State that picks its successor by weighted probability
type ProbabilityState[T any] struct {
	value         T
	states        []*ProbabilityState[T]
	probabilities []float64
	total         float64
}

// NewProbabilityState creates a state with value v and no successors
func NewProbabilityState[T any](v T) *ProbabilityState[T] {
	return &ProbabilityState[T]{
		value: v,
	}
}

// SetProbabilities sets the states that can follow, as pairs of a *ProbabilityState (nil to end the chain)
// and a float64 weight
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

// Next picks the next state by probability. A nil state ends the chain, so do states without probabilities
func (ps *ProbabilityState[T]) Next(_ []State[T]) State[T] {
	r := genny.Float64() * ps.total
	acc := 0.0

	for index, probability := range ps.probabilities {
		acc += probability
		if r < acc {
			if next := ps.states[index]; next != nil {
				return next
			}
			return nil
		}
	}

	return nil
}

func (ps *ProbabilityState[T]) Start(_ []State[T]) State[T] {
	return ps
}
