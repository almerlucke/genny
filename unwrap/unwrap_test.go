package unwrap_test

import (
	"reflect"
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/constant"
	"github.com/almerlucke/genny/sequence"
	"github.com/almerlucke/genny/unwrap"
)

func drain(g genny.Generator[int]) []int {
	var out []int
	for i := 0; i < 100 && !g.Done(); i++ {
		out = append(out, g.Generate())
	}
	return out
}

func TestUnwrapPlaysGeneratorsInTurn(t *testing.T) {
	u := unwrap.New[int](sequence.New[genny.Generator[int]](sequence.New(1, 2), sequence.New(3), sequence.New(4, 5)))
	if u.Continuous() || u.Done() {
		t.Fatal("fresh unwrap")
	}
	if got := drain(u); !reflect.DeepEqual(got, []int{1, 2, 3, 4, 5}) {
		t.Fatalf("got %v", got)
	}
}

func TestUnwrapReset(t *testing.T) {
	u := unwrap.New[int](sequence.New[genny.Generator[int]](sequence.New(1, 2), sequence.New(3)))
	drain(u)
	u.Reset()
	if u.Done() {
		t.Fatal("reset clears done")
	}
	if got := drain(u); !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("after reset got %v", got)
	}
	// reset halfway through the first generator
	u.Reset()
	u.Generate()
	u.Reset()
	if got := drain(u); !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("after mid reset got %v", got)
	}
}

func TestUnwrapContinuousGeneratorsGiveOneValueEach(t *testing.T) {
	u := unwrap.New[int](sequence.New[genny.Generator[int]](constant.New(1), constant.New(2)))
	if got := drain(u); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Fatalf("got %v", got)
	}
}

func TestUnwrapContinuousSource(t *testing.T) {
	u := unwrap.New[int](sequence.NewLoop[genny.Generator[int]](sequence.New(1, 2), sequence.New(3)))
	if !u.Continuous() {
		t.Fatal("follows source")
	}
	var got []int
	for i := 0; i < 7; i++ {
		got = append(got, u.Generate())
	}
	if !reflect.DeepEqual(got, []int{1, 2, 3, 1, 2, 3, 1}) || u.Done() {
		t.Fatalf("got %v", got)
	}
}
