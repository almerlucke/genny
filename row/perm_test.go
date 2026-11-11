package row

import (
	"slices"
	"testing"
	"time"

	"github.com/almerlucke/genny"
)

func TestPermutations(t *testing.T) {
	tests := []struct {
		name         string
		elems        []int
		n, groupLen  int
		want         [][]int
		wantLenCheck int
	}{
		{"all of three", []int{1, 2, 3}, 0, 0,
			[][]int{{1, 2, 3}, {1, 3, 2}, {2, 1, 3}, {2, 3, 1}, {3, 1, 2}, {3, 2, 1}}, 6},
		{"limited", []int{1, 2, 3}, 2, 3, [][]int{{1, 2, 3}, {1, 3, 2}}, 2},
		{"groups of two", []int{5, 6, 7}, -1, 2,
			[][]int{{5, 6}, {5, 7}, {6, 5}, {6, 7}, {7, 5}, {7, 6}}, 6},
		{"group too long", []int{1, 2}, 0, 9, [][]int{{1, 2}, {2, 1}}, 2},
		{"group of one", []int{4, 9}, 0, 1, [][]int{{4}, {9}}, 2},
		{"duplicates count by position", []int{1, 1}, 0, 0, [][]int{{1, 1}, {1, 1}}, 2},
		{"empty", nil, 0, 0, nil, 0},
	}
	for _, tt := range tests {
		got := Permutations(tt.elems, tt.n, tt.groupLen)
		if len(got) != len(tt.want) {
			t.Errorf("%s: %v", tt.name, got)
			continue
		}
		for i := range got {
			if !slices.Equal(got[i], tt.want[i]) {
				t.Errorf("%s: %v want %v", tt.name, got, tt.want)
				break
			}
		}
	}
}

func TestPermutationsCap(t *testing.T) {
	// 12! is far above the cap: clamped, not slow
	got := Permutations([]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, 0, 0)
	if len(got) != MaxPermutations {
		t.Errorf("%d rows", len(got))
	}
	if got := Permutations([]int{0, 1, 2, 3, 4, 5, 6, 7, 8}, 10_000_000, 0); len(got) != MaxPermutations {
		t.Errorf("%d rows", len(got))
	}
}

func isAllInterval(r []int, n int) bool {
	if len(r) != n || r[0] != 0 {
		return false
	}
	seen := make([]bool, n)
	ivs := make([]bool, n)
	for i, p := range r {
		if p < 0 || p >= n || seen[p] {
			return false
		}
		seen[p] = true
		if i > 0 {
			iv := ((p-r[i-1])%n + n) % n
			if iv == 0 || ivs[iv] {
				return false
			}
			ivs[iv] = true
		}
	}
	return true
}

func TestAllIntervals(t *testing.T) {
	for _, n := range []int{2, 4, 6, 8, 12} {
		rows := AllIntervals(n, 5)
		if len(rows) == 0 {
			t.Errorf("n=%d: no rows", n)
		}
		for _, r := range rows {
			if !isAllInterval(r, n) {
				t.Errorf("n=%d: %v is no all interval row", n, r)
			}
		}
	}
	// the search is deterministic: the first row for twelve notes
	if got := AllIntervals(12, 1)[0]; !slices.Equal(got, []int{0, 1, 3, 2, 7, 10, 8, 4, 11, 5, 9, 6}) {
		t.Errorf("%v", got)
	}
	// n = 4 has two: 0 1 3 2 and 0 2 3 1 (found by the search)
	if got := len(AllIntervals(4, 0)); got != 2 {
		t.Errorf("n=4: %d rows", got)
	}
	// limit is respected
	if got := len(AllIntervals(12, 7)); got != 7 {
		t.Errorf("%d rows", got)
	}
}

func TestAllIntervalsEdgesAndSpeed(t *testing.T) {
	if AllIntervals(0, 5) != nil || AllIntervals(-3, 5) != nil {
		t.Error("n < 1")
	}
	if got := AllIntervals(1, 5); len(got) != 1 || got[0][0] != 0 {
		t.Errorf("%v", got)
	}
	// odd n has no such row
	if got := AllIntervals(5, 0); len(got) != 0 {
		t.Errorf("n=5: %v", got)
	}
	start := time.Now()
	AllIntervals(12, 50)
	if d := time.Since(start); d > time.Second {
		t.Errorf("n=12 limit 50 took %v", d)
	}
}

func TestReader(t *testing.T) {
	rows := [][]int{{0, 1}, {1, 0}, {2, 3}}
	r := NewReader(rows, ReadInOrder)
	if r.Continuous() {
		t.Error("in order ends")
	}
	for _, want := range rows {
		if r.Done() {
			t.Fatal("done too early")
		}
		got := r.Generate()
		if !slices.Equal(got, want) {
			t.Errorf("%v want %v", got, want)
		}
		got[0] = 99 // a copy
	}
	if !r.Done() || r.Generate() != nil || rows[0][0] != 0 {
		t.Error("end or copy")
	}
	r.Reset()
	if r.Done() || !slices.Equal(r.Generate(), rows[0]) {
		t.Error("reset")
	}
}

func TestReaderShuffledAndRandom(t *testing.T) {
	var rows [][]int
	for i := range 20 {
		rows = append(rows, []int{i})
	}
	genny.Seed(3)
	r := NewReader(rows, ReadShuffled)
	var got []int
	for !r.Done() {
		got = append(got, r.Generate()[0])
	}
	if slices.IsSorted(got) {
		t.Error("not shuffled")
	}
	slices.Sort(got)
	for i, v := range got {
		if v != i {
			t.Fatalf("every row once: %v", got)
		}
	}

	rnd := NewReader(rows, ReadRandom)
	if !rnd.Continuous() || rnd.Done() {
		t.Error("random is endless")
	}
	seen := map[int]bool{}
	for range 200 {
		seen[rnd.Generate()[0]] = true
	}
	if len(seen) < 10 {
		t.Errorf("only %d different rows", len(seen))
	}
}

func TestReaderEmpty(t *testing.T) {
	for _, o := range []Order{ReadInOrder, ReadShuffled, ReadRandom} {
		r := NewReader(nil, o)
		if r.Generate() != nil {
			t.Error("zero value")
		}
		if o != ReadRandom && !r.Done() {
			t.Error("done at once")
		}
		r.Reset()
	}
}
