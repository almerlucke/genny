package markov_test

import (
	"math"
	"reflect"
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/markov"
)

// state is a state with a fixed successor, a nil next ends the chain
type state struct {
	value int
	next  markov.State[int]
	seen  *[][]int // records the values in the history passed to Next
}

func (s *state) Value() int { return s.value }

func (s *state) Next(history []markov.State[int]) markov.State[int] {
	if s.seen != nil {
		var h []int
		for _, st := range history {
			h = append(h, st.Value())
		}
		*s.seen = append(*s.seen, h)
	}
	return s.next
}

type starter struct{ first markov.State[int] }

func (s starter) Start(_ []markov.State[int]) markov.State[int] { return s.first }

func chain(seen *[][]int, values ...int) *state {
	var next markov.State[int]
	for i := len(values) - 1; i >= 0; i-- {
		s := &state{value: values[i], next: next, seen: seen}
		next = s
		if i == 0 {
			return s
		}
	}
	return nil
}

func TestMarkovWalksUntilStateEnds(t *testing.T) {
	m := markov.New[int](starter{chain(nil, 1, 2, 3)}, 2)
	if m.Continuous() {
		t.Fatal("markov is never continuous")
	}
	for i, want := range []int{1, 2, 3} {
		if m.Done() {
			t.Fatalf("done before value %d", i)
		}
		if got := m.Generate(); got != want {
			t.Fatalf("value %d: got %d want %d", i, got, want)
		}
	}
	if !m.Done() {
		t.Fatal("done after the last state")
	}
	if m.Generate() != 0 {
		t.Fatal("a done chain generates zero")
	}
	m.Reset()
	if m.Done() || m.Generate() != 1 {
		t.Fatal("reset restarts the chain")
	}
}

func TestMarkovStarterNilIsDoneAtOnce(t *testing.T) {
	m := markov.New[int](starter{nil}, 1)
	if !m.Done() || m.Generate() != 0 {
		t.Fatal("nil start gives a chain that is done at once")
	}
	m.Reset()
	if !m.Done() {
		t.Fatal("still done after reset")
	}
}

func TestMarkovHistory(t *testing.T) {
	var seen [][]int
	m := markov.New[int](starter{chain(&seen, 1, 2, 3, 4)}, 2)
	for !m.Done() {
		m.Generate()
	}
	// most recent state first, at most historySize states
	want := [][]int{nil, {1}, {2, 1}, {3, 2}}
	if !reflect.DeepEqual(seen, want) {
		t.Fatalf("history %v want %v", seen, want)
	}

	seen = nil
	m = markov.New[int](starter{chain(&seen, 1, 2, 3)}, 0)
	for !m.Done() {
		m.Generate()
	}
	if !reflect.DeepEqual(seen, [][]int{nil, nil, nil}) {
		t.Fatalf("history size 0 must pass no history, got %v", seen)
	}

	// Reset clears the history
	seen = nil
	m = markov.New[int](starter{chain(&seen, 1, 2)}, 3)
	m.Generate()
	m.Reset()
	m.Generate()
	if !reflect.DeepEqual(seen, [][]int{nil, nil}) {
		t.Fatalf("history after reset %v", seen)
	}
}

func TestProbabilityStateChain(t *testing.T) {
	a := markov.NewProbabilityState("a")
	b := markov.NewProbabilityState("b")
	a.SetProbabilities(b, 1.0)
	b.SetProbabilities(nil, 1.0)
	if a.Value() != "a" || a.Start(nil) != markov.State[string](a) {
		t.Fatal("value / start")
	}

	m := markov.New[string](a, 1)
	if got := []string{m.Generate(), m.Generate()}; !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("got %v", got)
	}
	if !m.Done() {
		t.Fatal("a nil target ends the chain")
	}
}

func TestProbabilityStateNextIsUntypedNil(t *testing.T) {
	end := markov.NewProbabilityState(1)
	end.SetProbabilities(nil, 1.0)
	if end.Next(nil) != nil {
		t.Fatal("a nil target must give an untyped nil")
	}

	none := markov.NewProbabilityState(1)
	if none.Next(nil) != nil {
		t.Fatal("a state without probabilities must give an untyped nil")
	}
	m := markov.New[int](none, 1)
	m.Generate()
	if !m.Done() {
		t.Fatal("chain ends after a state without probabilities")
	}

	zero := markov.NewProbabilityState(1)
	zero.SetProbabilities(none, 0.0)
	if zero.Next(nil) != nil {
		t.Fatal("zero total weight ends the chain")
	}
}

func TestProbabilityStateWeights(t *testing.T) {
	genny.Seed(123)
	a := markov.NewProbabilityState("a")
	b := markov.NewProbabilityState("b")
	c := markov.NewProbabilityState("c")
	a.SetProbabilities(b, 3.0, c, 1.0)

	count := map[string]int{}
	const n = 4000
	for i := 0; i < n; i++ {
		count[a.Next(nil).Value()]++
	}
	if count["a"] != 0 || count["b"]+count["c"] != n {
		t.Fatalf("unexpected targets %v", count)
	}
	if ratio := float64(count["b"]) / n; math.Abs(ratio-0.75) > 0.05 {
		t.Fatalf("b chosen %.3f of the time, want about 0.75", ratio)
	}
}

func TestProbabilityStateRepeatable(t *testing.T) {
	build := func() *markov.Markov[int] {
		a := markov.NewProbabilityState(0)
		b := markov.NewProbabilityState(1)
		a.SetProbabilities(a, 1.0, b, 1.0, nil, 0.2)
		b.SetProbabilities(a, 1.0, b, 1.0)
		return markov.New[int](a, 2)
	}
	run := func() []int {
		genny.Seed(8)
		m := build()
		var out []int
		for i := 0; i < 50 && !m.Done(); i++ {
			out = append(out, m.Generate())
		}
		return out
	}
	if a, b := run(), run(); !reflect.DeepEqual(a, b) {
		t.Fatalf("%v != %v", a, b)
	}
}
