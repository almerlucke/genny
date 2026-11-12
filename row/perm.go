package row

import (
	"slices"

	"github.com/almerlucke/genny"
)

// MaxPermutations is the most results Permutations and AllIntervals give when no limit is asked, and
// the most they give at all
const MaxPermutations = 100000

// Permutations arranges groupLen of the elements (every element is used at most once, by position)
// in every possible order and returns the arrangements in a fixed order: lexicographic by position,
// so for the elements 1 2 3 the first is 1 2 3, the next 1 3 2, and so on. A groupLen outside
// 1..len(elems) means all elements. At most n arrangements are returned, n <= 0 means all of
// them, and the result is never longer than MaxPermutations. No elements gives nil
func Permutations(elems []int, n int, groupLen int) [][]int {
	if len(elems) == 0 {
		return nil
	}
	if groupLen < 1 || groupLen > len(elems) {
		groupLen = len(elems)
	}
	if n <= 0 || n > MaxPermutations {
		n = MaxPermutations
	}

	var out [][]int
	used := make([]bool, len(elems))
	cur := make([]int, 0, groupLen)
	var rec func()
	rec = func() {
		if len(cur) == groupLen {
			out = append(out, slices.Clone(cur))
			return
		}
		for i, e := range elems {
			if used[i] {
				continue
			}
			used[i] = true
			cur = append(cur, e)
			rec()
			cur = cur[:len(cur)-1]
			used[i] = false
			if len(out) >= n {
				return
			}
		}
	}
	rec()
	return out
}

// AllIntervals finds all interval rows of n pitch classes: permutations of 0..n-1 that start on 0
// and in which the n-1 steps between the neighbours (taken mod n, upwards) are all different,
// so every interval 1..n-1 occurs once. For n = 12 these are the twelve tone all-interval rows.
// The rows come in a fixed order from a backtracking search that stops after limit rows, limit <= 0
// means MaxPermutations. The search is fast for small limits also for n = 12. There is no such row
// for n < 1, the one note row 0 for n = 1
func AllIntervals(n int, limit int) [][]int {
	if n < 1 {
		return nil
	}
	if limit <= 0 || limit > MaxPermutations {
		limit = MaxPermutations
	}

	var out [][]int
	cur := make([]int, 1, n)
	used := make([]bool, n)
	usedIv := make([]bool, n)
	used[0] = true
	var rec func()
	rec = func() {
		if len(cur) == n {
			out = append(out, slices.Clone(cur))
			return
		}
		last := cur[len(cur)-1]
		for p := 1; p < n && len(out) < limit; p++ {
			iv := ((p-last)%n + n) % n
			if used[p] || usedIv[iv] {
				continue
			}
			used[p], usedIv[iv] = true, true
			cur = append(cur, p)
			rec()
			cur = cur[:len(cur)-1]
			used[p], usedIv[iv] = false, false
		}
	}
	rec()
	return out
}

// Order is the way a Reader goes through its rows
type Order int

const (
	// ReadInOrder reads the rows one after the other, then is done
	ReadInOrder Order = iota
	// ReadShuffled reads every row once in a random order, then is done
	ReadShuffled
	// ReadRandom draws a random row every time, forever
	ReadRandom
)

// Reader is a generator of rows (for instance from Permutations or AllIntervals), read in order or at
// random. The rows it gives are copies
type Reader struct {
	rows  [][]int
	order Order
	idx   []int
	pos   int
}

// NewReader creates a reader of the rows. Without rows it is done at once and generates nil
func NewReader(rows [][]int, order Order) *Reader {
	r := &Reader{rows: rows, order: order}
	r.Reset()
	return r
}

// Reset starts again from the first row; a ReadShuffled reader shuffles again
func (r *Reader) Reset() {
	r.pos = 0
	r.idx = make([]int, len(r.rows))
	for i := range r.idx {
		r.idx[i] = i
	}
	if r.order == ReadShuffled {
		genny.Shuffle(len(r.idx), func(i, j int) { r.idx[i], r.idx[j] = r.idx[j], r.idx[i] })
	}
}

// Generate returns the next row, or nil if there is none
func (r *Reader) Generate() []int {
	if len(r.rows) == 0 {
		return nil
	}
	if r.order == ReadRandom {
		return slices.Clone(r.rows[genny.Intn(len(r.rows))])
	}
	if r.pos >= len(r.idx) {
		return nil
	}
	row := slices.Clone(r.rows[r.idx[r.pos]])
	r.pos++
	return row
}

// Continuous is true for a ReadRandom reader, the others end
func (r *Reader) Continuous() bool { return r.order == ReadRandom }

// Done is true when all rows have been read (at once without rows)
func (r *Reader) Done() bool {
	return r.order != ReadRandom && r.pos >= len(r.idx)
}

// Compile-time check that *Reader implements genny.Generator: if a method is removed or its signature
// changes, the build fails here, next to the type, instead of where it is first used. It costs nothing
// at run time (a nil pointer assigned to the blank identifier).
var _ genny.Generator[[]int] = (*Reader)(nil)
