package arpeggio

import (
	"reflect"
	"sort"
	"testing"

	"github.com/almerlucke/genny"
)

func collect(a *Arpeggio[int], n int) []int {
	out := make([]int, n)
	for i := range out {
		out[i] = a.Generate()
	}
	return out
}

func TestPatterns(t *testing.T) {
	v3 := []int{10, 11, 12}
	v4 := []int{10, 11, 12, 13}
	tests := []struct {
		name    string
		values  []int
		mode    Mode
		mirror  MirrorMode
		reverse bool
		want    []int
	}{
		{"up", v4, Up, None, false, []int{10, 11, 12, 13}},
		{"up reverse", v4, Up, None, true, []int{13, 12, 11, 10}},
		{"up exclusive", v4, Up, Exclusive, false, []int{10, 11, 12, 13, 12, 11}},
		{"up inclusive", v3, Up, Inclusive, false, []int{10, 11, 12, 12, 11, 10}},
		{"up inclusive reverse", v3, Up, Inclusive, true, []int{10, 11, 12, 12, 11, 10}},
		{"up exclusive reverse", v3, Up, Exclusive, true, []int{11, 12, 11, 10}},
		{"converge even", v4, Converge, None, false, []int{10, 13, 11, 12}},
		{"converge odd", []int{10, 11, 12, 13, 14}, Converge, None, false, []int{10, 14, 11, 13, 12}},
		{"converge one", []int{10}, Converge, None, false, []int{10}},
		{"converge two", []int{10, 11}, Converge, None, false, []int{10, 11}},
		{"alternate", v3, Alternate, None, false, []int{10, 11, 10, 12}},
		{"alternate two", []int{10, 11}, Alternate, None, false, []int{10, 11}},
		{"up one exclusive", []int{10}, Up, Exclusive, false, []int{10}},
		{"up one inclusive", []int{10}, Up, Inclusive, false, []int{10, 10}},
		{"up two exclusive", []int{10, 11}, Up, Exclusive, false, []int{10, 11}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := New(tt.values, tt.mode, tt.mirror, tt.reverse)
			got := collect(a, len(tt.want))
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %v want %v", got, tt.want)
			}
			if !a.Done() {
				t.Fatal("must be done after one pass")
			}
		})
	}
}

func TestDoneAndLoop(t *testing.T) {
	a := New([]int{1, 2, 3}, Up, None, false)
	if a.Continuous() || a.Done() {
		t.Fatal("fresh arpeggio")
	}
	collect(a, 2)
	if a.Done() {
		t.Fatal("not done before the last value")
	}
	a.Generate()
	if !a.Done() {
		t.Fatal("done after the last value")
	}
	if got := collect(a, 2); !reflect.DeepEqual(got, []int{3, 3}) {
		t.Fatalf("done arpeggio keeps the last value, got %v", got)
	}
	a.Reset()
	if a.Done() {
		t.Fatal("reset clears done")
	}
	if got := collect(a, 3); !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("got %v", got)
	}

	l := NewLoop([]int{1, 2}, Up, None, false)
	if !l.Continuous() {
		t.Fatal("loop is continuous")
	}
	if got := collect(l, 5); !reflect.DeepEqual(got, []int{1, 2, 1, 2, 1}) || l.Done() {
		t.Fatalf("got %v done %v", got, l.Done())
	}
}

func TestRandomIsAPermutationAndRepeatable(t *testing.T) {
	values := []int{1, 2, 3, 4, 5, 6}
	genny.Seed(42)
	first := collect(New(values, Random, None, false), 6)
	genny.Seed(42)
	second := collect(New(values, Random, None, false), 6)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("same seed gave %v and %v", first, second)
	}
	sorted := append([]int(nil), first...)
	sort.Ints(sorted)
	if !reflect.DeepEqual(sorted, values) {
		t.Fatalf("not a permutation: %v", first)
	}
	// the values slice itself is not shuffled
	if !reflect.DeepEqual(values, []int{1, 2, 3, 4, 5, 6}) {
		t.Fatalf("values were modified: %v", values)
	}
}

func TestEmptyPattern(t *testing.T) {
	cases := map[string]*Arpeggio[int]{
		"no values":        New[int](nil, Up, None, false),
		"no values loop":   NewLoop[int](nil, Converge, Inclusive, true),
		"one alternate":    New([]int{5}, Alternate, None, false),
		"no values random": New[int](nil, Random, Exclusive, false),
	}
	for name, a := range cases {
		t.Run(name, func(t *testing.T) {
			if !a.Done() || a.Generate() != 0 {
				t.Fatal("must be done and generate zero")
			}
			a.Reset()
			if !a.Done() || a.Generate() != 0 {
				t.Fatal("must stay done after Reset")
			}
		})
	}
}
