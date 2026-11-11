package tendency

import (
	"math"
	"slices"
	"testing"

	"github.com/almerlucke/genny"
)

func TestMaskAt(t *testing.T) {
	m := NewMask(false,
		Segment{N: 5, A1: 0, A2: 10, Z1: 4, Z2: 6},
		Segment{N: 0, A1: 99, A2: 99, Z1: 99, Z2: 99},
		Segment{N: 1, A1: 7, A2: 8},
		Segment{N: 2, A1: 3, A2: 1, Z1: 3, Z2: 1}, // swapped bounds
	)
	if m.Len() != 8 || m.Loops() {
		t.Fatalf("len %d loops %v", m.Len(), m.Loops())
	}
	tests := []struct {
		i      int
		lo, hi float64
	}{
		{-3, 0, 10}, // below 0 is 0
		{0, 0, 10},
		{2, 2, 8},
		{4, 4, 6},
		{5, 7, 8},
		{6, 1, 3},
		{7, 1, 3},
		{100, 1, 3}, // stays at the last band
	}
	for _, tc := range tests {
		if lo, hi := m.At(tc.i); lo != tc.lo || hi != tc.hi {
			t.Errorf("At(%d) = %v..%v, want %v..%v", tc.i, lo, hi, tc.lo, tc.hi)
		}
	}
	l := NewMask(true, Segment{N: 2, A1: 0, A2: 1, Z1: 2, Z2: 3})
	if lo, hi := l.At(3); lo != 2 || hi != 3 {
		t.Errorf("loop At(3) = %v..%v", lo, hi)
	}
	if lo, hi := l.At(4); lo != 0 || hi != 1 {
		t.Errorf("loop At(4) = %v..%v", lo, hi)
	}
	if lo, hi := NewMask(true).At(5); lo != 0 || hi != 0 {
		t.Error("empty mask")
	}
}

func TestValue(t *testing.T) {
	genny.Seed(1)
	m := NewMask(false,
		Segment{N: 50, A1: 0, A2: 0, Z1: 0, Z2: 100},
		Segment{N: 50, A1: 40, A2: 60, Z1: 40, Z2: 60},
	)
	v := NewValue(m)
	if v.Continuous() || v.Done() {
		t.Error("fresh non-looping mask")
	}
	var last float64
	for i := range 100 {
		x := v.Generate()
		lo, hi := m.At(i)
		if x < lo || x > hi {
			t.Fatalf("step %d: %v outside %v..%v", i, x, lo, hi)
		}
		last = x
	}
	if !v.Done() {
		t.Error("should be done after the last segment")
	}
	for range 5 {
		if v.Generate() != last {
			t.Error("done mask should keep the last value")
		}
	}
	v.Reset()
	if v.Done() || v.Generate() != 0 {
		t.Error("reset should restart at a band 0..0")
	}
}

func TestValueLoop(t *testing.T) {
	genny.Seed(2)
	m := NewMask(true, Segment{N: 3, A1: 10, A2: 10, Z1: 20, Z2: 20})
	v := NewValue(m)
	want := []float64{10, 15, 20, 10, 15, 20}
	for i, w := range want {
		if x := v.Generate(); x != w {
			t.Errorf("step %d: %v want %v", i, x, w)
		}
	}
	if !v.Continuous() || v.Done() {
		t.Error("looping mask never ends")
	}
}

func TestValueEmptyMask(t *testing.T) {
	for _, m := range []*Mask{NewMask(false), NewMask(true), NewMask(false, Segment{N: 0, A1: 1, A2: 2})} {
		v := NewValue(m)
		if !v.Done() || v.Generate() != 0 {
			t.Error("empty mask should be done at once with a zero value")
		}
	}
}

func TestBetaValue(t *testing.T) {
	genny.Seed(3)
	m := NewMask(true, Segment{N: 1, A1: 0, A2: 10, Z1: 0, Z2: 10})
	for _, c := range []struct{ a, b, mean float64 }{{1, 1, 5}, {1, 4, 2}, {4, 1, 8}} {
		v := NewBetaValue(m, c.a, c.b)
		var sum float64
		for range 20000 {
			x := v.Generate()
			if x < 0 || x > 10 {
				t.Fatalf("%v out of band", x)
			}
			sum += x
		}
		if mean := sum / 20000; math.Abs(mean-c.mean) > 0.2 {
			t.Errorf("beta(%v,%v) mean %v want %v", c.a, c.b, mean, c.mean)
		}
	}
}

func TestChoice(t *testing.T) {
	genny.Seed(4)
	items := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	// 0..50 percent is the first half, then 50..100 the second half, then everything out of range
	m := NewMask(false,
		Segment{N: 200, A1: 0, A2: 50, Z1: 0, Z2: 50},
		Segment{N: 200, A1: 50, A2: 100, Z1: 50, Z2: 100},
		Segment{N: 20, A1: -50, A2: 200, Z1: -50, Z2: 200},
	)
	c := NewChoice(m, items)
	seen := map[string]int{}
	for i := range 420 {
		x := c.Generate()
		idx := slices.Index(items, x)
		switch {
		case i < 200 && idx > 4, i >= 200 && i < 400 && idx < 5:
			t.Fatalf("step %d picked %s", i, x)
		}
		seen[x]++
	}
	if len(seen) != 10 {
		t.Errorf("not all elements picked: %v", seen)
	}
	if !c.Done() || c.Continuous() {
		t.Error("should be done")
	}
	if last := c.Generate(); last != c.Generate() {
		t.Error("keeps last pick")
	}

	// 100 percent picks the last element, not past it
	e := NewChoice(NewMask(true, Segment{N: 1, A1: 100, A2: 100}), items)
	if e.Generate() != "j" {
		t.Error("100 percent is the last element")
	}

	// empty list
	z := NewChoice(NewMask(true, Segment{N: 1, A1: 0, A2: 100}), []int(nil))
	if !z.Done() || z.Continuous() || z.Generate() != 0 {
		t.Error("empty list")
	}
	z.Reset()

	// copies the items
	src := []int{1}
	k := NewChoice(NewMask(true, Segment{N: 1, A1: 0, A2: 100}), src)
	src[0] = 9
	if k.Generate() != 1 {
		t.Error("items should be copied")
	}
}

