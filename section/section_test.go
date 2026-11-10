package section

import (
	"slices"
	"testing"

	"github.com/almerlucke/genny/constant"
	"github.com/almerlucke/genny/sequence"
)

func take(s *Section[string], n int) (out []string) {
	for range n {
		out = append(out, s.Generate())
	}
	return
}

func TestPlaysPartsOnce(t *testing.T) {
	s := New(Part[string]{Gen: constant.New("a"), Steps: 2}, Part[string]{Gen: constant.New("b"), Steps: 3})
	got := take(s, 5)
	if !slices.Equal(got, []string{"a", "a", "b", "b", "b"}) || !s.Done() || s.Continuous() {
		t.Errorf("%v done=%v", got, s.Done())
	}
	if take(s, 1)[0] != "b" {
		t.Error("done section should repeat the last value")
	}
	s.Reset()
	if s.Done() || s.Generate() != "a" {
		t.Error("reset")
	}
}

func TestLoopRestartAndCarryOn(t *testing.T) {
	melody := sequence.NewLoop("x", "y", "z")
	s := NewLoop(
		Part[string]{Gen: melody, Steps: 2},
		Part[string]{Gen: constant.New("-"), Steps: 1},
	)
	if got := take(s, 6); !slices.Equal(got, []string{"x", "y", "-", "z", "x", "-"}) {
		t.Errorf("carry on: %v", got)
	}

	melody = sequence.NewLoop("x", "y", "z")
	s = NewLoop(
		Part[string]{Gen: melody, Steps: 2, Restart: true},
		Part[string]{Gen: constant.New("-"), Steps: 1},
	)
	if got := take(s, 6); !slices.Equal(got, []string{"x", "y", "-", "x", "y", "-"}) {
		t.Errorf("restart: %v", got)
	}
}

func TestZeroStepsPlaysUntilDone(t *testing.T) {
	s := New(Part[string]{Gen: sequence.New("a", "b", "c")}, Part[string]{Gen: constant.New("end"), Steps: 1})
	if got := take(s, 4); !slices.Equal(got, []string{"a", "b", "c", "end"}) || !s.Done() {
		t.Errorf("%v", got)
	}
}
