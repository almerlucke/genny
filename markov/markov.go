package markov

import (
	"reflect"
)

type State[T any] interface {
	Value() T
	Next(history []State[T]) State[T]
}

type Starter[T any] interface {
	Start(history []State[T]) State[T]
}

type Markov[T any] struct {
	history      []State[T]
	historySize  int
	starter      Starter[T]
	currentState State[T]
}

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
	if len(m.history) < m.historySize {
		m.history = append([]State[T]{state}, m.history...)
		return
	}

	copy(m.history[1:], m.history)

	m.history[0] = state
}

func (m *Markov[T]) NextValue() (value T) {
	if reflect.ValueOf(m.currentState).IsNil() {
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

func (m *Markov[T]) Done() bool {
	return reflect.ValueOf(m.currentState).IsNil()
}

func (m *Markov[T]) Reset() {
	m.history = []State[T]{}
	m.currentState = m.starter.Start(m.history)
}