func TestBetaChoice(t *testing.T) {
	genny.Seed(5)
	m := NewMask(true, Segment{N: 1, A1: 0, A2: 100})
	c := NewBetaChoice(m, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, 5, 1)
	var sum int
	for range 10000 {
		sum += c.Generate()
	}
	if mean := float64(sum) / 10000; mean < 6.5 {
		t.Errorf("a > b should favour the end, mean %v", mean)
	}
}

func TestTwoMasks(t *testing.T) {
	genny.Seed(6)
	outer := NewMask(false, Segment{N: 300, A1: 0, A2: 10, Z1: 0, Z2: 10})
	inner := NewMask(false, Segment{N: 300, A1: 4, A2: 6, Z1: 4, Z2: 6})
	g := NewTwoMasks(outer, inner, 100)
	for range 300 {
		if v := g.Generate(); v < 0 || v > 10 || (v >= 4 && v <= 6) {
			t.Fatalf("value %v inside the inner band", v)
		}
	}
	if !g.Done() || g.Continuous() {
		t.Error("done after the outer mask")
	}
	last := g.Generate()
	if g.Generate() != last {
		t.Error("keeps last value")
	}

	// the inner band covers the outer one: falls back to an outer value, never blocks
	g = NewTwoMasks(
		NewMask(true, Segment{N: 1, A1: 1, A2: 2}),
		NewMask(true, Segment{N: 1, A1: 0, A2: 3}), 5)
	for range 10 {
		if v := g.Generate(); v < 1 || v > 2 {
			t.Fatalf("fallback value %v", v)
		}
	}
	g.Reset()

	// no inner band, negative retries
	g = NewTwoMasks(NewMask(true, Segment{N: 1, A1: 1, A2: 2}), NewMask(false), -1)
	if v := g.Generate(); v < 1 || v > 2 {
		t.Errorf("value %v", v)
	}
	if !NewTwoMasks(NewMask(false), NewMask(false), 1).Done() {
		t.Error("empty outer mask is done")
	}
}

func TestAnchor(t *testing.T) {
	tests := []struct {
		name  string
		shape []float64
		n     int
		first float64
		last  float64
		lo    float64
		hi    float64
		want  []float64
	}{
		{"straight line", nil, 5, 0, 8, 0, 10, []float64{0, 2, 4, 6, 8}},
		{"flat shape", []float64{3, 3}, 3, 2, 4, 0, 10, []float64{2, 3, 4}},
		{"zero values", []float64{1}, 0, 0, 1, 0, 1, nil},
		{"one value", []float64{1, 2}, 1, 5, 9, 0, 10, []float64{5}},
	}
	for _, tc := range tests {
		got := Anchor(tc.shape, tc.n, tc.first, tc.last, tc.lo, tc.hi)
		if !slices.Equal(got, tc.want) {
			t.Errorf("%s: got %v want %v", tc.name, got, tc.want)
		}
	}

	// a shape up and down is kept and starts and ends at the anchors, within the limits
	got := Anchor([]float64{0, 1, 0}, 9, 2, 3, 0, 10)
	if len(got) != 9 || got[0] != 2 || got[8] != 3 || got[4] < 9 {
		t.Errorf("tent: %v", got)
	}
	for _, v := range Anchor([]float64{0, 5, 1, 9, 2}, 20, -5, 50, 10, 0) { // swapped limits and anchors out of range
		if v < 0 || v > 10 {
			t.Errorf("value %v outside the limits", v)
		}
	}
}

func TestSpray(t *testing.T) {
	genny.Seed(7)
	in := []float64{0, 10, 20}
	out := Spray(in, 1)
	if len(out) != 3 || in[1] != 10 {
		t.Fatal("new slice of the same length")
	}
	for i := range 200 {
		out = Spray(in, 1)
		for j, v := range out {
			if math.Abs(v-in[j]) > 1 {
				t.Fatalf("round %d: %v too far from %v", i, v, in[j])
			}
		}
	}
	if !slices.Equal(Spray(in, -1), in) || len(Spray(nil, 1)) != 0 {
		t.Error("negative spread and empty input")
	}
}

func TestInvalid(t *testing.T) {
	m := NewMask(true, Segment{N: 1})
	for name, f := range map[string]func(){
		"negative N":      func() { NewMask(false, Segment{N: -1}) },
		"nil mask value":  func() { NewValue(nil) },
		"nil mask choice": func() { NewChoice[int](nil, nil) },
		"beta value":      func() { NewBetaValue(m, 0, 1) },
		"beta choice":     func() { NewBetaChoice(m, []int{1}, 1, 0) },
		"nil outer":       func() { NewTwoMasks(nil, m, 1) },
		"nil inner":       func() { NewTwoMasks(m, nil, 1) },
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
