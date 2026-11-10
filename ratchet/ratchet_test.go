package ratchet

import (
	"slices"
	"testing"

	"github.com/almerlucke/genny"
)

func TestOffsets(t *testing.T) {
	if got := Offsets(1); !slices.Equal(got, []float64{0}) {
		t.Error(got)
	}
	if got := Offsets(4); !slices.Equal(got, []float64{0, 0.25, 0.5, 0.75}) {
		t.Error(got)
	}
}

func TestPattern(t *testing.T) {
	p := NewPattern(1, 2, 3)
	var got []int
	for range 7 {
		got = append(got, len(p.Generate()))
	}
	if !slices.Equal(got, []int{1, 2, 3, 1, 2, 3, 1}) {
		t.Error(got)
	}
	p.Reset()
	if len(p.Generate()) != 1 {
		t.Error("reset")
	}
}

func TestRandom(t *testing.T) {
	genny.Seed(1)
	never := NewRandom(0, 4)
	always := NewRandom(1, 4)
	counts := map[int]int{}
	for range 4000 {
		if len(never.Generate()) != 1 {
			t.Fatal("p=0 ratcheted")
		}
		counts[len(always.Generate())]++
	}
	if counts[1] != 0 || counts[2] == 0 || counts[3] == 0 || counts[4] == 0 || counts[5] != 0 {
		t.Errorf("counts %v", counts)
	}
}

func TestInvalid(t *testing.T) {
	for name, f := range map[string]func(){
		"offsets 0":   func() { Offsets(0) },
		"max count 1": func() { NewRandom(0.5, 1) },
		"empty":       func() { NewPattern() },
		"count 0":     func() { NewPattern(1, 0) },
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s should panic", name)
				}
			}()
			f()
		}()
	}
}
