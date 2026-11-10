// Package progression makes chord progressions with a Markov chain over scale degrees. Degrees are
// 0-based (0 is I, 4 is V), use them with chord.Diatonic or scale.Degree
package progression

import (
	"github.com/almerlucke/genny/markov"
)

// Move is a possible next degree with a relative weight
type Move struct {
	To     int
	Weight float64
}

// Table gives for each degree the moves that can follow it. A degree without moves ends the
// progression, a move to a degree that is not in the table goes back to the start
type Table map[int][]Move

// New creates a progression generator that starts on degree start. Because a Markov chain is not
// continuous it is only done when a degree without moves is reached
func New(table Table, start int) *markov.Markov[int] {
	states := map[int]*markov.ProbabilityState[int]{}
	state := func(d int) *markov.ProbabilityState[int] {
		if s, ok := states[d]; ok {
			return s
		}
		s := markov.NewProbabilityState(d)
		states[d] = s
		return s
	}

	for from := range table {
		state(from)
	}
	startState := state(start)

	for from, moves := range table {
		args := make([]any, 0, 2*len(moves))
		for _, m := range moves {
			to := startState
			if _, ok := table[m.To]; ok {
				to = state(m.To)
			}
			args = append(args, to, m.Weight)
		}
		state(from).SetProbabilities(args...)
	}

	return markov.New[int](startState, 1)
}

// Some ready made tables
var (
	// Pop moves around I, IV, V and vi (I-V-vi-IV and relatives)
	Pop = Table{
		0: {{3, 3}, {4, 3}, {5, 3}, {1, 1}},
		1: {{4, 4}, {3, 1}},
		2: {{5, 3}, {3, 2}},
		3: {{0, 3}, {4, 3}, {5, 2}, {1, 1}},
		4: {{0, 4}, {5, 3}, {3, 2}},
		5: {{3, 4}, {4, 2}, {1, 2}, {0, 1}},
	}

	// Classical follows functional harmony: tonic, subdominant, dominant, back to tonic
	Classical = Table{
		0: {{3, 3}, {4, 3}, {5, 2}, {1, 2}, {2, 1}},
		1: {{4, 6}, {6, 2}},
		2: {{5, 3}, {3, 2}},
		3: {{4, 4}, {0, 2}, {1, 2}, {6, 1}},
		4: {{0, 7}, {5, 2}, {3, 1}},
		5: {{1, 4}, {3, 3}, {4, 2}},
		6: {{0, 6}, {2, 1}},
	}

	// Jazz favours ii-V-I turnarounds
	Jazz = Table{
		0: {{5, 3}, {1, 3}, {3, 2}, {2, 2}},
		1: {{4, 8}, {6, 1}, {1, 1}},
		2: {{5, 6}, {3, 2}},
		3: {{1, 4}, {4, 3}, {0, 2}, {6, 1}},
		4: {{0, 7}, {5, 2}, {2, 1}},
		5: {{1, 6}, {4, 2}, {3, 2}},
		6: {{2, 5}, {0, 3}},
	}

	// Blues uses I, IV and V only
	Blues = Table{
		0: {{0, 5}, {3, 3}, {4, 2}},
		3: {{0, 5}, {3, 2}, {4, 3}},
		4: {{0, 7}, {3, 3}},
	}
)
