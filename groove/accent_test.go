package groove

import (
	"slices"
	"testing"
)

func TestAccent(t *testing.T) {
	a := NewAccent(2, 2, 3) // 7/8
	var got []float64
	for range 14 {
		got = append(got, a.Generate())
	}
	bar := []float64{1, 0.55, 0.8, 0.55, 0.8, 0.55, 0.55}
	if !slices.Equal(got, append(slices.Clone(bar), bar...)) {
		t.Errorf("%v", got)
	}
	a.Reset()
	if a.Generate() != 1 {
		t.Error("reset")
	}
	b := NewAccent(1).Levels(9, 8, 7) // one step per bar: always the downbeat
	if b.Generate() != 9 || b.Generate() != 9 {
		t.Error("levels")
	}
}

func TestAccentInvalid(t *testing.T) {
	for name, f := range map[string]func(){
		"none":    func() { NewAccent() },
		"group 0": func() { NewAccent(2, 0) },
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
