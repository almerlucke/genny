// Package markov provides Markov, a generator that walks a Markov chain of States. ProbabilityState is
// a state that chooses its successor by weighted probability.
package markov

// State is a state of a Markov chain.
type State[T any] interface {
	// Value is the value generated while in this state
	Value() T
	// Next chooses the next state, given the history (most recent state first). To end the chain
	// Next must return an untyped nil (a plain `return nil`, not a nil pointer of a state type)
	Next(history []State[T]) State[T]
}

// Starter chooses the first state, nil (untyped, see State.Next) gives a chain that is done at once
type Starter[T any] interface {
	Start(history []State[T]) State[T]
}

// Markov generates values by walking from state to state. It is done when a state has no next state
type Markov[T any] struct {
	history      []State[T]
	historySize  int
	starter      Starter[T]
	currentState State[T]
}

// New creates a Markov chain, historySize is the number of past states passed to State.Next
func New[T any](starter Starter[T], historySize int) *Markov[T] {
	var history []State[T]

	return &Markov[T]{
		history:      history,
		historySize:  historySize,
		starter:      starter,
		currentState: starter.Start(history),
	}
}

func (m *Markov[T]) pushState(state State[T]) {
	if m.historySize <= 0 {
		return
	}

	if len(m.history) < m.historySize {
		m.history = append([]State[T]{state}, m.history...)
		return
	}

	copy(m.history[1:], m.history)

	m.history[0] = state
}

func (m *Markov[T]) Generate() (value T) {
	if m.currentState == nil {
		return
	}

	value = m.currentState.Value()

	nextState := m.currentState.Next(m.history)

	m.pushState(m.currentState)

	m.currentState = nextState

	return
}

func (m *Markov[T]) Continuous() bool {
	return false
}

// Done is true when the chain has ended
func (m *Markov[T]) Done() bool {
	return m.currentState == nil
}

func (m *Markov[T]) Reset() {
	m.history = []State[T]{}
	m.currentState = m.starter.Start(m.history)
}
