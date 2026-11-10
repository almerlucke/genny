package mutate

import (
	"slices"
	"testing"

	"github.com/almerlucke/genny"
)

func pass(s *Sequence[int], n int) []int {
	var out []int
	for range n {
		out = append(out, s.Generate())
	}
	return out
}

func TestNoMutationRepeats(t *testing.T) {
	s := New([]int{1, 2, 3}, Config[int]{})
	if got := pass(s, 9); !slices.Equal(got, []int{1, 2, 3, 1, 2, 3, 1, 2, 3}) {
		t.Errorf("%v", got)
	}
}

func TestReverseEveryPass(t *testing.T) {
	s := New([]int{1, 2, 3}, Config[int]{ReverseProb: 1})
	if got := pass(s, 9); !slices.Equal(got, []int{1, 2, 3, 3, 2, 1, 1, 2, 3}) {
		t.Errorf("%v", got)
	}
}

func TestMutationsKeepLengthAndNotes(t *testing.T) {
	genny.Seed(4)
	orig := []int{1, 2, 3, 4, 5, 6, 7, 8}
	s := New(orig, Config[int]{SwapProb: 1, RotateProb: 1, ReverseProb: 0.5})
	changed := false
	for range 200 {
		pass(s, len(orig))
		v := s.Values()
		if len(v) != len(orig) {
			t.Fatal(v)
		}
		sorted := slices.Sorted(slices.Values(v))
		if !slices.Equal(sorted, orig) {
			t.Fatalf("notes lost: %v", v)
		}
		changed = changed || !slices.Equal(v, orig)
	}
	if !changed {
		t.Error("never changed")
	}
	if !slices.Equal(orig, []int{1, 2, 3, 4, 5, 6, 7, 8}) {
		t.Error("input slice was changed")
	}
	s.Reset()
	if !slices.Equal(s.Values(), orig) {
		t.Error("reset")
	}
}

func TestReplace(t *testing.T) {
	s := New([]int{1, 1, 1, 1}, Config[int]{ReplaceProb: 1, Replace: func() int { return 9 }})
	pass(s, 4)
	if n := slices.Index(s.Values(), 9); n < 0 {
		t.Error("nothing replaced")
	}
	s = New([]int{1, 2, 3}, Config[int]{ReplaceProb: 1}) // copies another note
	for range 50 {
		pass(s, 3)
	}
	for _, v := range s.Values() {
		if v < 1 || v > 3 {
			t.Fatal(v)
		}
	}
}
