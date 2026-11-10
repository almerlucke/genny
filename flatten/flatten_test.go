package flatten_test

import (
	"reflect"
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/constant"
	"github.com/almerlucke/genny/flatten"
	"github.com/almerlucke/genny/sequence"
)

func drain(f genny.Generator[int]) []int {
	var out []int
	for i := 0; i < 100 && !f.Done(); i++ {
		out = append(out, f.Generate())
	}
	return out
}

func TestFlatten(t *testing.T) {
	f := flatten.New[int](sequence.New([]int{1, 2}, []int{3}, []int{4, 5, 6}))
	if f.Continuous() || f.Done() {
		t.Fatal("fresh flatten")
	}
	if got := drain(f); !reflect.DeepEqual(got, []int{1, 2, 3, 4, 5, 6}) {
		t.Fatalf("got %v", got)
	}
	f.Reset()
	if f.Done() {
		t.Fatal("reset clears done")
	}
	if got := drain(f); !reflect.DeepEqual(got, []int{1, 2, 3, 4, 5, 6}) {
		t.Fatalf("after reset got %v", got)
	}
}

func TestFlattenSkipsEmptySlices(t *testing.T) {
	tests := []struct {
		name string
		in   [][]int
		want []int
	}{
		{"empty first", [][]int{{}, {1, 2}}, []int{1, 2}},
		{"empty middle", [][]int{{1}, {}, nil, {2}}, []int{1, 2}},
		{"empty last", [][]int{{1}, {}}, []int{1}},
		{"empty trailing run", [][]int{{1, 2}, nil, {}}, []int{1, 2}},
		{"all empty", [][]int{{}, nil}, nil},
		{"single empty", [][]int{{}}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := flatten.New[int](sequence.New(tt.in...))
			got := drain(f)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %v want %v", got, tt.want)
			}
			if !f.Done() {
				t.Fatal("must be done")
			}
			if f.Generate() != 0 {
				t.Fatal("a done flatten generates zero")
			}
			f.Reset()
			if got := drain(f); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("after reset got %v want %v", got, tt.want)
			}
		})
	}
}

func TestFlattenContinuousSource(t *testing.T) {
	f := flatten.New[int](constant.New([]int{7, 8}))
	if !f.Continuous() {
		t.Fatal("follows source")
	}
	var got []int
	for i := 0; i < 5; i++ {
		got = append(got, f.Generate())
	}
	if !reflect.DeepEqual(got, []int{7, 8, 7, 8, 7}) || f.Done() {
		t.Fatalf("got %v", got)
	}
}
