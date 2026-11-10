package line

import (
	"reflect"
	"testing"
)

func take(l *Line, n int) []float64 {
	out := make([]float64, n)
	for i := range out {
		out[i] = l.Generate()
	}
	return out
}

func TestLine(t *testing.T) {
	l := New(0)
	if l.Continuous() {
		t.Fatal("line is not continuous")
	}
	if !l.Done() || l.Generate() != 0 {
		t.Fatal("a line without target holds its start")
	}

	l.To(10, 5)
	if l.Done() {
		t.Fatal("moving")
	}
	if got := take(l, 5); !reflect.DeepEqual(got, []float64{0, 2, 4, 6, 8}) {
		t.Fatalf("got %v", got)
	}
	if !l.Done() {
		t.Fatal("done after numSteps values")
	}
	if got := take(l, 2); !reflect.DeepEqual(got, []float64{10, 10}) {
		t.Fatalf("a finished line holds its target, got %v", got)
	}
}

func TestLineChainsFromTheLastTarget(t *testing.T) {
	l := New(0)
	l.To(10, 2)
	take(l, 2)
	l.To(0, 2)
	if got := take(l, 3); !reflect.DeepEqual(got, []float64{10, 5, 0}) {
		t.Fatalf("got %v", got)
	}
}

func TestLineReset(t *testing.T) {
	l := New(1)
	l.To(5, 4)
	first := take(l, 4)
	l.Reset()
	if l.Done() {
		t.Fatal("reset must start the line again")
	}
	if second := take(l, 4); !reflect.DeepEqual(first, second) || !reflect.DeepEqual(second, []float64{1, 2, 3, 4}) {
		t.Fatalf("first %v second %v", first, second)
	}
	if !l.Done() {
		t.Fatal("done")
	}
}

func TestLineFrom(t *testing.T) {
	l := New(0)
	l.To(10, 2)
	l.From(3)
	if !l.Done() || l.Generate() != 3 {
		t.Fatal("From jumps and holds")
	}
	l.Reset()
	if !l.Done() || l.Generate() != 3 {
		t.Fatal("Reset after From must not resume the old line")
	}
	l.To(5, 2)
	if got := take(l, 3); !reflect.DeepEqual(got, []float64{3, 4, 5}) {
		t.Fatalf("got %v", got)
	}
}

func TestLineNoSteps(t *testing.T) {
	for _, n := range []int{0, -3} {
		l := New(2)
		l.To(8, n)
		if !l.Done() {
			t.Fatalf("n=%d must be done", n)
		}
		if got := take(l, 3); !reflect.DeepEqual(got, []float64{2, 2, 2}) {
			t.Fatalf("n=%d got %v", n, got)
		}
	}
}
