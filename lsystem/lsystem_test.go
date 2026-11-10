package lsystem

import "testing"

func TestExpand(t *testing.T) {
	// Lindenmayer's algae
	rules := map[rune]string{'A': "AB", 'B': "A"}
	for n, want := range []string{"A", "AB", "ABA", "ABAAB", "ABAABABA"} {
		if got := Expand("A", rules, n); got != want {
			t.Errorf("n=%d: %s, want %s", n, got, want)
		}
	}
}

func TestExpandIsCapped(t *testing.T) {
	if got := Expand("A", map[rune]string{'A': "AAAA"}, 40); len(got) > MaxLen {
		t.Errorf("len %d", len(got))
	}
}

func TestSequence(t *testing.T) {
	rules := map[rune]string{'A': "A[B]A"}
	s := New("A", rules, 1, map[rune]int{'A': 60, 'B': 67})
	var got []int
	for !s.Done() {
		got = append(got, s.Generate())
	}
	if len(got) != 3 || got[0] != 60 || got[1] != 67 || got[2] != 60 {
		t.Errorf("%v", got)
	}
	l := NewLoop("A", nil, 3, map[rune]int{'A': 1})
	if !l.Continuous() || l.Generate() != 1 || l.Generate() != 1 {
		t.Error("loop")
	}
	if z := New("X", nil, 1, map[rune]int{}); z.Generate() != 0 {
		t.Error("no symbols")
	}
}
